#!/usr/bin/env bats

load helpers

export DRIVER=kvm
export NAME="bats-$DRIVER-test"

# Note: on some systems, security policies prevent running VMs from disks in /tmp
export MACHINE_STORAGE_PATH=${HOME}/.docker/machine-bats-test-$DRIVER

mkdir -p ${MACHINE_STORAGE_PATH}

# TODO  - Check for overlapping subnets on the test rig...
function setup() {
    virsh version > /dev/null
}

function validate() {
  if [ "$status" -ne 0 ] ; then
    echo "CMD OUTPUT: $output"
    false
  else
    true
  fi
}


@test "$DRIVER: machine should not exist" {
  run machine inspect $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: VM should not exist" {
  run virsh domstate $NAME
  [ "$status" -eq 1  ]
}


# NAT tests first
@test "$DRIVER: setup NAT network" {
    NET_NAME="${NAME}-nat"
    if ! virsh net-list --all | grep ${NET_NAME} > /dev/null; then
        cat << EOF > ${MACHINE_STORAGE_PATH}/${NET_NAME}
<network>
  <name>${NET_NAME}</name>
  <forward mode='nat'/>
  <ip address='192.168.3.1' netmask='255.255.255.0'>
    <dhcp>
      <range start='192.168.3.2' end='192.168.3.254'/>
    </dhcp>
  </ip>
</network>
EOF
        virsh net-define ${MACHINE_STORAGE_PATH}/${NET_NAME}
        virsh net-start ${NET_NAME}
    fi
}

ALT_CPU=2
ALT_MEM=2048
ALT_DISK=10000
@test "$DRIVER: NAT create" {
  NET_NAME="${NAME}-nat"
  run machine -D create -d $DRIVER --kvm-vcpu ${ALT_CPU} --kvm-memory ${ALT_MEM} --kvm-disk-size ${ALT_DISK} --kvm-network ${NET_NAME} $NAME
  validate
}

@test "$DRIVER: validate non-default memory" {
  virsh dumpxml ${NAME}
  # The actual VM memory will be a little higher
  virsh dumpxml ${NAME} | grep '<memory' | grep `expr ${ALT_MEM} / 100`
}

@test "$DRIVER: validate non-default vcpu" {
  virsh dumpxml ${NAME}
  virsh dumpxml ${NAME} | grep '<vcpu' | grep ${ALT_CPU}
}

@test "$DRIVER: validate non-default disk size" {
  DISK=${MACHINE_STORAGE_PATH}/machines/${NAME}/disk.img
  ls -lh ${DISK} | grep `expr ${ALT_DISK} / 1024`
}


@test "$DRIVER: NAT ls" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"$NAME"*  ]] || (echo ${output}; false)
}

@test "$DRIVER: NAT run busybox container" {
  run docker $(machine config $NAME) run busybox echo hello world
  validate
}

@test "$DRIVER: NAT url" {
  run machine url $NAME
  validate
}

@test "$DRIVER: NAT ip" {
  run machine ip $NAME
  validate
}

@test "$DRIVER: NAT ssh" {
  run machine ssh $NAME -- ls -lah /
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "total"  ]]
}

@test "$DRIVER: NAT stop" {
  run machine stop $NAME
  validate
}

@test "$DRIVER: NAT machine should show stopped after stop" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Stopped"*  ]]
}

@test "$DRIVER: NAT start" {
  run machine start $NAME
  validate
}

@test "$DRIVER: NAT machine should show running after start" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Running"*  ]]
}

@test "$DRIVER: NAT kill" {
  run machine kill $NAME
  validate
}

@test "$DRIVER: NAT machine should show stopped after kill" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Stopped"*  ]]
}

@test "$DRIVER: NAT restart and verify running after" {
  run machine -D restart $NAME
  echo "$output"
  validate
  run machine ls
  validate
  [[ ${lines[1]} == *"Running"*  ]]
}

@test "$DRIVER: NAT remove" {
  run machine rm -f $NAME
  validate
}

@test "$DRIVER: NAT machine should not exist" {
  run machine inspect $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: NAT VM should not exist" {
  run virsh domstate $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: cleanup NAT network" {
  NET_NAME="${NAME}-nat"
  virsh net-destroy ${NET_NAME} || true
  virsh net-undefine ${NET_NAME}
}



# Routed - Unlikely to be able to see outside world
@test "$DRIVER: setup routed network" {
    NET_NAME="${NAME}-routed"
    if ! virsh net-list --all | grep ${NET_NAME} > /dev/null; then
        cat << EOF > ${MACHINE_STORAGE_PATH}/${NET_NAME}
<network>
  <name>${NET_NAME}</name>
  <forward mode='route' dev="eth0"/>
  <ip address='192.168.4.1' netmask='255.255.255.0'>
    <dhcp>
      <range start='192.168.4.2' end='192.168.4.254'/>
    </dhcp>
  </ip>
</network>
EOF
        virsh net-define ${MACHINE_STORAGE_PATH}/${NET_NAME}
        virsh net-start ${NET_NAME}
    fi

}

@test "$DRIVER: ROUTED create" {
  NET_NAME="${NAME}-routed"
  run machine -D create -d $DRIVER --kvm-network ${NET_NAME} $NAME
  validate
}

@test "$DRIVER: ROUTED ls" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"$NAME"*  ]]
}

@test "$DRIVER: ROUTED kill" {
  run machine kill $NAME
  validate
}

@test "$DRIVER: ROUTED cleanup routed network" {
  NET_NAME="${NAME}-routed"
  virsh net-destroy ${NET_NAME} || true
  virsh net-undefine ${NET_NAME}
}

@test "$DRIVER: ROUTED remove" {
  run machine rm -f $NAME
  validate
}

@test "$DRIVER: ROUTED machine should not exist" {
  run machine inspect $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: ROUTED VM should not exist" {
  run virsh domstate $NAME
  [ "$status" -eq 1  ]
}



# Isolated - Wont have external visibility
@test "$DRIVER: setup isolated network" {
    NET_NAME="${NAME}-isolated"
    if ! virsh net-list --all | grep ${NET_NAME} > /dev/null; then
        cat << EOF > ${MACHINE_STORAGE_PATH}/${NET_NAME}
<network>
  <name>${NET_NAME}</name>
  <ip address="192.168.5.1" netmask="255.255.255.0">
    <dhcp>
      <range start="192.168.5.2" end="192.168.5.254" />
    </dhcp>
  </ip>
</network>
EOF
        virsh net-define ${MACHINE_STORAGE_PATH}/${NET_NAME}
        virsh net-start ${NET_NAME}
    fi
}

@test "$DRIVER: ISOLATED create" {
  NET_NAME="${NAME}-isolated"
  run machine -D create -d $DRIVER --kvm-network ${NET_NAME} $NAME
  validate
}

@test "$DRIVER: ISOLATED ls" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"$NAME"*  ]]
}

@test "$DRIVER: ISOLATED kill" {
  run machine kill $NAME
  validate
}

@test "$DRIVER: ISOLATED cleanup network" {
  NET_NAME="${NAME}-isolated"
  virsh net-destroy ${NET_NAME} || true
  virsh net-undefine ${NET_NAME}
}

@test "$DRIVER: ISOLATED remove" {
  run machine rm -f $NAME
  validate
}

@test "$DRIVER: ISOLATED machine should not exist" {
  run machine inspect $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: ISOLATED VM should not exist" {
  run virsh domstate $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: Final cleanup" {
  run rm -rf $MACHINE_STORAGE_PATH
  validate
}
