package egoscale

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/jinzhu/copier"
)

// SecurityGroup represent a firewalling set of rules
type SecurityGroup struct {
	ID                  string        `json:"id"`
	Account             string        `json:"account,omitempty"`
	Description         string        `json:"description,omitempty"`
	Domain              string        `json:"domain,omitempty"`
	DomainID            string        `json:"domainid,omitempty"`
	Name                string        `json:"name"`
	Project             string        `json:"project,omitempty"`
	ProjectID           string        `json:"projectid,omitempty"`
	VirtualMachineCount int           `json:"virtualmachinecount,omitempty"` // CloudStack 4.6+
	VirtualMachineIDs   []string      `json:"virtualmachineids,omitempty"`   // CloudStack 4.6+
	IngressRule         []IngressRule `json:"ingressrule"`
	EgressRule          []EgressRule  `json:"egressrule"`
	Tags                []ResourceTag `json:"tags,omitempty"`
	JobID               string        `json:"jobid,omitempty"`
	JobStatus           JobStatusType `json:"jobstatus,omitempty"`
}

// ResourceType returns the type of the resource
func (*SecurityGroup) ResourceType() string {
	return "SecurityGroup"
}

// Get loads the given Security Group
func (sg *SecurityGroup) Get(ctx context.Context, client *Client) error {
	if sg.ID == "" && sg.Name == "" {
		return fmt.Errorf("A SecurityGroup may only be searched using ID or Name")
	}
	resp, err := client.RequestWithContext(ctx, &ListSecurityGroups{
		ID:                sg.ID,
		SecurityGroupName: sg.Name,
	})

	if err != nil {
		return err
	}

	sgs := resp.(*ListSecurityGroupsResponse)
	count := len(sgs.SecurityGroup)
	if count == 0 {
		err := &ErrorResponse{
			ErrorCode: ParamError,
			ErrorText: fmt.Sprintf("SecurityGroup not found id: %s, name: %s", sg.ID, sg.Name),
		}
		return err
	} else if count > 1 {
		return fmt.Errorf("More than one SecurityGroup was found. Query: id: %s, name: %s", sg.ID, sg.Name)
	}

	return copier.Copy(sg, sgs.SecurityGroup[0])
}

// Delete deletes the given Security Group
func (sg *SecurityGroup) Delete(ctx context.Context, client *Client) error {
	if sg.ID == "" && sg.Name == "" {
		return fmt.Errorf("A SecurityGroup may only be deleted using ID or Name")
	}

	req := &DeleteSecurityGroup{
		Account:   sg.Account,
		DomainID:  sg.DomainID,
		ProjectID: sg.ProjectID,
	}

	if sg.ID != "" {
		req.ID = sg.ID
	} else {
		req.Name = sg.Name
	}

	return client.BooleanRequestWithContext(ctx, req)
}

// IngressRule represents the ingress rule
type IngressRule struct {
	RuleID                string              `json:"ruleid"`
	Account               string              `json:"account,omitempty"`
	Cidr                  string              `json:"cidr,omitempty"`
	Description           string              `json:"description,omitempty"`
	IcmpType              int                 `json:"icmptype,omitempty"`
	IcmpCode              int                 `json:"icmpcode,omitempty"`
	StartPort             int                 `json:"startport,omitempty"`
	EndPort               int                 `json:"endport,omitempty"`
	Protocol              string              `json:"protocol,omitempty"`
	Tags                  []ResourceTag       `json:"tags,omitempty"`
	SecurityGroupID       string              `json:"securitygroupid,omitempty"`
	SecurityGroupName     string              `json:"securitygroupname,omitempty"`
	UserSecurityGroupList []UserSecurityGroup `json:"usersecuritygrouplist,omitempty"`
	JobID                 string              `json:"jobid,omitempty"`
	JobStatus             JobStatusType       `json:"jobstatus,omitempty"`
}

// EgressRule represents the ingress rule
type EgressRule IngressRule

// UserSecurityGroup represents the traffic of another security group
type UserSecurityGroup struct {
	Group   string `json:"group,omitempty"`
	Account string `json:"account,omitempty"`
}

// SecurityGroupResponse represents a generic security group response
type SecurityGroupResponse struct {
	SecurityGroup SecurityGroup `json:"securitygroup"`
}

// CreateSecurityGroup represents a security group creation
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/createSecurityGroup.html
type CreateSecurityGroup struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// APIName returns the CloudStack API command name
func (*CreateSecurityGroup) APIName() string {
	return "createSecurityGroup"
}

func (*CreateSecurityGroup) response() interface{} {
	return new(CreateSecurityGroupResponse)
}

// CreateSecurityGroupResponse represents a new security group
type CreateSecurityGroupResponse SecurityGroupResponse

// DeleteSecurityGroup represents a security group deletion
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/deleteSecurityGroup.html
type DeleteSecurityGroup struct {
	Account   string `json:"account,omitempty"`  // Must be specified with Domain ID
	DomainID  string `json:"domainid,omitempty"` // Must be specified with Account
	ID        string `json:"id,omitempty"`       // Mutually exclusive with name
	Name      string `json:"name,omitempty"`     // Mutually exclusive with id
	ProjectID string `json:"project,omitempty"`
}

// APIName returns the CloudStack API command name
func (*DeleteSecurityGroup) APIName() string {
	return "deleteSecurityGroup"
}

func (*DeleteSecurityGroup) response() interface{} {
	return new(booleanSyncResponse)
}

// AuthorizeSecurityGroupIngress (Async) represents the ingress rule creation
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/authorizeSecurityGroupIngress.html
type AuthorizeSecurityGroupIngress struct {
	Account               string              `json:"account,omitempty"`
	CidrList              []string            `json:"cidrlist,omitempty"`
	Description           string              `json:"description,omitempty"`
	DomainID              string              `json:"domainid,omitempty"`
	IcmpType              int                 `json:"icmptype,omitempty"`
	IcmpCode              int                 `json:"icmpcode,omitempty"`
	StartPort             int                 `json:"startport,omitempty"`
	EndPort               int                 `json:"endport,omitempty"`
	ProjectID             string              `json:"projectid,omitempty"`
	Protocol              string              `json:"protocol,omitempty"`
	SecurityGroupID       string              `json:"securitygroupid,omitempty"`
	SecurityGroupName     string              `json:"securitygroupname,omitempty"`
	UserSecurityGroupList []UserSecurityGroup `json:"usersecuritygrouplist,omitempty"`
}

// APIName returns the CloudStack API command name
func (*AuthorizeSecurityGroupIngress) APIName() string {
	return "authorizeSecurityGroupIngress"
}

func (*AuthorizeSecurityGroupIngress) asyncResponse() interface{} {
	return new(AuthorizeSecurityGroupIngressResponse)
}

func (req *AuthorizeSecurityGroupIngress) onBeforeSend(params *url.Values) error {
	// ICMP code and type may be zero but can also be omitted...
	if req.Protocol == "ICMP" {
		params.Set("icmpcode", strconv.FormatInt(int64(req.IcmpCode), 10))
		params.Set("icmptype", strconv.FormatInt(int64(req.IcmpType), 10))
	}
	return nil
}

// AuthorizeSecurityGroupIngressResponse represents the new egress rule
// /!\ the Cloud Stack API document is not fully accurate. /!\
type AuthorizeSecurityGroupIngressResponse SecurityGroupResponse

// AuthorizeSecurityGroupEgress (Async) represents the egress rule creation
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/authorizeSecurityGroupEgress.html
type AuthorizeSecurityGroupEgress AuthorizeSecurityGroupIngress

// APIName returns the CloudStack API command name
func (*AuthorizeSecurityGroupEgress) APIName() string {
	return "authorizeSecurityGroupEgress"
}

func (*AuthorizeSecurityGroupEgress) asyncResponse() interface{} {
	return new(AuthorizeSecurityGroupEgressResponse)
}

func (req *AuthorizeSecurityGroupEgress) onBeforeSend(params *url.Values) error {
	return (*AuthorizeSecurityGroupIngress)(req).onBeforeSend(params)
}

// AuthorizeSecurityGroupEgressResponse represents the new egress rule
// /!\ the Cloud Stack API document is not fully accurate. /!\
type AuthorizeSecurityGroupEgressResponse CreateSecurityGroupResponse

// RevokeSecurityGroupIngress (Async) represents the ingress/egress rule deletion
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/revokeSecurityGroupIngress.html
type RevokeSecurityGroupIngress struct {
	ID string `json:"id"`
}

// APIName returns the CloudStack API command name
func (*RevokeSecurityGroupIngress) APIName() string {
	return "revokeSecurityGroupIngress"
}

func (*RevokeSecurityGroupIngress) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// RevokeSecurityGroupEgress (Async) represents the ingress/egress rule deletion
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/revokeSecurityGroupEgress.html
type RevokeSecurityGroupEgress struct {
	ID string `json:"id"`
}

// APIName returns the CloudStack API command name
func (*RevokeSecurityGroupEgress) APIName() string {
	return "revokeSecurityGroupEgress"
}

func (*RevokeSecurityGroupEgress) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// ListSecurityGroups represents a search for security groups
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/listSecurityGroups.html
type ListSecurityGroups struct {
	Account           string        `json:"account,omitempty"`
	DomainID          string        `json:"domainid,omitempty"`
	ID                string        `json:"id,omitempty"`
	IsRecursive       *bool         `json:"isrecursive,omitempty"`
	Keyword           string        `json:"keyword,omitempty"`
	ListAll           *bool         `json:"listall,omitempty"`
	Page              int           `json:"page,omitempty"`
	PageSize          int           `json:"pagesize,omitempty"`
	ProjectID         string        `json:"projectid,omitempty"`
	Type              string        `json:"type,omitempty"`
	SecurityGroupName string        `json:"securitygroupname,omitempty"`
	Tags              []ResourceTag `json:"tags,omitempty"`
	VirtualMachineID  string        `json:"virtualmachineid,omitempty"`
}

// APIName returns the CloudStack API command name
func (*ListSecurityGroups) APIName() string {
	return "listSecurityGroups"
}

func (*ListSecurityGroups) response() interface{} {
	return new(ListSecurityGroupsResponse)
}

// ListSecurityGroupsResponse represents a list of security groups
type ListSecurityGroupsResponse struct {
	Count         int             `json:"count"`
	SecurityGroup []SecurityGroup `json:"securitygroup"`
}

// CreateIngressRule creates a set of ingress rules
//
// Deprecated: use the API directly
func (exo *Client) CreateIngressRule(req *AuthorizeSecurityGroupIngress) ([]IngressRule, error) {
	resp, err := exo.Request(req)
	if err != nil {
		return nil, err
	}
	return resp.(*AuthorizeSecurityGroupIngressResponse).SecurityGroup.IngressRule, nil
}

// CreateEgressRule creates a set of egress rules
//
// Deprecated: use the API directly
func (exo *Client) CreateEgressRule(req *AuthorizeSecurityGroupEgress) ([]EgressRule, error) {
	resp, err := exo.Request(req)
	if err != nil {
		return nil, err
	}
	return resp.(*AuthorizeSecurityGroupEgressResponse).SecurityGroup.EgressRule, nil
}

// CreateSecurityGroupWithRules create a security group with its rules
// Warning: it doesn't rollback in case of a failure!
//
// Deprecated: use the API directly
func (exo *Client) CreateSecurityGroupWithRules(name string, ingress []AuthorizeSecurityGroupIngress, egress []AuthorizeSecurityGroupEgress) (*SecurityGroup, error) {
	req := &CreateSecurityGroup{
		Name: name,
	}
	resp, err := exo.Request(req)
	if err != nil {
		return nil, err
	}

	sg := resp.(*SecurityGroupResponse).SecurityGroup
	reqs := make([]asyncCommand, 0, len(ingress)+len(egress))
	// Egress rules
	for _, ereq := range egress {
		ereq.SecurityGroupID = sg.ID
		reqs = append(reqs, &ereq)

	}
	// Ingress rules
	for _, ireq := range ingress {
		ireq.SecurityGroupID = sg.ID
		reqs = append(reqs, &ireq)
	}

	for _, r := range reqs {
		_, err := exo.Request(r)
		if err != nil {
			return nil, err
		}
	}

	r, err := exo.Request(&ListSecurityGroups{
		ID: sg.ID,
	})
	if err != nil {
		return nil, err
	}

	sg = r.(*ListSecurityGroupsResponse).SecurityGroup[0]
	return &sg, nil
}
