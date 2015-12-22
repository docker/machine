#!/bin/sh

useradd vagrant -m -s /bin/bash
groupmod -A vagrant wheel

echo -e "vagrant ALL=(ALL) NOPASSWD: ALL\n" >> /etc/sudoers

mkdir ~vagrant/.ssh
wget --no-check-certificate \
     https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant.pub \
     -O ~vagrant/.ssh/authorized_keys
chown -R vagrant ~vagrant/.ssh
chmod -R go-rwsx ~vagrant/.ssh

perl -pi -e 's/^#UseDNS yes/UseDNS no/' /etc/ssh/sshd_config
perl -pi -e 's/^AllowTcpForwarding no//' /etc/ssh/sshd_config
perl -pi -e 's/^PermitTunnel no//' /etc/ssh/sshd_config
perl -pi -e 's/^MaxSessions \d+//' /etc/ssh/sshd_config

# disable password expiration
for uid in root vagrant
do
  chage -I -1 -E -1 -m 0 -M -1 $uid
done
