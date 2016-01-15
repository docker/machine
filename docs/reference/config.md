<!--[metadata]>
+++
title = "config"
description = "Show client configuration"
keywords = ["machine, config, subcommand"]
[menu.main]
parent="smn_machine_subcmds"
+++
<![end-metadata]-->

# config

Show the Docker client configuration for a machine.

    $ docker-machine config dev
    --tlsverify
    --tlscacert="/Users/ehazlett/.docker/machines/dev/ca.pem"
    --tlscert="/Users/ehazlett/.docker/machines/dev/cert.pem"
    --tlskey="/Users/ehazlett/.docker/machines/dev/key.pem"
    -H tcp://192.168.99.103:2376
