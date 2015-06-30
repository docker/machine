package main

import (
	"encoding/base64"
	"fmt"
	"github.com/nilshell/xmlrpc"
	"io/ioutil"
	"log"
	"reflect"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	client, _ := xmlrpc.NewClient("http://bugzilla-server/bugzilla/xmlrpc.cgi", nil)
	result := xmlrpc.Struct{}

	// login
	err := client.Call("User.login",
		xmlrpc.Struct{"login": "test@localhost.localdomain",
			"password": "password"},
		&result)
	check(err)
	//fmt.Printf("User.login returned: %v\n", result)

	// get attachment data
	bug_number := "3"

	err = client.Call(
		"Bug.attachments",
		xmlrpc.Struct{
			"ids": []string{bug_number},
			"include_fields": []string{
				"file_name", "id", "description",
				"data", "is_obsolete"},
		},
		&result,
	)
	check(err)
	fmt.Printf("Bug.attachments() returned: %v\n\n", result)

	a := reflect.ValueOf(result["bugs"].(xmlrpc.Struct)[bug_number])

	if a.Len() == 0 {
		fmt.Printf("Bug id %s has no attachments.\n", bug_number)
		return
	}

	file_map := make(map[string]string)
	for i := 0; i < a.Len(); i++ {
		attachment := a.Index(i).Interface().(xmlrpc.Struct)

		fmt.Printf("attachment [%d] ::\n", i)
		fmt.Printf("\tfile_name : %v\n", attachment["file_name"])
		fmt.Printf("\tdescription : %v\n", attachment["description"])
		fmt.Printf("\tid : %v\n", attachment["id"])
		fmt.Printf("\tis_obsolete : %v\n", attachment["is_obsolete"])
		fmt.Printf("\tdata : %v\n", attachment["data"].(string))

		if attachment["is_obsolete"].(int64) != int64(1) {
			file_map[attachment["file_name"].(string)] = attachment["data"].(string)
		}
	}

	// write out decoded files
	for file, base64_data := range file_map {
		data, err := base64.StdEncoding.DecodeString(base64_data)
		check(err)
		err = ioutil.WriteFile(file, data, 0644)
		check(err)
	}
}
