Roadmap
=======

Machine 1.0
-----------

This will be a stable version of the current design with additional drivers and complete documentation.

You can follow progress towards this release with the [GitHub milestone](https://github.com/docker/machine/milestones/1.0).

Future
------

There are two main areas for future development:

 - **Machine server:** Machine currently relies on storing configuration locally on the computer that the command-line client is run which makes it difficult to use in teams and for large deployments.
   
   Machines should instead be managed by a central server with a REST API. The command-line client would be a client for this server. To keep the current behaviour, and to manage local VMs, the server could run in an embedded mode inside the client.

 - **Swarm integration:** Machine should be able to create and manage [Swarm](https://github.com/docker/swarm) clusters. Perhaps it's even the default. Imagine this:

        $ docker-machine create -d digitalocean production
        $ docker-machine scale production=100

