<!--[metadata]>
+++
title = "Docker Machine"
description = "How to install Docker Machine"
keywords = ["machine, orchestration, install, installation, docker, documentation"]
[menu.main]
parent="mn_install"
weight=3
+++
<![end-metadata]-->

# Install Docker Machine

On OS X and Windows, Machine is installed along with other Docker products when
you install the Docker Toolbox. For details on installing Docker Toolbox, see
the <a href="https://docs.docker.com/installation/mac/" target="_blank">Mac OS X
installation</a> instructions or <a
href="https://docs.docker.com/installation/windows" target="_blank">Windows
installation</a> instructions.

If you only want Docker Machine, you can install the Machine binaries (the
latest versions of which are located at
<https://github.com/docker/machine/releases/> ) directly by following the
instructions in the next section.

## Installing Machine Directly

1.  Install <a href="https://docs.docker.com/installation/"
    target="_blank">the Docker binary</a>.

2.  Download the Docker Machine binary and extract it to your PATH.

        Linux:

            $ curl -L https://github.com/docker/machine/releases/download/v0.5.5/docker-machine_linux-amd64 >/usr/local/bin/docker-machine && \
            chmod +x /usr/local/bin/docker-machine

        OSX:

            $ curl -L https://github.com/docker/machine/releases/download/v0.5.5/docker-machine_darwin-amd64 >/usr/local/bin/docker-machine && \
            chmod +x /usr/local/bin/docker-machine

        Windows (using Git Bash):

            $ if [[ ! -d "$HOME/bin" ]]; then mkdir -p "$HOME/bin"; fi && \
            curl -L https://github.com/docker/machine/releases/download/v0.5.5/docker-machine_windows-amd64.exe > "$HOME/bin/docker-machine.exe" && \
            chmod +x "$HOME/bin/docker-machine.exe"

3.  Check the installation by displaying the Machine version:

        $ docker-machine version
        docker-machine version 0.5.5, build 02c4254

## Installing bash completion scripts

The Machine repository supplies several `bash` scripts that add features such
as:

-   command completion
-   a function that displays the active machine in your shell prompt
-   a function wrapper that adds a `docker-machine use` subcommand to switch the
    active machine

To install the scripts, copy or link them into your `/etc/bash_completion.d` or
`/usr/local/etc/bash_completion.d` directory. To enable the `docker-machine` shell
prompt, add `$(__docker_machine_ps1)` to your `PS1` setting in `~/.bashrc`.

    PS1='[\u@\h \W$(__docker_machine_ps1)]\$ '

You can find additional documentation in the comments at the
[top of each script](https://github.com/docker/machine/tree/master/contrib/completion/bash).

## Where to go next

-   [Docker Machine overview](index.md)
-   [Docker Machine driver reference](drivers/index.md)
-   [Docker Machine subcommand reference](reference/index.md)
