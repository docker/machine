/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmwarefusion

const vmx = `
.encoding = "UTF-8"
config.version = "8"
displayName = "{{.MachineName}}"
ethernet0.addressType = "generated"
ethernet0.connectionType = "nat"
ethernet0.linkStatePropagation.enable = "TRUE"
ethernet0.present = "TRUE"
ethernet0.virtualDev = "e1000"
ethernet0.wakeOnPcktRcv = "FALSE"
floppy0.present = "FALSE"
guestOS = "other26xlinux-64"
hpet0.present = "TRUE"
ide1:0.deviceType = "cdrom-image"
ide1:0.fileName = "{{.ISO}}"
ide1:0.present = "TRUE"
mem.hotadd = "TRUE"
memsize = "{{.Memory}}"
powerType.powerOff = "hard"
powerType.powerOn = "hard"
powerType.reset = "hard"
powerType.suspend = "hard"
scsi0.present = "TRUE"
scsi0.virtualDev = "lsilogic"
scsi0:0.fileName = "{{.MachineName}}.vmdk"
scsi0:0.present = "TRUE"
virtualHW.productCompatibility = "hosted"
virtualHW.version = "10"
msg.autoanswer = "TRUE"
uuid.action = "create"
`
