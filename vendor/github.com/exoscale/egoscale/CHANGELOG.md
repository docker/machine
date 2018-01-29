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
