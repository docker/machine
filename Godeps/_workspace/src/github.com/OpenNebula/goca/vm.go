package goca

import (
	"errors"
	"strconv"
)

type VM struct {
	XMLResource
	Id   uint
	Name string
}

type VMPool struct {
	XMLResource
}

type VM_STATE int

const (
	VM_INIT VM_STATE = iota
	VM_PENDING
	VM_HOLD
	VM_ACTIVE
	VM_STOPPED
	VM_SUSPENDED
	VM_DONE
	VM_FAILED
	VM_POWEROFF
	VM_UNDEPLOYED
)

func (s VM_STATE) String() string {
	return [...]string{
		"INIT",
		"PENDING",
		"HOLD",
		"ACTIVE",
		"STOPPED",
		"SUSPENDED",
		"DONE",
		"FAILED",
		"POWEROFF",
		"UNDEPLOYED",
	}[s]
}

type LCM_STATE int

const (
	LCM_INIT LCM_STATE = iota
	PROLOG
	BOOT
	RUNNING
	MIGRATE
	SAVE_STOP
	SAVE_SUSPEND
	SAVE_MIGRATE
	PROLOG_MIGRATE
	PROLOG_RESUME
	EPILOG_STOP
	EPILOG
	SHUTDOWN
	CANCEL
	FAILURE
	CLEANUP_RESUBMIT
	UNKNOWN
	HOTPLUG
	SHUTDOWN_POWEROFF
	BOOT_UNKNOWN
	BOOT_POWEROFF
	BOOT_SUSPENDED
	BOOT_STOPPED
	CLEANUP_DELETE
	HOTPLUG_SNAPSHOT
	HOTPLUG_NIC
	HOTPLUG_SAVEAS
	HOTPLUG_SAVEAS_POWEROFF
	HOTPLUG_SAVEAS_SUSPENDED
	SHUTDOWN_UNDEPLOY
	EPILOG_UNDEPLOY
	PROLOG_UNDEPLOY
	BOOT_UNDEPLOY
	HOTPLUG_PROLOG_POWEROFF
	HOTPLUG_EPILOG_POWEROFF
	BOOT_MIGRATE
	BOOT_FAILURE
	BOOT_MIGRATE_FAILURE
	PROLOG_MIGRATE_FAILURE
	PROLOG_FAILURE
	EPILOG_FAILURE
	EPILOG_STOP_FAILURE
	EPILOG_UNDEPLOY_FAILURE
	PROLOG_MIGRATE_POWEROFF
	PROLOG_MIGRATE_POWEROFF_FAILURE
	PROLOG_MIGRATE_SUSPEND
	PROLOG_MIGRATE_SUSPEND_FAILURE
	BOOT_UNDEPLOY_FAILURE
	BOOT_STOPPED_FAILURE
	PROLOG_RESUME_FAILURE
	PROLOG_UNDEPLOY_FAILURE
	DISK_SNAPSHOT_POWEROFF
	DISK_SNAPSHOT_REVERT_POWEROFF
	DISK_SNAPSHOT_DELETE_POWEROFF
	DISK_SNAPSHOT_SUSPENDED
	DISK_SNAPSHOT_REVERT_SUSPENDED
	DISK_SNAPSHOT_DELETE_SUSPENDED
	DISK_SNAPSHOT
	DISK_SNAPSHOT_REVERT
	DISK_SNAPSHOT_DELETE
)

func (l LCM_STATE) String() string {
	return [...]string{
		"LCM_INIT",
		"PROLOG",
		"BOOT",
		"RUNNING",
		"MIGRATE",
		"SAVE_STOP",
		"SAVE_SUSPEND",
		"SAVE_MIGRATE",
		"PROLOG_MIGRATE",
		"PROLOG_RESUME",
		"EPILOG_STOP",
		"EPILOG",
		"SHUTDOWN",
		"CANCEL",
		"FAILURE",
		"CLEANUP_RESUBMIT",
		"UNKNOWN",
		"HOTPLUG",
		"SHUTDOWN_POWEROFF",
		"BOOT_UNKNOWN",
		"BOOT_POWEROFF",
		"BOOT_SUSPENDED",
		"BOOT_STOPPED",
		"CLEANUP_DELETE",
		"HOTPLUG_SNAPSHOT",
		"HOTPLUG_NIC",
		"HOTPLUG_SAVEAS",
		"HOTPLUG_SAVEAS_POWEROFF",
		"HOTPLUG_SAVEAS_SUSPENDED",
		"SHUTDOWN_UNDEPLOY",
		"EPILOG_UNDEPLOY",
		"PROLOG_UNDEPLOY",
		"BOOT_UNDEPLOY",
		"HOTPLUG_PROLOG_POWEROFF",
		"HOTPLUG_EPILOG_POWEROFF",
		"BOOT_MIGRATE",
		"BOOT_FAILURE",
		"BOOT_MIGRATE_FAILURE",
		"PROLOG_MIGRATE_FAILURE",
		"PROLOG_FAILURE",
		"EPILOG_FAILURE",
		"EPILOG_STOP_FAILURE",
		"EPILOG_UNDEPLOY_FAILURE",
		"PROLOG_MIGRATE_POWEROFF",
		"PROLOG_MIGRATE_POWEROFF_FAILURE",
		"PROLOG_MIGRATE_SUSPEND",
		"PROLOG_MIGRATE_SUSPEND_FAILURE",
		"BOOT_UNDEPLOY_FAILURE",
		"BOOT_STOPPED_FAILURE",
		"PROLOG_RESUME_FAILURE",
		"PROLOG_UNDEPLOY_FAILURE",
		"DISK_SNAPSHOT_POWEROFF",
		"DISK_SNAPSHOT_REVERT_POWEROFF",
		"DISK_SNAPSHOT_DELETE_POWEROFF",
		"DISK_SNAPSHOT_SUSPENDED",
		"DISK_SNAPSHOT_REVERT_SUSPENDED",
		"DISK_SNAPSHOT_DELETE_SUSPENDED",
		"DISK_SNAPSHOT",
		"DISK_SNAPSHOT_REVERT",
		"DISK_SNAPSHOT_DELETE",
	}[l]
}

func NewVMPool(args ...int) (*VMPool, error) {
	var who, start_id, end_id, state int

	switch len(args) {
	case 0:
		who = PoolWhoMine
		start_id = -1
		end_id = -1
		state = -1
	case 1:
		who = args[0]
		start_id = -1
		end_id = -1
		state = -1
	case 3:
		who = args[0]
		start_id = args[1]
		end_id = args[2]
		state = -1
	case 4:
		who = args[0]
		start_id = args[1]
		end_id = args[2]
		state = args[3]
	default:
		return nil, errors.New("Wrong number of arguments")
	}

	response, err := client.Call("one.vmpool.info", who, start_id, end_id, state)
	if err != nil {
		return nil, err
	}

	vmpool := &VMPool{XMLResource{body: response.Body()}}

	return vmpool, err

}

func CreateVM(template string, pending bool) (uint, error) {
	response, err := client.Call("one.vm.allocate", template, pending)
	if err != nil {
		return 0, err
	}

	return uint(response.BodyInt()), nil
}

func NewVM(id uint) *VM {
	return &VM{Id: id}
}

func NewVMFromName(name string) (*VM, error) {
	vmpool, err := NewVMPool()
	if err != nil {
		return nil, err
	}

	id, err := vmpool.GetIdFromName(name, "/VM_POOL/VM")
	if err != nil {
		return nil, err
	}

	return NewVM(id), nil
}

func (vm *VM) Info() error {
	response, err := client.Call("one.vm.info", vm.Id)
	vm.body = response.Body()
	return err
}

func (vm *VM) State() (int, int, error) {
	vm_stateString, ok := vm.XPath("/VM/STATE")
	if ok != true {
		return -1, -1, errors.New("Unable to parse VM State")
	}

	lcm_stateString, ok := vm.XPath("/VM/LCM_STATE")
	if ok != true {
		return -1, -1, errors.New("Unable to parse LCM State")
	}

	vm_state, _ := strconv.Atoi(vm_stateString)
	lcm_state, _ := strconv.Atoi(lcm_stateString)

	return vm_state, lcm_state, nil
}

func (vm *VM) StateString() (string, string, error) {
	vm_state, lcm_state, err := vm.State()
	if err != nil {
		return "", "", err
	}
	return VM_STATE(vm_state).String(), LCM_STATE(lcm_state).String(), nil
}

func (vm *VM) Action(action string) error {
	_, err := client.Call("one.vm.action", action, vm.Id)
	return err
}

func (vm *VM) Resume() error {
	return vm.Action("resume")
}
func (vm *VM) Reboot() error {
	return vm.Action("reboot")
}

func (vm *VM) PowerOff() error {
	return vm.Action("poweroff-hard")
}

func (vm *VM) PowerOffHard() error {
	return vm.Action("poweroff-hard")
}

func (vm *VM) Shutdown() error {
	return vm.Action("shutdown-hard")
}

func (vm *VM) ShutdownHard() error {
	return vm.Action("shutdown-hard")
}
