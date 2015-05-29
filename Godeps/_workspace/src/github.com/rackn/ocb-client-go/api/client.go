package crowbar
// Apache 2 License 2015 by Rob Hirschfeld for RackN

import (
    "fmt"
    "net/http"
    "strconv"
	"errors"
)

// OCB assumes global session created with NewClient
var Session *OCBClient

const (
    API_PATH        = "/api/v2/"
    SSH_KEY_ROLE    = "crowbar-access_keys"
    TARGET_OS       = "provisioner-target_os"
)

func NewClient (User, Password, URL string) (*OCBClient, error) {   
    c := &OCBClient{URL: URL, Client: &http.Client{}, Challenge: &challenge{} }
    // retrieve the digest info from the 301 message 
    resp, e := http.Head(URL + API_PATH + "digest")
    if resp.StatusCode != 401 {
        return nil, fmt.Errorf("Expected Digest Challenge Missing on URL %s got %s", URL, resp.Status)
    } else if e != nil {
        return nil, e
    }
    var err error
    err = c.Challenge.parseChallenge(resp.Header.Get("WWW-Authenticate"))
    if err != nil {
        return nil, err
    }
    c.Challenge.Username = User
    c.Challenge.Password = Password
    return c, nil
}

// DEPLOYMENT SPECIFIC

func (d *Deployment) Get(deployment string) (err error) {
    return Session.Request("GET", "deployments/" + deployment, "deployment.obj", &d, nil)
}

func (d *Deployment) Add(newd *NewDeployment) (err error) {
    if newd.ParentID == 0 {
        system := &Deployment{}
        system.Get("system")
        newd.ParentID = system.ID
    }
    return Session.Request("POST", "deployments", "deployment.obj", d, newd)
}

func (d *Deployment) Update() (err error) {
    nd := NewDeployment{
        Name: d.Name,
        Description: d.Description,
        ParentID: d.ParentID}
    return Session.Request("PUT", "deployments/" + strconv.FormatInt(d.ID,10), "deployment.obj", d, nd)
}

func (d *Deployment) Delete() (err error) {
    return Session.Request("DELETE", "deployments/" + strconv.FormatInt(d.ID,10), "deployment.obj", d, nil)
}

func (d *Deployment) Commit() (err error) {
    return Session.Request("PUT", "deployments/" + strconv.FormatInt(d.ID,10) + "/commit", "json.array", nil, nil)
}


func (d *Deployment) Nodes() ([]Node, error) {
 	raw := &[]Node{}
 	err := Session.Request("GET", "deployments/" + strconv.FormatInt(d.ID,10) + "/nodes", "node.list", &raw, nil)
	list := make([]Node, len(*raw))
	list = *raw
    return list, err
}

func Deployments() ([]Deployment, error) {
 	raw := &[]Deployment{}
 	err := Session.Request("GET", "deployments", "deployment.list", &raw, nil)
	list := make([]Deployment, len(*raw))
	list = *raw
    return list, err
}

// NODE SPECIFIC

func (n *Node) Get(node string) (err error) {
    return Session.Request("GET", "nodes/" + node, "node.obj", &n, nil)
}

func (n *Node) Refresh() (err error) {
    nn := &Node{}
    err = Session.Request("GET", "nodes/" + strconv.FormatInt(n.ID,10), "node.obj", &nn, nil)
    n = nn
    return err 
}

func (n *Node) Role(role string) (nr *NodeRole, err error) {
	nr = &NodeRole{}
    return nr, Session.Request("GET", "nodes/" + strconv.FormatInt(n.ID,10) + "/node_roles/" + role, "node_role.obj", &nr, nil)
}

func (n *Node) Commit() (err error) {
	return Session.Request("PUT", "nodes/" + strconv.FormatInt(n.ID,10) + "/commit", "json.array", nil, nil)
}

func (n *Node) Propose() (err error) {
    return Session.Request("PUT", "nodes/" + strconv.FormatInt(n.ID,10) + "/propose", "json.array", nil, nil)
}

func (n *Node) Redeploy() (err error) {
    return Session.Request("PUT", "nodes/" + strconv.FormatInt(n.ID,10) + "/redeploy", "json.array", nil, nil)
}

// note: you MUST propose -> commit the node for this change to take effect
func (n *Node) AddSSHkey(index int, value string) (err error) {
    key := n.Name + "-" + strconv.Itoa(index) 
    nav := &NodeAttribValue{}
    nav.Value = make(map[string]string)
    nav.Value[key] = value
    return Session.Request("PUT", "nodes/"+ strconv.FormatInt(n.ID,10) + "/attribs/" + SSH_KEY_ROLE, "attrib.obj", nil, &nav)
}

func (n *Node) SetOS(value string) (err error) {
    // this simpler version of the REST API uses the attrib input instead of passing JSON
    return Session.Request("PUT", "nodes/"+ strconv.FormatInt(n.ID,10) + "/attribs/" + TARGET_OS + "?value=" + value, "attrib.obj", nil, nil)
}

func (n *Node) Attrib(attrib string) (*NodeAttrib, error) {
    na := &NodeAttrib{}
    err := Session.Request("GET", "nodes/" + strconv.FormatInt(n.ID,10) + "/attribs/" + attrib, "attrib.obj", &na, nil)
    return na, err
}

func (n *Node) Power(target string, alternate string) (err error) {
	array := make(map[string]string)
	err = Session.Request("GET", "nodes/" + strconv.FormatInt(n.ID,10) + "/power", "json.array", &array, nil)
    _, target_ok := array[target]
    _, alternate_ok := array[alternate]
    // make sure we can run command
    if !target_ok {
        if !alternate_ok {
            return fmt.Errorf("Power for node %s does not include %s or %s.  Choices are: %s", n.Name, target, alternate, array)
        } 
        target = alternate
    }
	return Session.Request("PUT", "nodes/" + strconv.FormatInt(n.ID,10) + "/power?poweraction=" + target, "json.array", nil, nil)
}

func (n *Node) Address(category string) ([]string, error) {
    addresses, err := n.Addresses()
    if err != nil {
        return nil, err
    }
    var addr *NodeAddress
    for i:=0; i<len(addresses) && addr == nil; i+=1 {
        if addresses[i].Category == category {
            addr = &addresses[i]
        }
    }
    if addr == nil {
        return nil, fmt.Errorf("No Addreses in category %s for node %s found in %v", category, n.Name, addresses)
    }
    return addr.Addresses, err
}

func (n *Node) Addresses() ([]NodeAddress, error) {
    raw := &[]NodeAddress{}
    err := Session.Request("GET", "nodes/" + strconv.FormatInt(n.ID,10) + "/addresses", "addresses.array", &raw, nil)
    addresses := make([]NodeAddress, len(*raw))
    addresses = *raw
    return addresses, err
}

func (n *Node) Add(newn *NewNode) (err error) {
    if newn.DeploymentID == 0 {
    	return errors.New("Cannot create a node without a deployment.")
    }
    return Session.Request("POST", "nodes", "node.obj", n, newn)
}

func (n *Node) Update() (err error) {
    nn := NewNode{
        Name: n.Name,
        Description: n.Description,
        Alias: n.Alias,
        Available: n.Available,
        Bootenv: n.Bootenv,
        DeploymentID: n.DeploymentID,
        Order: n.Order}
    return Session.Request("PUT", "nodes/" + strconv.FormatInt(n.ID,10), "node.obj", n, nn)
}

func (n *Node) Delete() (err error) {
    return Session.Request("DELETE", "nodes/" + strconv.FormatInt(n.ID,10), "node.obj", n, nil)
}

func Nodes() ([]Node, error) {
 	raw := &[]Node{}
 	err := Session.Request("GET", "nodes", "node.list", &raw, nil)
	list := make([]Node, len(*raw))
	list = *raw
    return list, err
}

func (r *Role) Get(role string) (err error) {
    return Session.Request("GET", "roles/" + role, "role.obj", &r, nil)
}

// NODE ROLE SPECIFIC

// look for failed node roles and retry them
func (n *Node) Retry() (err error) {
    raw := &[]NodeRole{}
    err = Session.Request("GET", "nodes/" + strconv.FormatInt(n.ID,10) + "/node_roles", "node_role.list", &raw, nil)
    roles := make([]NodeRole, len(*raw))
    roles = *raw
    if err != nil {
        return err
    }
    for i := range roles {
        if roles[i].State == -1 {
            roles[i].Retry()
        }
    }
    return nil
}

// retry a specific node role

func (nr *NodeRole) Delete() (err error) {
    return Session.Request("DELETE", "node_roles/" + strconv.FormatInt(nr.ID,10), "node_role.obj", nil, nil)
}


func (nr *NodeRole) Retry() (err error) {
    return Session.Request("PUT", "node_roles/" + strconv.FormatInt(nr.ID,10) + "/retry", "node_role.obj", nil, nil)
}

func (nr *NodeRole) Add() (err error) {
    nnr := NewNodeRole{
        DeploymentID: nr.DeploymentID,
        RoleID: nr.RoleID,
        NodeID: nr.NodeID,
        Order: nr.Order,
        ProposedData: nr.ProposedData}
	if nnr.Order == 0 {
		nnr.Order = 1000
	}
    return Session.Request("POST", "node_roles", "node_role.obj", nr, nnr)
}

// SYSTEM WIDE REQUESTS

func OsAvailable(os string) bool {
    available := AvailableOS()
    for _, check := range available {
        if check == os {
            return true
        }
    }
    return false
}

func AvailableOS() []string {
    na := &NodeAttrib{}
    // search for available os information by scanning all the nodes (should be id =2)
    err := Session.Request("GET", "nodes/1/attribs/provisioner-available-oses", "attrib.obj", &na, nil)
    for i := 2; err != nil && i<1000; i += 1 {
        err = Session.Request("GET", "nodes/" + strconv.Itoa(i) + "/attribs/provisioner-available-oses", "attrib.obj", &na, nil)
    }
    aos := make([]string, 0)
    if err == nil {
        for k := range na.Value {
            aos = append(aos, k)
        }
    }
    return aos
}