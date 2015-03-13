# Machine Roadmap

Machine currently works really well for development and test environments. The
goal is to make it work better for provisioning and managing production
environments.

This is not a simple task -- production is inherently far more complex than
development -- but there are three things which are big steps towards that goal:
**client/server achitecture**, **swarm integration** and **flexible
provisioning**.

(Note: this document is a high-level overview of where we are taking Machine.
For what is coming in specific releases, see our [upcoming
milestones](https://github.com/docker/machine/milestones).)

## Client/server architecture

Machine currently only works with a single user. Any machines you create are
stored on your local filesystem, meaning only you can control them. This means
you can't use it in teams or for large deployments.

Machines should instead have some kind of central storage. This could be done by
having a client/server architecture, where a central server keeps track of what
machines exist and the credentials used to access them. Clients could then
connect to this server to create and manage machines.

Once we have a server, a number of other things become possible: monitoring the
health of running hosts, access control, automatically scaling them, etc.

Here are some of the pieces of work being done to make this happen:

 - [**Internal refactoring**](https://github.com/docker/machine/issues/553):
   Some of Machine's internals need to be reorganised to make an external API
   possible.
 - **Add a REST API for internal API**: The first step towards a
   server is adding a REST API for the stuff that Machine does already.

## Swarm integration

For production, Machine should be able to create and manage Swarm clusters.

This probably means there is a new high-level object in Machine called a Swarm,
which abstracted away the machines inside of it. You would be able to create and
configure a Swarm on a provider, then scale it up and down.

## Flexible provisioning

Currently, the operating systems and methods that Machine uses to provision
hosts are fixed. In production, you may want to be able to customise this
depending on your environment and requirements.

 - **Customize the provisioning**: It should be possible to do custom
   provisioning of your server. This will probably be done by [setting up
   machines with cloudinit](https://github.com/docker/machine/issues/124).
 - **Customize the operating system**: It should be possible to deploy operating
   systems which aren't boot2docker or Ubuntu.
 - **Customize the Docker Engine options**: It should be possible to pass
   options to the Docker daemon inside the machine.
