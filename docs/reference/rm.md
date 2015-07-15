<!--[metadata]>
+++
title = "rm"
description = "Remove a machine."
keywords = ["machine, rm, subcommand"]
[menu.main]
parent="smn_machine_subcmds"
+++
<![end-metadata]-->

# rm

Remove a machine. This will remove the local reference as well as delete it
on the cloud provider or virtualization management platform.

```
$ docker-machine ls
NAME   ACTIVE   DRIVER       STATE     URL
foo0            virtualbox   Running   tcp://192.168.99.105:2376
foo1            virtualbox   Running   tcp://192.168.99.106:2376
$ docker-machine rm foo1
$ docker-machine ls
NAME   ACTIVE   DRIVER       STATE     URL
foo0            virtualbox   Running   tcp://192.168.99.105:2376
```