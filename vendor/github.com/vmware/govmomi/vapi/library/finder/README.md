# The content library finder
The govmomi package now includes a finder for the content library, [`github.com/vmware/govmomi/vapi/library.Finder`](https://github.com/akutz/govmomi/blob/feature/content-library/vapi/library/finder/finder.go). This finder supports searching for objects in the content library using an inventory path design similar to the [standard govmomi finder](https://github.com/vmware/govmomi/blob/master/find/finder.go) and includes support for wildcard matching courtesy of golang's [`path.Match`](https://golang.org/pkg/path/#Match). For example:

| Pattern | Result |
|---------|--------|
| `*` | Gets all of the content libraries |
| `*/` | Gets all of the content library items for all of the content libraries. |
| `/Public/` | Gets all of the content library items for the content library named `Public` |
| `/*/*/` | Gets all of the content library files for all of the content library items for all of the content libraries. |
| `Public/*/photon2*` | Gets all of the files that begin with `photon2` in all of the content library items for the content library named `Public` |

The use of a wildcard character in the search string results in a full listing of content library objects for that part of the path's server-side analogue. If `Public/Photon2*` is searched, then all of the content library items for the `Public` library are listed in order to perform the wildcard matching. However, if `Public/PhotonOS-2-GA` then the server-side API call [`library:item:find`](https://vdc-repo.vmware.com/vmwb-repository/dcr-public/1cd28284-3b72-4885-9e31-d1c6d9e26686/71ef7304-a6c9-43b3-a3cd-868b2c236c81/doc/operations/com/vmware/content/library/item.find-operation.html) is used to find an item with the name `PhotonOS-2-GA`.

## Performance
Search strings that do not use wildcard characters **should** be **more** efficient as these searches rely on server-side find calls and do not require dumping all of the objects. However, this is only true for systems with a large number of objects in the content libraries. This is due to the number of round-trips required to lookup an object with a direct path. For example:

| Search path | Roundtrips |
|------|------------|
| `Public/Photon2` | 4 |
| `Public*/Photon2*` | 2 |

The *absolute* search path takes twice as many roundtrips compared to the search path with wildcards:

### Absolute path search logic
1. Find library ID for library with name using server-side find API
2. Get library with library ID
3. Find item ID for item with name using server-side find API
4. Get item with item ID

### Wildcard search logic
1. Get all of the libraries and filter the pattern on the client-side
2. Get all of the items for the library and filter the pattern on the client-=side

### Searches at scale
While a system that has few content library objects benefits from wildcard search logic, the fact is that the above approach regarding absolute paths proves out to be **much** more efficient for systems with large numbers of content library objects.


## `govc library.ls`

### Listing all the objects in the content library
```shell
$ govc library.ls '*/*/'
/ISOs/CentOS-7-x86_64-Minimal-1804/CentOS-7-x86_64-Minimal-1804.iso
/ISOs/CoreOS Production/coreos_production_iso_image.iso
/ISOs/VMware-VCSA-all-6.7.0-8217866.iso/VMware-VCSA-all-6.7.0-8217866.iso
/ISOs/VMware-VIM-all-6.7.0-8217866.iso/VMware-VIM-all-6.7.0-8217866.iso
/ISOs/ubuntu-16.04.5-server-amd64/ubuntu-16.04.5-server-amd64.iso
/ISOs/photon-2.0-304b817/photon-2.0-304b817.iso
/OVAs/VMware-vCenter-Server-Appliance-6.7.0.10000-8217866_OVF10.ova/VMware-vCenter-Server-Appliance-6.7.0.10000-8217866_OVF10.ova
/OVAs/coreos_production_vmware_ova/coreos_production_vmware_ova.ovf
/OVAs/coreos_production_vmware_ova/coreos_production_vmware_ova-1.vmdk
/OVAs/centos_cloud_template/centos_cloud_template-2.iso
/OVAs/centos_cloud_template/centos_cloud_template.ovf
/OVAs/centos_cloud_template/centos_cloud_template-1.vmdk
/OVAs/centos_cloud_template/centos_cloud_template-3.nvram
/OVAs/photon-custom-hw13-2.0-304b817/photon-ova-disk1.vmdk
/OVAs/photon-custom-hw13-2.0-304b817/photon-ova.ovf
/OVAs/yakity-centos/yakity-centos-2.nvram
/OVAs/yakity-centos/yakity-centos-1.vmdk
/OVAs/yakity-centos/yakity-centos.ovf
/OVAs/yakity-photon/yakity-photon-1.vmdk
/OVAs/yakity-photon/yakity-photon.ovf
/OVAs/yakity-photon/yakity-photon-2.nvram
/OVAs/ubuntu-16.04-server-cloudimg-amd64/ubuntu-16.04-server-cloudimg-amd64.ovf
/OVAs/ubuntu-16.04-server-cloudimg-amd64/ubuntu-16.04-server-cloudimg-amd64-1.vmdk
/sk8-TestUploadOVA/photon2-cloud-init/photon2-cloud-init.ovf
/sk8-TestUploadOVA/photon2-cloud-init/photon2-cloud-init-1.vmdk
/Public/photon2-cloud-init/photon2-cloud-init.ovf
/Public/photon2-cloud-init/photon2-cloud-init-1.vmdk
```

## `govc library.info`

### Getting the info for all the objects in the content library
```shell
$ govc library.info '*/*/'
Name:       CentOS-7-x86_64-Minimal-1804.iso
  Path:     /ISOs/CentOS-7-x86_64-Minimal-1804/CentOS-7-x86_64-Minimal-1804.iso
  Size:     824637106200
  Version:  1
Name:       coreos_production_iso_image.iso
  Path:     /ISOs/CoreOS Production/coreos_production_iso_image.iso
  Size:     824637441624
  Version:  1
Name:       VMware-VCSA-all-6.7.0-8217866.iso
  Path:     /ISOs/VMware-VCSA-all-6.7.0-8217866.iso/VMware-VCSA-all-6.7.0-8217866.iso
  Size:     824637106368
  Version:  1
Name:       VMware-VIM-all-6.7.0-8217866.iso
  Path:     /ISOs/VMware-VIM-all-6.7.0-8217866.iso/VMware-VIM-all-6.7.0-8217866.iso
  Size:     824637106504
  Version:  1
Name:       ubuntu-16.04.5-server-amd64.iso
  Path:     /ISOs/ubuntu-16.04.5-server-amd64/ubuntu-16.04.5-server-amd64.iso
  Size:     824637441944
  Version:  1
Name:       photon-2.0-304b817.iso
  Path:     /ISOs/photon-2.0-304b817/photon-2.0-304b817.iso
  Size:     824637106672
  Version:  1
Name:       VMware-vCenter-Server-Appliance-6.7.0.10000-8217866_OVF10.ova
  Path:     /OVAs/VMware-vCenter-Server-Appliance-6.7.0.10000-8217866_OVF10.ova/VMware-vCenter-Server-Appliance-6.7.0.10000-8217866_OVF10.ova
  Size:     824637106880
  Version:  1
Name:       coreos_production_vmware_ova.ovf
  Path:     /OVAs/coreos_production_vmware_ova/coreos_production_vmware_ova.ovf
  Size:     824637107032
  Version:  1
Name:       coreos_production_vmware_ova-1.vmdk
  Path:     /OVAs/coreos_production_vmware_ova/coreos_production_vmware_ova-1.vmdk
  Size:     824637107072
  Version:  1
Name:       centos_cloud_template-2.iso
  Path:     /OVAs/centos_cloud_template/centos_cloud_template-2.iso
  Size:     824636997760
  Version:  1
Name:       centos_cloud_template.ovf
  Path:     /OVAs/centos_cloud_template/centos_cloud_template.ovf
  Size:     824636997792
  Version:  1
Name:       centos_cloud_template-1.vmdk
  Path:     /OVAs/centos_cloud_template/centos_cloud_template-1.vmdk
  Size:     824636997832
  Version:  1
Name:       centos_cloud_template-3.nvram
  Path:     /OVAs/centos_cloud_template/centos_cloud_template-3.nvram
  Size:     824636997856
  Version:  1
Name:       photon-ova-disk1.vmdk
  Path:     /OVAs/photon-custom-hw13-2.0-304b817/photon-ova-disk1.vmdk
  Size:     824637107280
  Version:  1
Name:       photon-ova.ovf
  Path:     /OVAs/photon-custom-hw13-2.0-304b817/photon-ova.ovf
  Size:     824637107328
  Version:  1
Name:       yakity-centos-2.nvram
  Path:     /OVAs/yakity-centos/yakity-centos-2.nvram
  Size:     824637253000
  Version:  2
Name:       yakity-centos-1.vmdk
  Path:     /OVAs/yakity-centos/yakity-centos-1.vmdk
  Size:     824637253024
  Version:  2
Name:       yakity-centos.ovf
  Path:     /OVAs/yakity-centos/yakity-centos.ovf
  Size:     824637253056
  Version:  2
Name:       yakity-photon-1.vmdk
  Path:     /OVAs/yakity-photon/yakity-photon-1.vmdk
  Size:     824637107504
  Version:  5
Name:       yakity-photon.ovf
  Path:     /OVAs/yakity-photon/yakity-photon.ovf
  Size:     824637107536
  Version:  5
Name:       yakity-photon-2.nvram
  Path:     /OVAs/yakity-photon/yakity-photon-2.nvram
  Size:     824637107560
  Version:  5
Name:       ubuntu-16.04-server-cloudimg-amd64.ovf
  Path:     /OVAs/ubuntu-16.04-server-cloudimg-amd64/ubuntu-16.04-server-cloudimg-amd64.ovf
  Size:     824637442112
  Version:  1
Name:       ubuntu-16.04-server-cloudimg-amd64-1.vmdk
  Path:     /OVAs/ubuntu-16.04-server-cloudimg-amd64/ubuntu-16.04-server-cloudimg-amd64-1.vmdk
  Size:     824637442136
  Version:  1
Name:       photon2-cloud-init.ovf
  Path:     /sk8-TestUploadOVA/photon2-cloud-init/photon2-cloud-init.ovf
  Size:     824636903184
  Version:  1
Name:       photon2-cloud-init-1.vmdk
  Path:     /sk8-TestUploadOVA/photon2-cloud-init/photon2-cloud-init-1.vmdk
  Size:     824636903208
  Version:  1
Name:       photon2-cloud-init.ovf
  Path:     /Public/photon2-cloud-init/photon2-cloud-init.ovf
  Size:     824637442304
  Version:  3
Name:       photon2-cloud-init-1.vmdk
  Path:     /Public/photon2-cloud-init/photon2-cloud-init-1.vmdk
  Size:     824637442328
  Version:  3
```
