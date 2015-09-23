#!/usr/bin/env bats

load helpers

export DRIVER=virtualbox
export NAME="bats-$DRIVER-test"
export MACHINE_STORAGE_PATH=/tmp/machine-bats-test-$DRIVER

@test "$DRIVER: create" {
  run machine --tls-san foo.bar.tld --tls-san 10.42.42.42 create -d $DRIVER $NAME
}

@test "$DRIVER: verify that server cert contains the extra SANs" {
    machine ssh $NAME -- openssl x509 -in /var/lib/boot2docker/server.pem -text | grep 'DNS:foo.bar.tld'
    machine ssh $NAME -- openssl x509 -in /var/lib/boot2docker/server.pem -text | grep 'IP Address:10.42.42.42'
}

@test "$DRIVER: verify that server cert SANs are still there after 'regenerate-certs'" {
    machine regenerate-certs -f $NAME
    machine ssh $NAME -- openssl x509 -in /var/lib/boot2docker/server.pem -text | grep 'DNS:foo.bar.tld'
    machine ssh $NAME -- openssl x509 -in /var/lib/boot2docker/server.pem -text | grep 'IP Address:10.42.42.42'
}

@test "cleanup" {
  machine rm $NAME
}
