## Installation

1. Download the file for your OS and architecture.
2. Move the binary to your PATH.

e.g., for Mac OSX:

```console
$ curl -L https://github.com/docker/machine/releases/download/{{VERSION}}/docker-machine_darwin-amd64 >/usr/local/bin/docker-machine && \
  chmod +x /usr/local/bin/docker-machine
```

Linux:

```console
$ curl -L https://github.com/docker/machine/releases/download/{{VERSION}}/docker-machine_linux-amd64 >/usr/local/bin/docker-machine && \
  chmod +x /usr/local/bin/docker-machine
```

Windows (using [git bash](https://git-for-windows.github.io/)):

```console
$ if [[ ! -d "$HOME/bin" ]]; then mkdir -p "$HOME/bin"; fi && \
  curl -L https://github.com/docker/machine/releases/download/{{VERSION}}/docker-machine_windows-amd64.exe > "$HOME/bin/docker-machine.exe" && \
  chmod +x "$HOME/bin/docker-machine.exe"
```

## Changelog

*Edit the changelog below by hand*

{{CHANGELOG}}

## Thank You

Thank you very much to our active users and contributors. If you have filed detailed bug reports, THANK YOU!
Please continue to do so if you encounter any issues. It's your hard work that makes Docker Machine better.

The following authors contributed changes to this release:

{{CONTRIBUTORS}}

Great thanks to all of the above! We appreciate it. Keep up the great work!

## Checksums

{{CHECKSUM}}

