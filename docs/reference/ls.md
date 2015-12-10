<!--[metadata]>
+++
title = "ls"
description = "List machines"
keywords = ["machine, ls, subcommand"]
[menu.main]
parent="smn_machine_subcmds"
+++
<![end-metadata]-->

# ls

    Usage: docker-machine ls [OPTIONS] [arg...]

    List machines

    Options:

       --quiet, -q					                Enable quiet mode
       --filter [--filter option --filter option]	Filter output based on conditions provided
       --timeout, -t				                Timeout in seconds, default to 10s

## Timeout

The `ls` command tries to reach each host in parallel. If a given host does not answer in less than 10 seconds, the `ls` command
will state that this host is in `timeout`. In some circumstances (poor connection, high load or while troubleshooting) you may want to
increase or decrease this value. You can use the `-t` flag for this purpose with a numerical value in seconds.

### Example

    $ docker-machine ls -t 12
    NAME      ACTIVE   DRIVER       STATE     URL                         SWARM   DOCKER   ERRORS
    default   -        virtualbox   Running   tcp://192.168.99.100:2376           v1.9.0

## Filtering

The filtering flag (`-f` or `--filter)` format is a `key=value` pair. If there is more
than one filter, then pass multiple flags (e.g. `--filter "foo=bar" --filter "bif=baz"`)

The currently supported filters are:

-   driver (driver name)
-   swarm  (swarm master's name)
-   state  (`Running|Paused|Saved|Stopped|Stopping|Starting|Error`)
-   name   (Machine name returned by driver, supports [golang style](https://github.com/google/re2/wiki/Syntax) regular expressions)
-   label  (Machine created with `--engine-label` option, can be filtered with `label=<key>[=<value>]`)

## Examples

    $ docker-machine ls
    NAME   ACTIVE   DRIVER       STATE     URL
    dev    -        virtualbox   Stopped
    foo0   -        virtualbox   Running   tcp://192.168.99.105:2376
    foo1   -        virtualbox   Running   tcp://192.168.99.106:2376
    foo2   *        virtualbox   Running   tcp://192.168.99.107:2376

    $ docker-machine ls --filter driver=virtualbox --filter state=Stopped
    NAME   ACTIVE   DRIVER       STATE     URL   SWARM
    dev    -        virtualbox   Stopped

    $ docker-machine ls --filter label=com.class.app=foo1 --filter label=com.class.app=foo2
    NAME   ACTIVE   DRIVER       STATE     URL
    foo1   -        virtualbox   Running   tcp://192.168.99.105:2376
    foo2   *        virtualbox   Running   tcp://192.168.99.107:2376
