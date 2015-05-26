# Machine Roadmap

(Note: this document is a high-level overview of where we are taking Machine.
For what is coming in specific releases, see our [upcoming
milestones](https://github.com/docker/machine/milestones).)

Docker Machine's focus in terms of use case for the time being is on development
and test environments.  It is intended to be usable for production in the
future, but in order to do so the stability and usability of the experience must
be improved further for the use cases of dev and test first (the fruit hangs
lower here).

Additionally, the project should provide bulding blocks for other tools which
need the primitives that Docker Machine has, but does not neccessarily expose in
a robust way.

Following are the outlined high-level goals for the 0.4.0 release intended to
meet these objectives.

## Decouple `libmachine` API and `docker-machine` CLI

This objective would create Go bindings for Docker Machine which are usable by
embedding in another Go program directly.  This feature is desirable by many in
the community, e.g. Rancher, who currently want to rely on Docker Machine as a
building block but are currently forced to use the provided binary and shell out
to it.

A proposal for doing this is in motion
[here](https://github.com/docker/machine/issues/1211).  This also feeds into the
next goal (Declarative Configuration) quite well, since having a decoupled API
and CLI allows us to implement such a configuration file more easily (we can
directly translate the "external" representation of desired cluster state into
an "internal" API implementation without additional middleware layers).

## Declarative Configuration

While the Machine CLI in its current state is excellent for playing around and
creating quick prototypes, eventually it will be neccessary for a variety of
reasons to codify what Machine is doing or is intended to do in file format.

Some reasons include:

1. Convenience of being able to `docker-machine up` / `docker-machine apply`
(much like how `docker-compose` works today) and have a machine or group of
machines with the proper settings bootstrap from one command. This would help
with automatic creation of swarm clusters and development environments.
2. Auditability.  With such a file, the desired configuration of the system
could be checked into version control, reviewed by operators and administrators,
and bring us one step closer to being able to safely create, destroy, and patch
infrastructure and/or development environments.  Right now everything is
"ad-hoc", which is fraught with peril for users wanting to closely track the
state of their system.
3. Wanting to change configurable options about the setup "in flight".  For
instance, the user may have originally created the Docker daemon using an option
such as `--dns=8.8.8.8`, which they now wish to change or get rid of without
having to tear down the existing machine and create a new one.  This seems like
more of a development-centric feature than one useful for production, so a
discussion on how it might fit in (or not fit in) with immutable infrastructure
patterns seems apropos.

The [declarative configuration](https://github.com/docker/machine/issues/773)
proposed is intended to accomplish this goal.  It should be noted that the scope
of this configuration will be limited to declaring the desired machines, Docker
configuration, and Swarm configuration.  There are no plans to expand it to
support anything else, and one of the big demands from the community to to
preserve the existing functionality in Machine and not turn it into a
general-purpose configuration management tool.

For that reason, the scope will be deliberately limited to the aforementioned
things (Docker, Swarm, and provider-specific machine parameters).

## Run `docker-compose` file(s) on created machines

There are many situations in which it would be useful to execute a
`docker-compose.yml` file, or a series of `docker-compose.yml` files, on a
created machine _after_ Docker is installed and configured but _before_ Swarm
has been bootstrapped.  Some examples:

1. Additional provisioning steps at the discretion of the user.  They may want
to use a tool such as Ansible to install and configure firewalls on the host,
for example.  This would provide them with a building block to do so.
2. Wanting to boot a Zookeeper, Consul, or etcd container for Swarm discovery
before Machine's Swarm configuration and booting takes place.
3. Wanting to run a pre-defined "stack" of containers, so that e.g. a web
application users are wanting to develop using Docker, a Jenkins server, etc.
can be booted with just one `docker-machine up` command.

This would add support for doing so.

An additional sample of what such a file might look like, in TOML for kicks, is
available here: https://gist.github.com/nathanleclaire/769cbeb1b45744524531

## Support for creating multiple instances at once

Machine cannot currently create multiple instances at once as proposed in
https://github.com/docker/machine/pull/715 and elsewhere.  This is a vital user
experience feature: it will make the creation of multi-machine setups easier and
more reliable (currently users have to hack around using the shell and other
tricks).  It will allow us to bootstrap Swarm clusters faster and more easily,
discover tricky Machine issues (since we can discover issues which may require
many creates more quickly) and consequently get a lot more valuable testing and
feedback in for Swarm and Machine.

## __TENTATIVE__: Primitive "Rivet" / Extension Support

The demand for additional drivers and features in Docker Machine is quite high.
The Docker Machine team wants to provide a very consistent and very stable core
experience, but still provide users the freedom to experiment and develop using
the same building blocks for their use cases.  We are not equipped to support
every use case, but we _can_ provide the primitives so that others can extend
Docker Machine.

To that extent, we want to introduce something like
[Rivet](https://github.com/ehazlett/rivet) ASAP.  It may not be exactly how
Rivet is today in its final form, but the fundamental cornerstones would be:

1. An extensible "plugin" interface to allow for whichever drivers people want
to use, and also to "hook into" Machine's native events (`start`, `stop`,
`create` etc.) to extend their functionality.  For instance, to update
`/etc/hosts` whenever a host is created, removed, or changes IP addresses.
2. Doing this all in containers to solve the problem of actually distributing
and using these plugins.

You can imagine something along the lines of:

```console
$ docker-machine create -d ehazlett/kvm --kvm-memory 2048 --kvm-image ... my-kvm-box
```

Where `ehazlett/kvm` is the name of an image on the Docker Hub that is
configured to plug into Rivet the correct way.

To that extent, it has been discussed to potentially have only local providers
(VirtualBox, VMware Fusion, Hyper-V, etc.) directly inside of Machine to solve
the "boostrap" problem (i.e. how to run containers if you have no Docker yet on
Windows or OSX), and run the rest of the drivers (even the core supported ones)
inside of containers for dogfooding purposes.

The purpose of this tool would be to provide a playground for prototyping and
innovating on top of the Machine core, in addition to providing the basis for
users to add features which we are not keen on supporting directly due to their
complex nature or limited utility to all but a certain subset of users.

## __TENTATIVE__: Uniform Resource Model

Right now there is no uniform model for the resources Machine creates (and,
indeed, sometimes it is obligated to created resources such as security groups
to ensure that the correct ports are open for connection, SSH keys are created
properly, and so on).  This causes all sorts of issues due to inconsistency and
confusion of responsibilty: currently the drivers are expected to clean up all
of the resources they create, but do not always do so reliably, etc.

The Uniform Resource Model would start small by defining an interface such as
follows:

```
type Resource interface {
        Create() (error)
        Read() (bool, error) // Can now distinguish between an "error read" and the resource not existing - no direct way of tackling this today
        Update() (error) // Would be used to change parameters such as instance size when needed - should be able to be called idempotently
        Delete() (error)
        Parents() ([]Resource, error)
}
```

Instead of invoking `Create`, `Delete`, etc. directly on the instances, they
would instead be part of a uniform resource creation wrapper so that other,
dependent resources (a limited subset, since Machine should only account for
what it absolutely has to) can be created reliably and verified to exist as part
of creation.  Without such a unifying model, Machine will continue to be plagued
by "out-of-state" bugs which it has no way of recovering, except by
brute-forcing checks for the relevant resources in every location.

This is also intended to make it possible to recover from error conditions, e.g.
something upstream failed during provision, but we still want to be able to
`Create` our `ProvisionedResource`, which is something that would theoretically
be possible with this model, since Machine could detect the current state of the
system and see where on the "map" (declared configuration) to re-attempt from.
Because we have a very limited sub-set of things we wish Machine to account for
with respect to provisioning, this seems viable.

The design of such a system will have to be considered carefully, and ideally
rolled out slowly, so that it can be proven to be stable before moving all of
the resources over to use it.  We would start by moving all of the `Driver`
implementations to fulfill this interface, then moving all of the associated
resources (SSH keys, security groups, etc.) over to the new model as it becomes
more mature.  As time goes on, properties which can be modified by `Update` can
be added, until they are supported Machine will simply error out with a message
saying that updating that parameters is not supported.

This goal is the most tentative of all listed here, but the problems it
addresses are something which must be tackled in some fashion eventually if
Machine is to become a reliable, and error-resilient, tool.
