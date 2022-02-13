For visioning we use `X.Y.Z-gitlab.G` pattern, where:

- `X.Y.Z` is directly taken from the upstream version. For now we add
  all our changes on top of `0.16.2` which is the last stable tag of
  Docker Machine.
- `G` is the incremental number increased for each of our releases.

### Release 0.16.2-gitlab.{{G}}

- [ ] `git checkout main && git pull`
- [ ] Increment the value of {{G}} by checking the value defined in https://gitlab.com/gitlab-org/ci-cd/docker-machine/-/blob/main/version/version.go#L9.
- [ ] Update [version/version.go](https://gitlab.com/gitlab-org/ci-cd/docker-machine/-/blob/93376765782dc284064f3e4ccf87d8500e983888/version/version.go#L9) to `0.16.2-gitlab.{{G}}`.
- [ ] Add file `git add version/version.go`
- [ ] Commit `git commit -m "Bump version to 0.16.2-gitlab.{{G}}"`
- [ ] Create git tag `git tag -s v0.16.2-gitlab.{{G}} -m "Version 0.16.2-gitlab.{{G}}"`
- [ ] Push tag `git push origin v0.16.2-gitlab.{{G}}`
- [ ] Push to main `git push origin main`
