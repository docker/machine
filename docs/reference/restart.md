<!--[metadata]>
+++
title = "restart"
description = "Restart a machine"
keywords = ["machine, restart, subcommand"]
[menu.main]
identifier="machine.restart"
parent="smn_machine_subcmds"
+++
<![end-metadata]-->

# restart

Restart a machine. Oftentimes this is equivalent to
`docker-machine stop; docker-machine start`. But some cloud driver try to implement a clever restart which keeps the same
ip address.

    $ docker-machine restart dev
    Waiting for VM to start...
