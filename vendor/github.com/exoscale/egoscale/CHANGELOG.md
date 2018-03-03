Changelog
=========

0.9.10
------

- fix: typo made ListAll required in ListPublicIPAddresses
- fix: all bool are now *bool, respecting CS default value
- feat: (*VM).DefaultNic() to obtain the main Nic

0.9.9
-----

- fix: affinity groups virtualmachineIds attribute
- fix: uuidList is not a list of strings

0.9.8
-----

- feat: add RootDiskSize to RestoreVirtualMachine
- fix: monotonic polling using Context

0.9.7
-----

- feat: add Taggable interface to expose ResourceType
- feat: add (Create|Update|Delete|List)InstanceGroup(s)
- feat: add RegisterUserKeys
- feat: add ListResourceLimits
- feat: add ListAccounts

0.9.6
-----

- fix: update UpdateVirtualMachine userdata
- fix. Network's name/displaytext might be empty

0.9.5
-----

- fix: serialization of slice

0.9.4
-----

- fix: constants

0.9.3
-----

- change: userdata expects a string
- change: no pointer in sub-struct's

0.9.2
-----

- bug: createNetwork is a sync call
- bug: typo in listVirtualMachines' domainid
- bug: serialization of map[string], e.g. UpdateVirtualMachine
- change: IPAddress's use net.IP type
- feat: helpers VM.NicsByType, VM.NicByNetworkID, VM.NicByID
- feat: addition of CloudStack ApiErrorCode constants

0.9.1
-----

- bug: sync calls returns succes as a string rather than a bool
- change: unexport BooleanResponse types
- feat: original CloudStack error response can be obtained

0.9.0
-----

Big refactoring, addition of the documentation, compliance to golint.

0.1.0
-----

Initial library
