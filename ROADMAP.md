# Machine Roadmap

Machine currently works really well for development and test environments. The
goal is to make it work better for provisioning and managing production
environments.

(Note: this document is a high-level overview of where we are taking Machine.
For what is coming in specific releases, see our [upcoming
milestones](https://github.com/docker/machine/milestones).)

### Boot2Docker OS Migration
In 0.3.0 we added support for migrating B2D VMs to Machine and deprecated the Boot2Docker CLI.  This will finalize the migration by moving to an official Docker supported minimal Machine operating system for local environments.

### Provisioning Stability
The creation process varies greatly from provider to provider.  Issues like network partitions, provider API hiccups, etc can cause issues.  Currently, Machine does not handle those very well.  Errors are not always reported well and the user must remove whatever has been created and start over.  This affects users using Machine directly as well as indirectly (i.e. Kitematic, Rancher, Machinery).  We should improve the provision process to handle various types of failure and better report them when they occur.

### Local Provider Improvements
In 0.3.0 we added several improvements for the VirtualBox provider however, the user experience is still a bit lacking for the other "local" providers.  Machine 0.3.0 also added support for VMware shared directories however not all features for local drivers match.  For example, VirtualBox on Linux does not map shared directories.  Hyper-V does not have shared directories at all.  We should work to get the local drivers as stable and feature complete as possible.
