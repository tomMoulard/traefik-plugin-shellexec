# traefik-plugin-shellexec

[![Build Status](https://github.com/tommoulard/traefik-plugin-shellexec/workflows/Main/badge.svg?branch=master)](https://github.com/tommoulard/traefik-plugin-shellexec/actions)

WIP

> This plugin is not currently possible
> since [Yaegi](https://github.com/traefik/yaegi) does not support [yet](https://github.com/traefik/yaegi/issues/1129) the `os/exec` package.

This is a plugin for [Traefik](https://traefik.io) which terminates a connection by executing a shell script.

## Usage

### Configuration

For now, there is no configuration required.

Here is an example of a file provider dynamic configuration (given here in YAML), where the interesting part is the `http.middlewares` section:

```yaml
# Dynamic configuration

http:
  routers:
    my-router:
      rule: host(`demo.localhost`)
      service: service-foo
      entryPoints:
        - web
      middlewares:
        - traefik-plugin-shellexec

  services:
   service-foo:
      loadBalancer:
        servers:
          - url: http://127.0.0.1:5000

  middlewares:
    traefik-plugin-shellexec:
      plugin:
        traefik-plugin-shellexec:
          enabled: true
```

### Running

Once you have a running instance of this plugin, you can call the middleware with a `POST` request.
Arguments on how to run commands on the host are provided using the body of the request.

The body can look like:
```json
{
  "command": "cat -",
  "stdin": "test"
}
```

The equivalent of this request is: `echo "test" | cat -`.

The output of this request will result in:

```json
{
  "return_code": "0",
  "stderr": "",
  "stdout": "test"
}
```

### Dev Mode

Traefik also offers a developer mode that can be used for temporary testing of plugins not hosted on GitHub.
To use a plugin in dev mode, the Traefik static configuration must define the module name (as is usual for Go packages) and a path to a [Go workspace](https://golang.org/doc/gopath_code.html#Workspaces), which can be the local GOPATH or any directory.

```yaml
# Static configuration
pilot:
  token: xxxxx

experimental:
  devPlugin:
    goPath: /plugins/go
    moduleName: github.com/tommoulard/traefik-plugin-shellexec
```

(In the above example, the `shellexec` plugin will be loaded from the path `/plugins/go/src/github.com/tommoulard/traefik-plugin-shellexec`.)

#### Dev Mode Limitations

Note that only one plugin can be tested in dev mode at a time, and when using dev mode, Traefik will shut down after 30 minutes.

## TODO

 - Support more than `POST` requests
    - Query params
    - Headers
 - add more data to response
    - Time spent doing the command (can use `time` for now)
    - pid of the process (can use `$!` for now)
    - env (can use `env` for now)
