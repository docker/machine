package crowbar
// Apache 2 License 2015 by Rob Hirschfeld for RackN

import (
    "fmt"
    "net/http"
    "strconv"
	"errors"
)

// Cost assumes global session created with NewClient
var session *OCBClient

const (
    API_PATH        = "/api/v2/"
    SSH_KEY_ROLE    = "crowbar-access_keys"
    TARGET_OS       = "provisioner-target_os"
)

func NewClient (User, Password, URL string) (*OCBClient, error) {   
    c := OCBClient{URL: URL, Client: &http.Client{}, Challenge: &challenge{} }
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
    return &c, nil
}

// DEPLOYMENT SPECIFIC

func (d *Deployment) get(deployment string) (err error) {
    return session.Request("GET", "deployments/" + deployment, "deployment.obj", &d, nil)
}

func (d *Deployment) add(newd *NewDeployment) (err error) {
    if newd.ParentID == 0 {
        system := &Deployment{}
        system.get("system")
        newd.ParentID = system.ID
    }
    return session.Request("POST", "deployments", "deployment.obj", d, newd)
}

func (d *Deployment) update() (err error) {
    nd := NewDeployment{
        Name: d.Name,
        Description: d.Description,
        ParentID: d.ParentID}
    return session.Request("PUT", "deployments/" + strconv.FormatInt(d.ID,10), "deployment.obj", d, nd)
}

func (d *Deployment) delete() (err error) {
    return session.Request("DELETE", "deployments/" + strconv.FormatInt(d.ID,10), "deployment.obj", d, nil)
}

func (d *Deployment) commit() (err error) {
    return session.Request("PUT", "deployments/" + strconv.FormatInt(d.ID,10) + "/commit", "json.array", nil, nil)
}


func (d *Deployment) nodes() ([]Node, error) {
 	raw := &[]Node{}
 	err := session.Request("GET", "deployments/" + strconv.FormatInt(d.ID,10) + "/nodes", "node.list", &raw, nil)
	list := make([]Node, len(*raw))
	list = *raw
    return list, err
}

func deployments() ([]Deployment, error) {
 	raw := &[]Deployment{}
 	err := session.Request("GET", "deployments", "deployment.list", &raw, nil)
	list := make([]Deployment, len(*raw))
	list = *raw
    return list, err
}

// NODE SPECIFIC

func (n *Node) get(node string) (err error) {
    return session.Request("GET", "nodes/" + node, "node.obj", &n, nil)
}

func (n *Node) refresh() (err error) {
    nn := &Node{}
    err = session.Request("GET", "nodes/" + strconv.FormatInt(n.ID,10), "node.obj", &nn, nil)
    n = nn
    return err 
}

func (n *Node) role(role string) (nr *NodeRole, err error) {
	nr = &NodeRole{}
    return nr, session.Request("GET", "nodes/" + strconv.FormatInt(n.ID,10) + "/node_roles/" + role, "node_role.obj", &nr, nil)
}

func (n *Node) commit() (err error) {
	return session.Request("PUT", "nodes/" + strconv.FormatInt(n.ID,10) + "/commit", "json.array", nil, nil)
}

func (n *Node) propose() (err error) {
    return session.Request("PUT", "nodes/" + strconv.FormatInt(n.ID,10) + "/propose", "json.array", nil, nil)
}

func (n *Node) redeploy() (err error) {
    return session.Request("PUT", "nodes/" + strconv.FormatInt(n.ID,10) + "/redeploy", "json.array", nil, nil)
}

// note: you MUST propose -> commit the node for this change to take effect
func (n *Node) addSSHkey(index int, value string) (err error) {
    key := n.Name + "-" + strconv.Itoa(index) 
    nav := &NodeAttribValue{}
    nav.Value = make(map[string]string)
    nav.Value[key] = value
    return session.Request("PUT", "nodes/"+ strconv.FormatInt(n.ID,10) + "/attribs/" + SSH_KEY_ROLE, "attrib.obj", nil, &nav)
}

func (n *Node) setOS(value string) (err error) {
    // this simpler version of the REST API uses the attrib input instead of passing JSON
    return session.Request("PUT", "nodes/"+ strconv.FormatInt(n.ID,10) + "/attribs/" + TARGET_OS + "?value=" + value, "attrib.obj", nil, nil)
}

func (n *Node) attrib(attrib string) (*NodeAttrib, error) {
    na := &NodeAttrib{}
    err := session.Request("GET", "nodes/" + strconv.FormatInt(n.ID,10) + "/attribs/" + attrib, "attrib.obj", &na, nil)
    return na, err
}

func (n *Node) power(target string, alternate string) (err error) {
	array := make(map[string]string)
	err = session.Request("GET", "nodes/" + strconv.FormatInt(n.ID,10) + "/power", "json.array", &array, nil)
    _, target_ok := array[target]
    _, alternate_ok := array[alternate]
    // make sure we can run command
    if !target_ok {
        if !alternate_ok {
            return fmt.Errorf("Power for node %s does not include %s or %s.  Choices are: %s", n.Name, target, alternate, array)
        } 
        target = alternate
    }
	return session.Request("PUT", "nodes/" + strconv.FormatInt(n.ID,10) + "/power?poweraction=" + target, "json.array", nil, nil)
}

func (n *Node) address(category string) ([]string, error) {
    addresses, err := n.addresses()
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

func (n *Node) addresses() ([]NodeAddress, error) {
    raw := &[]NodeAddress{}
    err := session.Request("GET", "nodes/" + strconv.FormatInt(n.ID,10) + "/addresses", "addresses.array", &raw, nil)
    addresses := make([]NodeAddress, len(*raw))
    addresses = *raw
    return addresses, err
}

func (n *Node) add(newn *NewNode) (err error) {
    if newn.DeploymentID == 0 {
    	return errors.New("Cannot create a node without a deployment.")
    }
    return session.Request("POST", "nodes", "node.obj", n, newn)
}

func (n *Node) update() (err error) {
    nn := NewNode{
        Name: n.Name,
        Description: n.Description,
        Alias: n.Alias,
        Available: n.Available,
        Bootenv: n.Bootenv,
        DeploymentID: n.DeploymentID,
        Order: n.Order}
    return session.Request("PUT", "nodes/" + strconv.FormatInt(n.ID,10), "node.obj", n, nn)
}

func (n *Node) delete() (err error) {
    return session.Request("DELETE", "nodes/" + strconv.FormatInt(n.ID,10), "node.obj", n, nil)
}

func nodes() ([]Node, error) {
 	raw := &[]Node{}
 	err := session.Request("GET", "nodes", "node.list", &raw, nil)
	list := make([]Node, len(*raw))
	list = *raw
    return list, err
}

func (r *Role) get(role string) (err error) {
    return session.Request("GET", "roles/" + role, "role.obj", &r, nil)
}

// NODE ROLE SPECIFIC

// look for failed node roles and retry them
func (n *Node) retry() (err error) {
    raw := &[]NodeRole{}
    err = session.Request("GET", "nodes/" + strconv.FormatInt(n.ID,10) + "/node_roles", "node_role.list", &raw, nil)
    roles := make([]NodeRole, len(*raw))
    roles = *raw
    if err != nil {
        return err
    }
    for i := range roles {
        if roles[i].State == -1 {
            roles[i].retry()
        }
    }
    return nil
}

// retry a specific node role

func (nr *NodeRole) delete() (err error) {
    return session.Request("DELETE", "node_roles/" + strconv.FormatInt(nr.ID,10), "node_role.obj", nil, nil)
}


func (nr *NodeRole) retry() (err error) {
    return session.Request("PUT", "node_roles/" + strconv.FormatInt(nr.ID,10) + "/retry", "node_role.obj", nil, nil)
}

func (nr *NodeRole) add() (err error) {
    nnr := NewNodeRole{
        DeploymentID: nr.DeploymentID,
        RoleID: nr.RoleID,
        NodeID: nr.NodeID,
        Order: nr.Order,
        ProposedData: nr.ProposedData}
	if nnr.Order == 0 {
		nnr.Order = 1000
	}
    return session.Request("POST", "node_roles", "node_role.obj", nr, nnr)
}

// SYSTEM WIDE REQUESTS

func osAvailable(os string) bool {
    available := availableOS()
    for _, check := range available {
        if check == os {
            return true
        }
    }
    return false
}

func availableOS() []string {
    na := &NodeAttrib{}
    // search for available os information by scanning all the nodes (should be id =2)
    err := session.Request("GET", "nodes/1/attribs/provisioner-available-oses", "attrib.obj", &na, nil)
    for i := 2; err != nil && i<1000; i += 1 {
        err = session.Request("GET", "nodes/" + strconv.Itoa(i) + "/attribs/provisioner-available-oses", "attrib.obj", &na, nil)
    }
    aos := make([]string, 0)
    if err == nil {
        for k := range na.Value {
            aos = append(aos, k)
        }
    }
    return aos
}