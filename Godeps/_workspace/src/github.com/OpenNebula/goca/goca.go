package goca

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/kolo/xmlrpc"
	"launchpad.net/xmlpath"
)

var (
	client *oneClient
)

const (
	PoolWhoMine  = -3
	PoolWhoAll   = -2
	PoolWhoGroup = -1
)

type oneClient struct {
	token             string
	xmlrpcClient      *xmlrpc.Client
	xmlrpcClientError error
}

type response struct {
	status  bool
	body    string
	bodyInt int
}

type XMLResource struct {
	body string
}

type XMLIter struct {
	iter *xmlpath.Iter
}

type XMLNode struct {
	node *xmlpath.Node
}

func init() {
	err := SetClient()
	if err != nil {
		log.Fatal(err)
	}
}

func Client() *oneClient {
	return client
}

func SetClient(args ...string) error {
	var auth_token string
	var one_auth_path string

	if len(args) == 1 {
		auth_token = args[0]
	} else {
		one_auth_path = os.Getenv("ONE_AUTH")
		if one_auth_path == "" {
			one_auth_path = os.Getenv("HOME") + "/.one/one_auth"
		}

		token, err := ioutil.ReadFile(one_auth_path)
		if err == nil {
			auth_token = strings.TrimSpace(string(token))
		} else {
			auth_token = ""
		}
	}

	one_xmlrpc := os.Getenv("ONE_XMLRPC")
	if one_xmlrpc == "" {
		one_xmlrpc = "http://localhost:2633/RPC2"
	}

	xmlrpcClient, xmlrpcClientError := xmlrpc.NewClient(one_xmlrpc, nil)

	client = &oneClient{
		token:             auth_token,
		xmlrpcClient:      xmlrpcClient,
		xmlrpcClientError: xmlrpcClientError,
	}

	return nil
}

func SystemVersion() (string, error) {
	response, err := client.Call("one.system.version")
	if err != nil {
		return "", err
	}

	return response.Body(), nil
}

func (c *oneClient) Call(method string, args ...interface{}) (*response, error) {
	var (
		ok bool

		status  bool
		body    string
		bodyInt int64
	)

	if c.xmlrpcClientError != nil {
		return nil, errors.New(fmt.Sprintf("Unitialized client. Token: '%s', xmlrpcClient: '%s'", c.token, c.xmlrpcClientError))
	}

	result := []interface{}{}

	xmlArgs := make([]interface{}, len(args)+1)

	xmlArgs[0] = c.token
	copy(xmlArgs[1:], args[:])

	err := c.xmlrpcClient.Call(method, xmlArgs, &result)
	if err != nil {
		log.Fatal(err)
	}

	status, ok = result[0].(bool)
	if ok == false {
		log.Fatal("Unexpected XML-RPC response. Expected: Index 0 Boolean")
	}

	body, ok = result[1].(string)
	if ok == false {
		bodyInt, ok = result[1].(int64)
		if ok == false {
			log.Fatal("Unexpected XML-RPC response. Expected: Index 0 Int or String")
		}
	}

	// TODO: errCode? result[2]

	r := &response{status, body, int(bodyInt)}

	if status == false {
		err = errors.New(body)
	}

	return r, err
}

func (r *response) Body() string {
	return r.body
}

func (r *response) BodyInt() int {
	return r.bodyInt
}

func (r *XMLResource) Body() string {
	return r.body
}

func (r *XMLResource) XPath(xpath string) (string, bool) {
	path := xmlpath.MustCompile(xpath)
	b := bytes.NewBufferString(r.Body())

	root, _ := xmlpath.Parse(b)

	return path.String(root)
}

func (r *XMLResource) XPathIter(xpath string) *XMLIter {
	path := xmlpath.MustCompile(xpath)
	b := bytes.NewBufferString(string(r.Body()))

	root, _ := xmlpath.Parse(b)

	return &XMLIter{iter: path.Iter(root)}
}

func (r *XMLResource) GetIdFromName(name string, xpath string) (uint, error) {
	var id int
	var match bool = false

	iter := r.XPathIter(xpath)
	for iter.Next() {
		node := iter.Node()

		n, _ := node.XPathNode("NAME")
		if n == name {
			if match {
				return 0, errors.New("Multiple resources with that name.")
			}

			idString, _ := node.XPathNode("ID")
			id, _ = strconv.Atoi(idString)
			match = true
		}
	}

	if match {
		return uint(id), nil
	} else {
		return 0, errors.New("Resource not found.")
	}
}

func (i *XMLIter) Next() bool {
	return i.iter.Next()
}

func (i *XMLIter) Node() *XMLNode {
	return &XMLNode{node: i.iter.Node()}
}

func (n *XMLNode) XPathNode(xpath string) (string, bool) {
	path := xmlpath.MustCompile(xpath)
	return path.String(n.node)
}
