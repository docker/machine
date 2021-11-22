---
stage: Verify
group: Runner
info: To determine the technical writer assigned to the Stage/Group associated with this page, see https://about.gitlab.com/handbook/engineering/ux/technical-writing/#assignments
---

# Install Docker Machine

1. Download the [appropriate `docker-machine` binary](https://gitlab.com/gitlab-org/ci-cd/docker-machine/-/releases).
   Copy the binary to a location accessible to `PATH` and make it
   executable. For example, to download and install `v0.16.2-gitlab.11`:

    ```shell
    curl -O "https://gitlab-docker-machine-downloads.s3.amazonaws.com/v0.16.2-gitlab.11/docker-machine-Linux-x86_64"
    cp docker-machine-Linux-x86_64 /usr/local/bin/docker-machine
    chmod +x /usr/local/bin/docker-machine
    ```