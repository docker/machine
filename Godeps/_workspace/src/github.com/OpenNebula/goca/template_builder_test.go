package goca

import (
	"fmt"
)

func Example() {
	template := NewTemplateBuilder()

	template.AddValue("cpu", 1)
	template.AddValue("memory", "64")
	vector := template.NewVector("disk")
	vector.AddValue("image_id", "119")
	vector.AddValue("dev_prefix", "vd")
	vector = template.NewVector("nic")
	vector.AddValue("network_id", "3")
	vector.AddValue("model", "virtio")
	template.AddValue("vcpu", "2")

	fmt.Println(template)

	// Output:
	// CPU="1"
	// MEMORY="64"
	// DISK=[
	//     IMAGE_ID="119",
	//     DEV_PREFIX="vd" ]
	// NIC=[
	//     NETWORK_ID="3",
	//     MODEL="virtio" ]
	// VCPU="2"
}
