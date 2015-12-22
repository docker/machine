# changelog

### (unreleased)

* Add `PowerOn, InjectOvfEnv, WaitForIP` options to `import.ovf` and `import.ova` option spec file

* Add `import.spec` to produce an example json document

* Add `-options` to `import.ovf` and `import.ova`

* Add `-folder` to `import.ovf` and `import.ova`

* Add `host.add` command to add host to datacenter.

* Add `GOVC_USERNAME` and `GOVC_PASSWORD` to allow overriding username and/or
  password (used when they contain special characters that prevent them from
  being embedded in the URL).

* Retry twice on temporary network errors.

* Add `host.autostart` commands to manage VM autostart.

* Add `-persist-session` flag to control whether or not the session is
  persisted to disk (defaults to true).

### 0.1.0 (2015-03-17)

Prior to this version the changes to govc's command set were not documented.
