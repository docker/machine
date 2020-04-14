package azureutil

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-12-01/network"
	"github.com/rancher/machine/drivers/azure/logutil"
	"github.com/rancher/machine/libmachine/log"
)

type cleanupResource interface {
	// Get retrieves if the resource and saves its reference to the instance
	// for further using, returned error is used to determine if the resource
	// exists
	Get(ctx context.Context, a AzureClient) error

	// Delete deletes the resource
	Delete(ctx context.Context, a AzureClient) error

	// CanBeDeleted checks the resource reference if it has dependent
	// resources attached to it preventing it from being deleted or if
	// it has tags that indicate that it should not be deleted
	CanBeDeleted(ctx context.Context, a AzureClient) bool

	// ResourceType returns human-readable name of the type of the resource.
	ResourceType() string

	// LogFields returns the logging fields used during cleanup logging.
	LogFields() logutil.Fields
}

// cleanupResourceIfExists checks if the resource exists, if it does and it
// does not have any attached resources, then deletes the resource. If the
// resource does not exist or is not eligible for cleanup, returns nil. If an
// error is encountered, returns the error.
func (a AzureClient) cleanupResourceIfExists(ctx context.Context, r cleanupResource) error {
	f := r.LogFields()
	log.Info(fmt.Sprintf("Attempting to clean up %s resource...", r.ResourceType()), f)
	err := r.Get(ctx, a)
	exists, err := checkResourceExistsFromError(err)
	if err != nil {
		return err
	}
	if !exists {
		log.Debug(fmt.Sprintf("%s resource does not exist. Skipping.", r.ResourceType()), f)
		return nil
	}

	if r.CanBeDeleted(ctx, a) {
		log.Debug(fmt.Sprintf("%s does not have any attached dependent resources or tags to skip deletion.", r.ResourceType()), f)
		log.Info(fmt.Sprintf("Removing %s resource...", r.ResourceType()), f)
		return r.Delete(ctx, a)
	}
	log.Info(fmt.Sprintf("%s is still in use by other resources or is tagged to skip deletion, skipping removal.", r.ResourceType()), f)
	return nil
}

// subnetCleanup manages cleanup of Subnet resources
type subnetCleanup struct {
	rg, vnet, name string
	ref            network.Subnet
}

func (c *subnetCleanup) Get(ctx context.Context, a AzureClient) (err error) {
	serviceClient := a.subnetsClient()
	c.ref, err = serviceClient.Get(ctx, c.rg, c.vnet, c.name, "")
	return err
}

func (c *subnetCleanup) Delete(ctx context.Context, a AzureClient) error {
	serviceClient := a.subnetsClient()
	future, err := serviceClient.Delete(ctx, c.rg, c.vnet, c.name)
	if err != nil {
		return err
	}
	err = future.WaitForCompletionRef(ctx, serviceClient.Client)
	if err != nil {
		return err
	}
	_, err = future.Result(serviceClient)
	return err
}

func (c *subnetCleanup) ResourceType() string { return "Subnet" }

func (c *subnetCleanup) LogFields() logutil.Fields { return logutil.Fields{"name": c.name} }

func (c *subnetCleanup) CanBeDeleted(ctx context.Context, a AzureClient) bool {
	c.Get(ctx, a) // updates c.ref
	return c.ref.SubnetPropertiesFormat.IPConfigurations == nil || len(*c.ref.SubnetPropertiesFormat.IPConfigurations) == 0
}

// vnetCleanup manages cleanup of Virtual Network resources.
type vnetCleanup struct {
	rg, name string
	ref      network.VirtualNetwork
}

func (c *vnetCleanup) Get(ctx context.Context, a AzureClient) (err error) {
	serviceClient := a.virtualNetworksClient()
	c.ref, err = serviceClient.Get(ctx, c.rg, c.name, "")
	return err
}

func (c *vnetCleanup) Delete(ctx context.Context, a AzureClient) error {
	serviceClient := a.virtualNetworksClient()
	future, err := serviceClient.Delete(ctx, c.rg, c.name)
	if err != nil {
		return err
	}
	err = future.WaitForCompletionRef(ctx, serviceClient.Client)
	if err != nil {
		return err
	}
	_, err = future.Result(serviceClient)
	return err
}

func (c *vnetCleanup) ResourceType() string { return "Virtual Network" }

func (c *vnetCleanup) LogFields() logutil.Fields { return logutil.Fields{"name": c.name} }

func (c *vnetCleanup) CanBeDeleted(ctx context.Context, a AzureClient) bool {
	c.Get(ctx, a) // updates c.ref
	return c.ref.VirtualNetworkPropertiesFormat.Subnets == nil || len(*c.ref.VirtualNetworkPropertiesFormat.Subnets) == 0
}

// avSetCleanup manages cleanup of Availability Set resources.
type avSetCleanup struct {
	rg, name string
	ref      compute.AvailabilitySet
}

func (c *avSetCleanup) Get(ctx context.Context, a AzureClient) (err error) {
	serviceClient := a.availabilitySetsClient()
	c.ref, err = serviceClient.Get(ctx, c.rg, c.name)
	return err
}

func (c *avSetCleanup) Delete(ctx context.Context, a AzureClient) error {
	serviceClient := a.availabilitySetsClient()
	_, err := serviceClient.Delete(ctx, c.rg, c.name)
	return err
}

func (c *avSetCleanup) ResourceType() string { return "Availability Set" }

func (c *avSetCleanup) LogFields() logutil.Fields { return logutil.Fields{"name": c.name} }

func (c *avSetCleanup) CanBeDeleted(ctx context.Context, a AzureClient) bool {
	c.Get(ctx, a) // updates c.ref
	return c.ref.AvailabilitySetProperties.VirtualMachines == nil && len(*c.ref.AvailabilitySetProperties.VirtualMachines) == 0
}

type nsgCleanup struct {
	rg, name   string
	usedInPool bool
	ref        network.SecurityGroup
}

func (c *nsgCleanup) Get(ctx context.Context, a AzureClient) (err error) {
	serviceClient := a.securityGroupsClient()
	c.ref, err = serviceClient.Get(ctx, c.rg, c.name, "")
	return err
}

func (c *nsgCleanup) Delete(ctx context.Context, a AzureClient) error {
	serviceClient := a.securityGroupsClient()
	future, err := serviceClient.Delete(ctx, c.rg, c.name)
	if err != nil {
		return err
	}
	err = future.WaitForCompletionRef(ctx, serviceClient.Client)
	if err != nil {
		return err
	}
	_, err = future.Result(serviceClient)
	return err
}

func (c *nsgCleanup) ResourceType() string { return "Network Security Group" }

func (c *nsgCleanup) LogFields() logutil.Fields { return logutil.Fields{"name": c.name} }

func (c *nsgCleanup) CanBeDeleted(ctx context.Context, a AzureClient) bool {
	c.Get(ctx, a) // updates c.ref
	sgProp := c.ref.SecurityGroupPropertiesFormat
	orphaned := (sgProp.Subnets == nil || len(*sgProp.Subnets) == 0) &&
		(sgProp.NetworkInterfaces == nil || len(*sgProp.NetworkInterfaces) == 0)
	return orphaned
}

type publicIPCleanup struct {
	rg, name string
}

func (c *publicIPCleanup) Get(ctx context.Context, a AzureClient) (err error) {
	serviceClient := a.publicIPAddressClient()
	_, err = serviceClient.Get(ctx, c.rg, c.name, "")
	return err
}

func (c *publicIPCleanup) Delete(ctx context.Context, a AzureClient) error {
	serviceClient := a.publicIPAddressClient()
	future, err := serviceClient.Delete(ctx, c.rg, c.name)
	if err != nil {
		return err
	}
	err = future.WaitForCompletionRef(ctx, serviceClient.Client)
	if err != nil {
		return err
	}
	_, err = future.Result(serviceClient)
	return err
}

func (c *publicIPCleanup) ResourceType() string { return "Public IP Address" }

func (c *publicIPCleanup) LogFields() logutil.Fields { return logutil.Fields{"name": c.name} }

func (c *publicIPCleanup) CanBeDeleted(ctx context.Context, a AzureClient) bool { return true }

type networkInterfaceCleanup struct {
	rg, name string
}

func (c *networkInterfaceCleanup) Get(ctx context.Context, a AzureClient) (err error) {
	serviceClient := a.networkInterfacesClient()
	_, err = serviceClient.Get(ctx, c.rg, c.name, "")
	return err
}

func (c *networkInterfaceCleanup) Delete(ctx context.Context, a AzureClient) error {
	serviceClient := a.networkInterfacesClient()
	future, err := serviceClient.Delete(ctx, c.rg, c.name)
	if err != nil {
		return err
	}
	err = future.WaitForCompletionRef(ctx, serviceClient.Client)
	if err != nil {
		return err
	}
	_, err = future.Result(serviceClient)
	return err
}

func (c *networkInterfaceCleanup) ResourceType() string { return "Network Interface" }

func (c *networkInterfaceCleanup) LogFields() logutil.Fields { return logutil.Fields{"name": c.name} }

func (c *networkInterfaceCleanup) CanBeDeleted(ctx context.Context, a AzureClient) bool { return true }

type vmCleanup struct {
	rg, name string
	ref      compute.VirtualMachine
}

func (c *vmCleanup) Get(ctx context.Context, a AzureClient) (err error) {
	serviceClient := a.virtualMachinesClient()
	c.ref, err = serviceClient.Get(ctx, c.rg, c.name, "")
	return err
}

func (c *vmCleanup) Delete(ctx context.Context, a AzureClient) error {
	serviceClient := a.virtualMachinesClient()
	future, err := serviceClient.Delete(ctx, c.rg, c.name)
	if err != nil {
		return err
	}
	err = future.WaitForCompletionRef(ctx, serviceClient.Client)
	if err != nil {
		return err
	}
	_, err = future.Result(serviceClient)
	return err
}

func (c *vmCleanup) ResourceType() string { return "Virtual Machine" }

func (c *vmCleanup) LogFields() logutil.Fields { return logutil.Fields{"name": c.name} }

func (c *vmCleanup) CanBeDeleted(ctx context.Context, a AzureClient) bool { return true }
