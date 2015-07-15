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

```
Usage: docker-machine ls [OPTIONS] [arg...]

List machines

Options:

   --quiet, -q					Enable quiet mode
   --filter [--filter option --filter option]	Filter output based on conditions provided
```

## Filtering

The filtering flag (`-f` or `--filter)` format is a `key=value` pair. If there is more
than one filter, then pass multiple flags (e.g. `--filter "foo=bar" --filter "bif=baz"`)

The currently supported filters are:

* driver (driver name)
* swarm (swarm master's name)
* state (`Running|Paused|Saved|Stopped|Stopping|Starting|Error`)

## Examples

```
$ docker-machine ls
NAME   ACTIVE   DRIVER       STATE     URL
dev             virtualbox   Stopped
foo0            virtualbox   Running   tcp://192.168.99.105:2376
foo1            virtualbox   Running   tcp://192.168.99.106:2376
foo2   *        virtualbox   Running   tcp://192.168.99.107:2376
```

```
$ docker-machine ls --filter driver=virtualbox --filter state=Stopped
NAME   ACTIVE   DRIVER       STATE     URL   SWARM
dev             virtualbox   Stopped
```
