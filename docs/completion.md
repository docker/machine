<!--[metadata]>
+++
title = "Command-line Completion"
description = "Install Machine command-line completion"
keywords = ["machine, docker, orchestration, cli,  reference"]
[menu.main]
identifier="machine_completion"
parent="workw_machine"
weight=99
+++
<![end-metadata]-->

# Command-line Completion

Docker Machine comes with [command completion](http://en.wikipedia.org/wiki/Command-line_completion)
for the bash and zsh shell.

## Installing Command Completion

### Bash

Make sure bash completion is installed. If you use a current Linux in a non-minimal installation, bash completion should be available.
On a Mac, install with `brew install bash-completion`

Place the completion scripts in `/etc/bash_completion.d/` (`` `brew --prefix`/etc/bash_completion.d/`` on a Mac), using e.g.

    files=(docker-machine docker-machine-wrapper docker-machine-prompt)
    for f in "${files[@]}"; do
      curl -L https://raw.githubusercontent.com/docker/machine/v$(docker-machine --version | tr -ds ',' ' ' | awk 'NR==1{print $(3)}')/contrib/completion/bash/$f.bash > `brew --prefix`/etc/bash_completion.d/$f
    done

Completion will be available upon next login.


### Zsh

Place the completion scripts in your `/path/to/zsh/completion`, using e.g. `~/.zsh/completion/`

    mkdir -p ~/.zsh/completion
    curl -L https://raw.githubusercontent.com/docker/machine/v$(docker-machine --version | tr -ds ',' ' ' | awk 'NR==1{print $(3)}')/contrib/completion/zsh/_docker-machine > ~/.zsh/completion/_docker-machine

Include the directory in your `$fpath`, e.g. by adding in `~/.zshrc`

    fpath=(~/.zsh/completion $fpath)

Make sure `compinit` is loaded or do it by adding in `~/.zshrc`

    autoload -Uz compinit && compinit -i

Then reload your shell

    exec $SHELL -l


<!--[metadata]>
## Available completions

**TODO**
<![end-metadata]-->
