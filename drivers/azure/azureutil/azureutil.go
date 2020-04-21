package azureutil

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rancher/machine/drivers/azure/logutil"
	"github.com/rancher/machine/libmachine/log"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-12-01/network"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

const (
	storageAccountPrefix = "vhds" // do not contaminate to user's existing storage accounts

	powerStatePollingInterval = time.Second * 5
	waitStartTimeout          = time.Minute * 10
	waitPowerOffTimeout       = time.Minute * 5
)

var (
	// Private IPv4 address space per RFC 1918.
	defaultVnetAddressPrefixes = []string{
		"192.168.0.0/16",
		"10.0.0.0/8",
		"172.16.0.0/12"}
)

// AzureClient contains the information necessary to instantiate Azure service clients
type AzureClient struct {
	env            azure.Environment
	subscriptionID string
	auth           autorest.Authorizer
}

// New creates a new Azure client
func New(env azure.Environment, subsID string, auth autorest.Authorizer) *AzureClient {
	return &AzureClient{env, subsID, auth}
}

// RegisterResourceProviders registers current subscription to the specified
// resource provider namespaces if they are not already registered. Namespaces
// are case-insensitive.
func (a AzureClient) RegisterResourceProviders(ctx context.Context, namespaces ...string) error {
	providersClient := a.providersClient()
	for _, ns := range namespaces {
		p, err := providersClient.Get(ctx, ns, "")
		if err != nil {
			if strings.Contains(err.Error(), "Status=403") {
				return fmt.Errorf("Service principal does not have adequate role-based access privileges: %s", err)
			}
			return err
		}
		if to.String(p.RegistrationState) == "Registered" {
			log.Debugf("Already registered for %q", ns)
		} else {
			log.Info("Registering subscription to resource provider.", logutil.Fields{
				"ns":   ns,
				"subs": a.subscriptionID,
			})
			if _, err := providersClient.Register(ctx, ns); err != nil {
				return err
			}
		}
	}
	return nil
}

// CreateResourceGroup creates a Resource Group if not exists
func (a AzureClient) CreateResourceGroup(ctx context.Context, name, location string) error {
	ok, err := a.resourceGroupExists(ctx, name)
	if err != nil {
		return err
	}
	if ok {
		log.Infof("Resource group %q already exists.", name)
		return nil
	}

	log.Info("Creating resource group.", logutil.Fields{
		"name":     name,
		"location": location})
	_, err = a.resourceGroupsClient().CreateOrUpdate(ctx, name,
		resources.Group{
			Location: to.StringPtr(location),
		})
	return err
}

func (a AzureClient) resourceGroupExists(ctx context.Context, name string) (bool, error) {
	log.Info("Querying existing resource group.", logutil.Fields{"name": name})
	_, err := a.resourceGroupsClient().Get(ctx, name)
	return checkResourceExistsFromError(err)
}

// CreateNetworkSecurityGroup either creates or updates the definition of the requested security group with the specified rules and adds it to our DeploymentContext
func (a AzureClient) CreateNetworkSecurityGroup(ctx context.Context, deploymentCtx *DeploymentContext, resourceGroup, name, location string, rules *[]network.SecurityRule) error {
	log.Info("Configuring network security group.", logutil.Fields{
		"name":     name,
		"location": location})
	securityGroupsClient := a.securityGroupsClient()
	future, err := securityGroupsClient.CreateOrUpdate(ctx, resourceGroup, name,
		network.SecurityGroup{
			Location: to.StringPtr(location),
			SecurityGroupPropertiesFormat: &network.SecurityGroupPropertiesFormat{
				SecurityRules: rules,
			},
		})
	if err != nil {
		return err
	}
	if err = future.WaitForCompletionRef(ctx, securityGroupsClient.Client); err != nil {
		return err
	}
	nsg, err := future.Result(securityGroupsClient)
	deploymentCtx.NetworkSecurityGroupID = to.String(nsg.ID)
	return err
}

// DeleteNetworkSecurityGroupIfExists checks to see if the security group exists and accordingly deletes it
func (a AzureClient) DeleteNetworkSecurityGroupIfExists(ctx context.Context, resourceGroup, name string) error {
	return a.cleanupResourceIfExists(ctx, &nsgCleanup{rg: resourceGroup, name: name})
}

// CreatePublicIPAddress creates a public IP address and adds it to our DeploymentContext
func (a AzureClient) CreatePublicIPAddress(ctx context.Context, deploymentCtx *DeploymentContext, resourceGroup, name, location string, isStatic bool, dnsLabel string) error {
	log.Info("Creating public IP address.", logutil.Fields{
		"name":   name,
		"static": isStatic})

	var ipType network.IPAllocationMethod
	if isStatic {
		ipType = network.Static
	} else {
		ipType = network.Dynamic
	}

	var dns *network.PublicIPAddressDNSSettings
	if dnsLabel != "" {
		dns = &network.PublicIPAddressDNSSettings{
			DomainNameLabel: to.StringPtr(dnsLabel),
		}
	}

	publicIPAddressClient := a.publicIPAddressClient()
	future, err := publicIPAddressClient.CreateOrUpdate(ctx, resourceGroup, name,
		network.PublicIPAddress{
			Location: to.StringPtr(location),
			PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: ipType,
				DNSSettings:              dns,
			},
		})
	if err != nil {
		return err
	}
	err = future.WaitForCompletionRef(ctx, publicIPAddressClient.Client)
	if err != nil {
		return err
	}
	ip, err := future.Result(publicIPAddressClient)

	deploymentCtx.PublicIPAddressID = to.String(ip.ID)
	return err
}

// DeletePublicIPAddressIfExists checks to see if the IP Address exists and accordingly deletes it
func (a AzureClient) DeletePublicIPAddressIfExists(ctx context.Context, resourceGroup, name string) error {
	return a.cleanupResourceIfExists(ctx, &publicIPCleanup{rg: resourceGroup, name: name})
}

// CreateVirtualNetworkIfNotExists checks to see if a virtual network exists with the name provided and either updates or creates it accordingly
func (a AzureClient) CreateVirtualNetworkIfNotExists(ctx context.Context, resourceGroup, name, location string) error {
	f := logutil.Fields{
		"name":     name,
		"rg":       resourceGroup,
		"location": location,
	}
	log.Info("Querying if virtual network already exists.", f)

	exists, err := a.virtualNetworkExists(ctx, resourceGroup, name, location)
	if err != nil {
		return err
	}
	if exists {
		log.Info("Virtual network already exists.", f)
		return nil
	}

	log.Info("Creating virtual network.", f)
	virtualNetworksClient := a.virtualNetworksClient()
	future, err := virtualNetworksClient.CreateOrUpdate(ctx, resourceGroup, name,
		network.VirtualNetwork{
			Location: to.StringPtr(location),
			VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
				AddressSpace: &network.AddressSpace{
					AddressPrefixes: to.StringSlicePtr(defaultVnetAddressPrefixes),
				},
			},
		})
	if err != nil {
		return err
	}
	if err = future.WaitForCompletionRef(ctx, virtualNetworksClient.Client); err != nil {
		return err
	}
	_, err = future.Result(virtualNetworksClient)
	return err
}

func (a AzureClient) virtualNetworkExists(ctx context.Context, resourceGroup, name, location string) (bool, error) {
	vnet, err := a.virtualNetworksClient().Get(ctx, resourceGroup, name, "")
	if err == nil && to.String(vnet.Location) != location {
		return false, fmt.Errorf(
			"Virtual Network must exist in the same region, provided name %s exists in region %s, expected %s",
			name, to.String(vnet.Location), location)
	}
	return checkResourceExistsFromError(err)
}

// CleanupVirtualNetworkIfExists removes a subnet if there are no subnets
// attached to it. Note that this method is not safe for multiple concurrent
// writers, in case of races, deployment of a machine could fail or resource
// might not be cleaned up.
func (a AzureClient) CleanupVirtualNetworkIfExists(ctx context.Context, resourceGroup, name string) error {
	return a.cleanupResourceIfExists(ctx, &vnetCleanup{rg: resourceGroup, name: name})
}

// CreateSubnet creates or updates a subnet if it does not already exist.
func (a AzureClient) CreateSubnet(ctx context.Context, deploymentCtx *DeploymentContext, resourceGroup, virtualNetwork, name, subnetPrefix string) error {
	subnet, err := a.subnetsClient().Get(ctx, resourceGroup, virtualNetwork, name, "")
	exists, err := checkResourceExistsFromError(err)
	if err != nil {
		log.Warn("Unexpected get subnet operation error %v: ", err)
		return err
	}
	if exists {
		log.Info("Subnet already exists.")
		deploymentCtx.SubnetID = to.String(subnet.ID)
		return err
	}
	// If the subnet is not found, create it
	log.Info("Configuring subnet.", logutil.Fields{
		"name": name,
		"vnet": virtualNetwork,
		"cidr": subnetPrefix})
	subnetsClient := a.subnetsClient()
	future, err := subnetsClient.CreateOrUpdate(ctx, resourceGroup, virtualNetwork, name,
		network.Subnet{
			SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
				AddressPrefix: to.StringPtr(subnetPrefix),
			},
		})
	if err != nil {
		return err
	}
	if err = future.WaitForCompletionRef(ctx, subnetsClient.Client); err != nil {
		return err
	}
	subnet, err = future.Result(subnetsClient)
	deploymentCtx.SubnetID = to.String(subnet.ID)
	return err

}

// CleanupSubnetIfExists removes a subnet if there are no IP configurations
// (through NICs) are attached to it. Note that this method is not safe for
// multiple concurrent writers, in case of races, deployment of a machine could
// fail or resource might not be cleaned up.
func (a AzureClient) CleanupSubnetIfExists(ctx context.Context, resourceGroup, virtualNetwork, name string) error {
	return a.cleanupResourceIfExists(ctx, &subnetCleanup{
		rg: resourceGroup, vnet: virtualNetwork, name: name,
	})
}

// CreateNetworkInterface creates a network interface
func (a AzureClient) CreateNetworkInterface(ctx context.Context, deploymentCtx *DeploymentContext, resourceGroup, name, location, publicIPAddressID, subnetID, nsgID, privateIPAddress string) error {
	// NOTE(ahmetalpbalkan) This method is expected to fail if the user
	// specified Azure location is different than location of the virtual
	// network as Azure does not support cross-region virtual networks. In this
	// situation, user will get an explanatory API error from Azure.
	log.Info("Creating network interface.", logutil.Fields{"name": name})

	var publicIP *network.PublicIPAddress
	if publicIPAddressID != "" {
		publicIP = &network.PublicIPAddress{ID: to.StringPtr(publicIPAddressID)}
	}

	var privateIPAllocMethod = network.Dynamic
	if privateIPAddress != "" {
		privateIPAllocMethod = network.Static
	}
	networkInterfacesClient := a.networkInterfacesClient()
	future, err := networkInterfacesClient.CreateOrUpdate(ctx, resourceGroup, name, network.Interface{
		Location: to.StringPtr(location),
		InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
			NetworkSecurityGroup: &network.SecurityGroup{
				ID: to.StringPtr(nsgID),
			},
			IPConfigurations: &[]network.InterfaceIPConfiguration{
				{
					Name: to.StringPtr("ip"),
					InterfaceIPConfigurationPropertiesFormat: &network.InterfaceIPConfigurationPropertiesFormat{
						PrivateIPAddress:          to.StringPtr(privateIPAddress),
						PrivateIPAllocationMethod: privateIPAllocMethod,
						PublicIPAddress:           publicIP,
						Subnet: &network.Subnet{
							ID: to.StringPtr(subnetID),
						},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}
	if err = future.WaitForCompletionRef(ctx, networkInterfacesClient.Client); err != nil {
		return err
	}
	nic, err := future.Result(networkInterfacesClient)
	deploymentCtx.NetworkInterfaceID = to.String(nic.ID)
	return err
}

// DeleteNetworkInterfaceIfExists deletes a network interface if it exists
func (a AzureClient) DeleteNetworkInterfaceIfExists(ctx context.Context, resourceGroup, name string) error {
	return a.cleanupResourceIfExists(ctx, &networkInterfaceCleanup{rg: resourceGroup, name: name})
}

// CreateStorageAccount sees if the storage account provided exists or otherwise creates a storage account for you and stores the data into DeploymentContext
func (a AzureClient) CreateStorageAccount(ctx context.Context, deploymentCtx *DeploymentContext, resourceGroup, location string, storageType storage.SkuName) error {
	s, err := a.findOrCreateStorageAccount(ctx, resourceGroup, location, storageType)
	deploymentCtx.StorageAccount = s
	return err
}

func (a AzureClient) findOrCreateStorageAccount(ctx context.Context, resourceGroup, location string, storageType storage.SkuName) (*storage.AccountProperties, error) {
	s, err := a.findStorageAccount(ctx, resourceGroup, location, storageAccountPrefix, storageType)
	if err != nil {
		return nil, err
	}
	if s != nil {
		return s, nil
	}

	log.Debug("No eligible storage account found.", logutil.Fields{
		"location": location,
		"sku":      storageType})
	return a.createStorageAccount(ctx, resourceGroup, location, storageType)
}

func (a AzureClient) findStorageAccount(ctx context.Context, resourceGroup, location, prefix string, storageType storage.SkuName) (*storage.AccountProperties, error) {
	f := logutil.Fields{
		"sku":      storageType,
		"prefix":   prefix,
		"location": location}
	log.Debug("Querying existing storage accounts.", f)
	l, err := a.storageAccountsClient().ListByResourceGroup(ctx, resourceGroup)
	if err != nil {
		return nil, err
	}

	if !l.IsEmpty() {
		for _, v := range *l.Value {
			log.Debug("Iterating...", logutil.Fields{
				"name":     to.String(v.Name),
				"sku":      storageType,
				"location": to.String(v.Location),
			})
			if to.String(v.Location) == location && v.Sku.Name == storageType && strings.HasPrefix(to.String(v.Name), prefix) {
				log.Debug("Found eligible storage account.", logutil.Fields{"name": to.String(v.Name)})
				log.Info("Using existing storage account.", logutil.Fields{
					"name": to.String(v.Name),
					"sku":  storageType,
				})
				return v.AccountProperties, nil
			}
		}
	}
	log.Debug("No account matching the pattern is found.", f)
	return nil, err
}

func (a AzureClient) createStorageAccount(ctx context.Context, resourceGroup, location string, storageType storage.SkuName) (*storage.AccountProperties, error) {
	name := randomAzureStorageAccountName() // if it's not random enough, then you're unlucky

	f := logutil.Fields{
		"name":     name,
		"location": location,
		"sku":      storageType,
	}

	log.Info("Creating storage account.", f)
	storageAccountsClient := a.storageAccountsClient()
	future, err := storageAccountsClient.Create(ctx, resourceGroup, name,
		storage.AccountCreateParameters{
			Location: to.StringPtr(location),
			Sku:      &storage.Sku{Name: storageType},
		})
	if err != nil {
		return nil, err
	}
	if err = future.WaitForCompletionRef(ctx, storageAccountsClient.Client); err != nil {
		return nil, err
	}
	account, err := future.Result(storageAccountsClient)
	return account.AccountProperties, nil
}

// VirtualMachineExists sees if a virtual machine exists
func (a AzureClient) VirtualMachineExists(ctx context.Context, resourceGroup, name string) (bool, error) {
	_, err := a.virtualMachinesClient().Get(ctx, resourceGroup, name, "")
	return checkResourceExistsFromError(err)
}

// DeleteVirtualMachineIfExists checks to see if a VM exists and deletes it accordingly
// It then
func (a AzureClient) DeleteVirtualMachineIfExists(ctx context.Context, resourceGroup, name string) error {
	vmCleanupInfo := vmCleanup{rg: resourceGroup, name: name}
	err := vmCleanupInfo.Get(ctx, a)
	vm := vmCleanupInfo.ref // value is set if err != nil

	if exists, err := checkResourceExistsFromError(err); err != nil || !exists {
		// Either returns an error or nil if !exists
		return err
	}
	if err = a.cleanupResourceIfExists(ctx, &vmCleanupInfo); err != nil {
		return err
	}

	// Remove disk
	if vmProperties := vm.VirtualMachineProperties; vmProperties != nil {
		// TODO: remove unattached managed disk, requires azure sdk upgrade
		if managedDisk := vmProperties.StorageProfile.OsDisk.ManagedDisk; managedDisk != nil {
			diskName := ResourceNaming(name).OSDisk()
			log.Infof("Disk [%s] in resource group [%s] must be removed manually.", diskName, resourceGroup)
		}
		// if vhd is not nil then disk is unmanaged and disk blob should be removed
		if vhd := vmProperties.StorageProfile.OsDisk.Vhd; vhd != nil {
			return a.removeOSDiskBlob(ctx, resourceGroup, name, to.String(vhd.URI))
		}
	}
	return nil
}

func (a AzureClient) removeOSDiskBlob(ctx context.Context, resourceGroup, vmName, vhdURL string) error {
	// NOTE(ahmetalpbalkan) Currently Azure APIs do not offer a Delete Virtual
	// Machine functionality which deletes the attached disks along with the VM
	// as well. Therefore we find out the storage account from OS disk URL and
	// fetch storage account keys to delete the container containing the disk.
	log.Debug("Attempting to remove OS disk.", logutil.Fields{"vm": vmName})
	log.Debugf("OS Disk vhd URL: %q", vhdURL)

	blobContainersClient := a.blobContainersClient()
	vhdContainer := ResourceNaming(vmName).OSDiskContainer()
	storageAccount := extractStorageAccountFromVHDURL(vhdURL)
	if storageAccount == "" {
		log.Warn("Could not extract the storage account name from URL. Please clean up the disk yourself.")
		return nil
	}
	f := logutil.Fields{
		"account":   storageAccount,
		"container": vhdContainer}

	log.Debug("Removing container of disk blobs.", f)
	_, err := blobContainersClient.Get(ctx, resourceGroup, storageAccount, vhdContainer)
	exists, err := checkResourceExistsFromError(err)
	if err != nil {
		log.Warnf("Encountered error while checking if VHD container exists: %s. Please clean up the disk yourself.", err)
		return nil
	}
	if exists {
		if _, err := blobContainersClient.Delete(ctx, resourceGroup, storageAccount, vhdContainer); err != nil {
			log.Warnf("Encountered error while trying to delete VHD container: %s. Please clean up the disk yourself.", err)
			return nil
		}
	} else {
		log.Debugf("Did not find container in storage account")
	}

	lcrp, err := blobContainersClient.List(ctx, resourceGroup, storageAccount, "", "")
	if err != nil {
		return err
	}
	if len(lcrp.Values()) == 0 {
		log.Debugf("No storage containers left. Deleting virtual machine storage account.")
		resp, err := a.storageAccountsClient().Delete(ctx, resourceGroup, storageAccount)
		if err != nil {
			return err
		}
		log.Debugf("Storage account deletion happened: %v", resp.Response.Status)
	}
	return nil
}

// CreateVirtualMachine creates a VM according to the specifications and adds an SSH key to access the VM
func (a AzureClient) CreateVirtualMachine(ctx context.Context, resourceGroup, name, location, size, availabilitySetID, networkInterfaceID,
	username, sshPublicKey, imageName, customData string, storageAccount *storage.AccountProperties, isManaged bool,
	storageType string, diskSize int32) error {
	// TODO: "VM created from Image cannot have blob based disks. All disks have to be managed disks."
	imgReference, err := a.getImageReference(ctx, imageName, location)
	if err != nil {
		return err
	}

	log.Info("Creating virtual machine.", logutil.Fields{
		"name":     name,
		"location": location,
		"size":     size,
		"username": username,
		"osImage":  imageName,
	})

	sshKeyPath := fmt.Sprintf("/home/%s/.ssh/authorized_keys", username)
	log.Debugf("SSH key will be placed at: %s", sshKeyPath)

	var osProfile = &compute.OSProfile{
		ComputerName:  to.StringPtr(name),
		AdminUsername: to.StringPtr(username),
		LinuxConfiguration: &compute.LinuxConfiguration{
			DisablePasswordAuthentication: to.BoolPtr(true),
			SSH: &compute.SSHConfiguration{
				PublicKeys: &[]compute.SSHPublicKey{
					{
						Path:    to.StringPtr(sshKeyPath),
						KeyData: to.StringPtr(sshPublicKey),
					},
				},
			},
		},
	}

	if customData != "" {
		osProfile.CustomData = to.StringPtr(customData)
	}

	virtualMachinesClient := a.virtualMachinesClient()
	future, err := virtualMachinesClient.CreateOrUpdate(ctx, resourceGroup, name,
		compute.VirtualMachine{
			Location: to.StringPtr(location),
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				AvailabilitySet: &compute.SubResource{
					ID: to.StringPtr(availabilitySetID),
				},
				HardwareProfile: &compute.HardwareProfile{
					VMSize: compute.VirtualMachineSizeTypes(size),
				},
				NetworkProfile: &compute.NetworkProfile{
					NetworkInterfaces: &[]compute.NetworkInterfaceReference{
						{
							ID: to.StringPtr(networkInterfaceID),
						},
					},
				},
				OsProfile: osProfile,
				StorageProfile: &compute.StorageProfile{
					ImageReference: imgReference,
					OsDisk:         getOSDisk(name, storageAccount, isManaged, storageType, diskSize),
				},
			},
		})
	if err != nil {
		return err
	}
	if err = future.WaitForCompletionRef(ctx, virtualMachinesClient.Client); err != nil {
		return err
	}
	_, err = future.Result(virtualMachinesClient)
	return err
}

// getImageReference parses a publisher:offer:sku:version or parses the string as a custom image reference
func (a AzureClient) getImageReference(ctx context.Context, image, location string) (*compute.ImageReference, error) {
	if strings.Contains(strings.ToLower(image), "/images/") {
		// image represents an ARM resource identifer for a custom image
		return &compute.ImageReference{
			ID: to.StringPtr(image),
		}, nil
	}
	if urn := strings.Split(image, ":"); len(urn) == 4 {
		// image represents an ARM resource identifier for a gallery image version
		return &compute.ImageReference{
			Publisher: to.StringPtr(urn[0]),
			Offer:     to.StringPtr(urn[1]),
			Sku:       to.StringPtr(urn[2]),
			Version:   to.StringPtr(urn[3]),
		}, nil
	}
	return nil, fmt.Errorf("image provided must be an image URN or an ARM resource identifier")
}

// GetOSDisk creates and returns pointer to a disk that is configured for either managed or unmanaged disks depending
// on setting.
func getOSDisk(name string, account *storage.AccountProperties, isManaged bool, storageType string, diskSize int32) *compute.OSDisk {
	var osdisk *compute.OSDisk
	if isManaged {
		osdisk = &compute.OSDisk{
			Name:         to.StringPtr(ResourceNaming(name).OSDisk()),
			Caching:      compute.CachingTypesReadWrite,
			CreateOption: compute.DiskCreateOptionTypesFromImage,
			ManagedDisk: &compute.ManagedDiskParameters{
				StorageAccountType: compute.StorageAccountTypes(storageType),
			},
			DiskSizeGB: to.Int32Ptr(diskSize),
		}
	} else {
		osDiskBlobURL := osDiskStorageBlobURL(account, name)
		log.Debugf("OS disk blob will be placed at: %s", osDiskBlobURL)
		osdisk = &compute.OSDisk{
			Name:         to.StringPtr(ResourceNaming(name).OSDisk()),
			Caching:      compute.CachingTypesReadWrite,
			CreateOption: compute.DiskCreateOptionTypesFromImage,
			Vhd: &compute.VirtualHardDisk{
				URI: to.StringPtr(osDiskBlobURL),
			},
			DiskSizeGB: to.Int32Ptr(diskSize),
		}
	}
	return osdisk
}

// OSDiskExists queries to see if a particular OS Disk exists in the given resource group
func (a AzureClient) OSDiskExists(ctx context.Context, resourceGroup, name string) (bool, error) {
	disksClient := a.disksClient()
	_, err := disksClient.Get(ctx, resourceGroup, name)
	exists, err := checkResourceExistsFromError(err)
	if err != nil {
		return false, fmt.Errorf("Failed to query disks client: %s", err)
	}
	return exists, nil
}

// GetVirtualMachinePowerState returns the VM's power state
func (a AzureClient) GetVirtualMachinePowerState(ctx context.Context, resourceGroup, name string) (VMPowerState, error) {
	log.Debug("Querying instance view for power state.")
	vmInstanceView, err := a.virtualMachinesClient().InstanceView(ctx, resourceGroup, name)
	if err != nil {
		log.Errorf("Error querying instance view: %v", err)
		return Unknown, err
	}
	return powerStateFromInstanceView(&vmInstanceView), nil
}

// CreateAvailabilitySetIfNotExists checks that managed disk option match availability set if it already exists. If the
// availability set does not already exists than it is created with configured parameters.
func (a AzureClient) CreateAvailabilitySetIfNotExists(ctx context.Context, deploymentCtx *DeploymentContext, resourceGroup, name, location string, isManaged bool, faultCount int32, updateCount int32) error {
	var avSet compute.AvailabilitySet
	f := logutil.Fields{"name": name}
	log.Info("Configuring availability set.", f)

	avSetCleanupInfo := &avSetCleanup{rg: resourceGroup, name: name}
	err := avSetCleanupInfo.Get(ctx, a)
	exists, err := checkResourceExistsFromError(err)
	if err != nil {
		return fmt.Errorf("error getting availability set: %v", err)
	}
	if !exists {
		// availability set will be created because it has not been found
		// sku name dictates whether availability set is managed; Classic = non-managed, Aligned = managed
		skuName := "Classic"
		if isManaged {
			skuName = "Aligned"
		}
		avSet, err = a.availabilitySetsClient().CreateOrUpdate(ctx, resourceGroup, name,
			compute.AvailabilitySet{
				Location: to.StringPtr(location),
				AvailabilitySetProperties: &compute.AvailabilitySetProperties{
					PlatformFaultDomainCount:  to.Int32Ptr(faultCount),
					PlatformUpdateDomainCount: to.Int32Ptr(updateCount),
				},
				Sku: &compute.Sku{
					Name: to.StringPtr(skuName),
				},
			})
		if err != nil {
			return err
		}
	} else {
		avSet = avSetCleanupInfo.ref // value is set if err != nil
		// availability set has been found, and will only be checked for compatibility
		log.Infof("Availability set [%s] exists, will ignore configured faultDomainCount and updateDomainCount", name)
		if avSet.Sku == nil {
			return fmt.Errorf("cannot read sku of existing availability set")
		}
		if isManaged && to.String(avSet.Sku.Name) != "Aligned" {
			return fmt.Errorf("cannot convert non-managed availability set %s to managed availability set", name)
		}
		if !isManaged && to.String(avSet.Sku.Name) != "Classic" {
			return fmt.Errorf("cannot convert managed availability set %s to non-managed availability set", name)
		}
	}

	deploymentCtx.AvailabilitySetID = to.String(avSet.ID)
	return nil
}

// CleanupAvailabilitySetIfExists removes an availability set if there are no
// virtual machines attached to it. Note that this method is not safe for
// multiple concurrent writers, in case of races, deployment of a machine could
// fail or resource might not be cleaned up.
func (a AzureClient) CleanupAvailabilitySetIfExists(ctx context.Context, resourceGroup, name string) error {
	return a.cleanupResourceIfExists(ctx, &avSetCleanup{rg: resourceGroup, name: name})
}

// GetPublicIPAddress attempts to get public IP address from the Public IP
// resource. If IP address is not allocated yet, returns empty string. If
// useFqdn is set to true, the a FQDN hostname will be returned.
func (a AzureClient) GetPublicIPAddress(ctx context.Context, resourceGroup, name string, useFqdn bool) (string, error) {
	f := logutil.Fields{"name": name}
	log.Debug("Querying public IP address.", f)
	ip, err := a.publicIPAddressClient().Get(ctx, resourceGroup, name, "")
	if err != nil {
		return "", err
	}
	if ip.PublicIPAddressPropertiesFormat == nil {
		log.Debug("publicIP.Properties is nil. Could not determine IP address", f)
		return "", nil
	}

	if useFqdn { // return FQDN value on public IP
		log.Debug("Will attempt to return FQDN.", f)
		if ip.PublicIPAddressPropertiesFormat.DNSSettings == nil || ip.PublicIPAddressPropertiesFormat.DNSSettings.Fqdn == nil {
			return "", errors.New("FQDN not found on public IP address")
		}
		return to.String(ip.PublicIPAddressPropertiesFormat.DNSSettings.Fqdn), nil
	}
	return to.String(ip.PublicIPAddressPropertiesFormat.IPAddress), nil
}

// GetPrivateIPAddress attempts to retrieve private IP address of the specified
// network interface name.  If IP address is not allocated yet, returns empty
// string.
func (a AzureClient) GetPrivateIPAddress(ctx context.Context, resourceGroup, name string) (string, error) {
	f := logutil.Fields{"name": name}
	log.Debug("Querying network interface.", f)
	nic, err := a.networkInterfacesClient().Get(ctx, resourceGroup, name, "")
	if err != nil {
		return "", err
	}
	if nic.InterfacePropertiesFormat == nil || nic.InterfacePropertiesFormat.IPConfigurations == nil ||
		len(*nic.InterfacePropertiesFormat.IPConfigurations) == 0 {
		log.Debug("No IPConfigurations found on NIC", f)
		return "", nil
	}
	return to.String((*nic.InterfacePropertiesFormat.IPConfigurations)[0].InterfaceIPConfigurationPropertiesFormat.PrivateIPAddress), nil
}

// StartVirtualMachine starts the virtual machine and waits until it reaches
// the goal state (running) or times out.
func (a AzureClient) StartVirtualMachine(ctx context.Context, resourceGroup, name string) error {
	log.Info("Starting virtual machine.", logutil.Fields{"vm": name})
	virtualMachinesClient := a.virtualMachinesClient()
	future, err := virtualMachinesClient.Start(ctx, resourceGroup, name)
	if err != nil {
		return err
	}
	if err = future.WaitForCompletionRef(ctx, virtualMachinesClient.Client); err != nil {
		return err
	}
	if _, err := future.Result(virtualMachinesClient); err != nil {
		return err
	}
	return a.waitVMPowerState(ctx, resourceGroup, name, Running, waitStartTimeout)
}

// StopVirtualMachine power offs the virtual machine and waits until it reaches
// the goal state (stopped) or times out.
func (a AzureClient) StopVirtualMachine(ctx context.Context, resourceGroup, name string, skipShutdown bool) error {
	log.Info("Stopping virtual machine.", logutil.Fields{"vm": name})
	virtualMachinesClient := a.virtualMachinesClient()
	future, err := a.virtualMachinesClient().PowerOff(ctx, resourceGroup, name, &skipShutdown)
	if err != nil {
		return err
	}
	if err = future.WaitForCompletionRef(ctx, virtualMachinesClient.Client); err != nil {
		return err
	}
	if _, err := future.Result(virtualMachinesClient); err != nil {
		return err
	}
	return a.waitVMPowerState(ctx, resourceGroup, name, Stopped, waitPowerOffTimeout)
}

// RestartVirtualMachine restarts the virtual machine and waits until it reaches
// the goal state (stopped) or times out.
func (a AzureClient) RestartVirtualMachine(ctx context.Context, resourceGroup, name string) error {
	log.Info("Restarting virtual machine.", logutil.Fields{"vm": name})
	virtualMachinesClient := a.virtualMachinesClient()
	future, err := a.virtualMachinesClient().Restart(ctx, resourceGroup, name)
	if err != nil {
		return err
	}
	if err = future.WaitForCompletionRef(ctx, virtualMachinesClient.Client); err != nil {
		return err
	}
	if _, err := future.Result(virtualMachinesClient); err != nil {
		return err
	}
	return a.waitVMPowerState(ctx, resourceGroup, name, Running, waitStartTimeout)
}

// waitVMPowerState polls the Virtual Machine instance view until it reaches the
// specified goal power state or times out. If checking for virtual machine
// state fails or waiting times out, an error is returned.
func (a AzureClient) waitVMPowerState(ctx context.Context, resourceGroup, name string, goalState VMPowerState, timeout time.Duration) error {
	// NOTE(ahmetalpbalkan): Azure APIs for Start and Stop are actually async
	// operations on which our SDK blocks and does polling until the operation
	// is complete.
	//
	// By the time the issued power cycle operation is complete, the VM will be
	// already in the goal PowerState. Hence, this method will return in the
	// first check, however there is no harm in being defensive.
	log.Debug("Waiting until VM reaches goal power state.", logutil.Fields{
		"vm":        name,
		"goalState": goalState,
		"timeout":   timeout,
	})

	chErr := make(chan error)
	go func(ch chan error) {
		for {
			select {
			case <-ch:
				// channel closed
				return
			default:
				state, err := a.GetVirtualMachinePowerState(ctx, resourceGroup, name)
				if err != nil {
					ch <- err
					return
				}
				if state != goalState {
					log.Debug(fmt.Sprintf("Waiting %v...", powerStatePollingInterval),
						logutil.Fields{
							"goalState": goalState,
							"state":     state,
						})
					time.Sleep(powerStatePollingInterval)
				} else {
					log.Debug("Reached goal power state.",
						logutil.Fields{"state": state})
					ch <- nil
					return
				}
			}
		}
	}(chErr)

	select {
	case <-time.After(timeout):
		close(chErr)
		return fmt.Errorf("Waiting for goal state %q timed out after %v", goalState, timeout)
	case err := <-chErr:
		return err
	}
}

// checkExistsFromError inspects an error and returns a true if err is nil,
// false if error is an autorest.Error with StatusCode=404 and will return the
// error back if error is another status code or another type of error.
func checkResourceExistsFromError(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	v, ok := err.(autorest.DetailedError)
	if ok && v.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, v
}

// osDiskStorageBlobURL gives the full url of the VHD blob where the OS disk for
// the given VM should be stored.
func osDiskStorageBlobURL(account *storage.AccountProperties, vmName string) string {
	if account == nil {
		return ""
	}

	containerURL := osDiskStorageContainerURL(account, vmName) // has trailing slash
	blobName := ResourceNaming(vmName).OSDiskBlob()
	return containerURL + blobName
}

// osDiskStorageContainerURL crafts a URL with a trailing slash pointing
// to the full Azure Blob Container URL for given VM name.
func osDiskStorageContainerURL(account *storage.AccountProperties, vmName string) string {
	return fmt.Sprintf("%s%s/", to.String(account.PrimaryEndpoints.Blob), ResourceNaming(vmName).OSDiskContainer())
}

// extractStorageAccountFromVHDURL parses a blob URL and extracts the Azure
// Storage account name from the URL, namely first subdomain of the hostname and
// the Azure Storage service base URL (e.g. core.windows.net). If it could not
// be parsed, returns empty string.
func extractStorageAccountFromVHDURL(vhdURL string) string {
	u, err := url.Parse(vhdURL)
	if err != nil {
		log.Warn(fmt.Sprintf("URL parse error: %v", err), logutil.Fields{"url": vhdURL})
		return ""
	}
	parts := strings.SplitN(u.Host, ".", 2)
	if len(parts) != 2 {
		log.Warnf("Could not split account name and storage base URL: %s", vhdURL)
		return ""
	}
	return parts[0]
}
