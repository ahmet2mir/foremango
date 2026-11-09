# Examples

Runnable reference code for `foremango`, one self-contained `package main`
per directory. The server and credentials come from the environment;
resource IDs in the code are placeholders, so use IDs that exist on your
Foreman to run them for real.

```sh
FOREMAN_URL=https://foreman.example.com \
FOREMAN_USERNAME=admin FOREMAN_PASSWORD=... \
go run ./examples/host_with_parameters
```

* [`host_with_parameters`](host_with_parameters) - create a host and
  associate parameters with it, both inline at creation time
  (`ForemanHost.HostParametersAttributes`) and afterwards as a separate
  call (`CreateParameter`), then list them (`QueryParameter`).
