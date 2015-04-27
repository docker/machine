# Docker Machine

Machine lets you create Docker hosts on your computer, on cloud providers, and
inside your own data center. It creates servers, installs Docker on them, then
configures the Docker client to talk to them.

It works a bit like this:

```console
$ docker-machine create -d virtualbox dev
INFO[0000] Creating SSH key...
INFO[0000] Creating VirtualBox VM...
INFO[0007] Starting VirtualBox VM...
INFO[0007] Waiting for VM to start...
INFO[0041] "dev" has been created and is now the active machine.
INFO[0041] To point your Docker client at it, run this in your shell: eval "$(docker-machine env dev)"

$ docker-machine ls
NAME   ACTIVE   DRIVER       STATE     URL                         SWARM
dev    *        virtualbox   Running   tcp://192.168.99.127:2376

$ eval "$(docker-machine env dev)"

$ docker run busybox echo hello world
Unable to find image 'busybox:latest' locally
511136ea3c5a: Pull complete
df7546f9f060: Pull complete
ea13149945cb: Pull complete
4986bf8c1536: Pull complete
hello world

$ docker-machine create -d digitalocean --digitalocean-access-token=secret staging
INFO[0000] Creating SSH key...
INFO[0001] Creating Digital Ocean droplet...
INFO[0002] Waiting for SSH...
INFO[0070] Configuring Machine...
INFO[0109] "staging" has been created and is now the active machine.
INFO[0109] To point your Docker client at it, run this in your shell: eval "$(docker-machine env staging)"

$ docker-machine ls
NAME      ACTIVE   DRIVER         STATE     URL                          SWARM
dev                virtualbox     Running   tcp://192.168.99.127:2376
staging   *        digitalocean   Running   tcp://104.236.253.181:2376
```

## Installation and documentation

Full documentation [is available here](https://docs.docker.com/machine/).

## Contributing

Want to hack on Machine? Please start with the [Contributing Guide](https://github.com/docker/machine/blob/master/CONTRIBUTING.md).

