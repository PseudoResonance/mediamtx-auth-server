# MediaMTX HTTP Auth Server

A simple [HTTP auth server](https://mediamtx.org/docs/usage/authentication) for MediaMTX using a PostgreSQL backend.

Designed especially for creating temporary access URLs tied to a persistent publishing URL.

## Endpoints

|Path|Description|
|--|--|
|`/auth`|Authentication endpoint to provide to MediaMTX|
|`/connection`|Connection opened/closed endpoint to provide to MediaMTX|
|`/forward`|Forward auth endpoint for thumbnail server|
|`/healthz`|Healthcheck endpoint|

## Command Line Arguments

|Flag|Default|Description|
|--|--|--|
|`-c`|`config.yaml`|Path to the config file/directory|

## Environment Variables

|Variable|Description|
|--|--|
|`CONFIG_PATH`|Path to the config file/directory|
|`BIND_ADDRESS`|Local address to bind server to|
|`BIND_PORT`|Port to bind server to|
|`DB_HOSTNAME`|Database hostname|
|`DB_PORT`|Database port|
|`DB_DATABASE`|Database name|
|`DB_USERNAME`|Database username|
|`DB_PASSWORD`|Database password|

## Configuration

If the config file is unavailable at start, a default file will be generated.

If the supplied config path is a directory instead of a file, all YAML files within the directory will be loaded and merged in alphabetical order. Later files will overwrite the values of prior files.

Descriptions of the config entries are in [the default config](config.default.yaml).

## MediaMTX Configuration

Something similar to the following should be used in the MediaMTX config.

Please note that currently only the FFmpeg image contains wget.

```yaml
authMethod: http
authHTTPAddress: http://localhost:8080/auth
runOnConnect: wget -qO /dev/null "http://localhost:8080/connection?action=connect&type=$MTX_CONN_TYPE&id=$MTX_CONN_ID"
runOnDisconnect: wget -qO /dev/null "http://localhost:8080/connection?action=disconnect&type=$MTX_CONN_TYPE&id=$MTX_CONN_ID"
```

## Forward Auth Thumbnail Server Configuration

To protect thumbnails behind auth as well, the server hosting/proxying the thumbnails can use forward auth.

The original URI and connecting IP should be forwarded as well, and the corresponding header names must be configured.

```yaml
forwardAuth:
    # Header which contains the original request URI
    uriHeader: "X-Forwarded-Uri"
    # Header containing the original request IP
    ipHeader: "X-Forwarded-For"
    # Whatever prefix (or no prefix) prepends each request path
    basePath: "/thumbnails"
```

Caddy forward auth example:

```caddyfile
:443 {
	handle /thumbnails/* {
		forward_auth * localhost:8080 {
			uri /forward
		}

		root * /usr/share/caddy
		file_server *
	}

	handle {
		respond * "Forbidden" 403 {
			close
		}
	}
}
```

## License

Licensed under the Apache License, Version 2.0.
