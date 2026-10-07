# Examples

Runnable reference code for `foremango`, one self-contained `package main`
per directory. Every value in them - server URL, credentials, resource
IDs - is a placeholder; point them at a real Foreman instance (with IDs
that actually exist there) to run for real.

```sh
go run ./examples/host_with_parameters
```

* [`host_with_parameters`](host_with_parameters) - create a host and
  associate parameters with it, both inline at creation time
  (`ForemanHost.HostParameters`) and afterwards as a separate call
  (`CreateParameter`).
