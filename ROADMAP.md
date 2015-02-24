# Machine Roadmap

These are the ideas for the roadmap.  They are not all slated for the next
release.  You can follow progress towards the current release with the 
[GitHub milestone](https://github.com/docker/machine/milestones/0.2.0).

## Internal API
 - [ ] Refactor Driver architecture to be more concise for providers
 - [ ] Integration Test suite for drivers
 - [ ] Refactor SSH (`ssh` -> `crypto/ssh`)
 - [ ] Logging Unification (internal and provider logging)

## Provisioning
 - [ ] Cloudinit as standard provisioning method
 - [ ] Customization of the Docker Engine options
 - [ ] Alternate to b2d for local providers

## Swarm
 - [ ] Full Swarm Integration
   - [ ] Scale cluster (create/start/stop/remove nodes)
 - [ ] Swarm configuration (labels, strategies)

## Machine Server
 - [ ] Provide REST API for operations (?)
 - [ ] Certificate management (?)
   - [ ] Issue client certificates (activate/deactivate)
   - [ ] Certificate rotation
 - [ ] SSH key management (?)
   - [ ] Manage keys (activate/deactivate)
   - [ ] Key rotation
 - [ ] Metrics (?)

