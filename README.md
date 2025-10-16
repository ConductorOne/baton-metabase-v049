![Baton Logo](./baton-logo.png)

# `baton-metabase-v049` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-metabase-v049.svg)](https://pkg.go.dev/github.com/conductorone/baton-metabase-v049) ![main ci](https://github.com/conductorone/baton-metabase-v049/actions/workflows/main.yaml/badge.svg)

`baton-metabase-v049` is a connector for built using the [Baton SDK](https://github.com/conductorone/baton-sdk).

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

# Prerequisites
For the connector to work properly, install the free open-source version of Metabase v0.49.x.
Versions lower or upper than v0.49 are not supported.

* Official releases: https://github.com/metabase/metabase/releases
* Docker Hub images: https://hub.docker.com/r/metabase/metabase/tags?name=0.49

  Example:
  ```
  docker run -d -p 3000:3000 \
  --name metabase \
  metabase/metabase:v0.49.0
  ```

The previous commands starts Metabase and exposes it on port 3000 of the server.
The connector requires the --metabase-base-url parameter, which should be set to the URL where this Metabase instance is accessible (e.g., https://metabase.customer.com for production).
For example:
If Metabase is running on a server with domain metabase.customer.com and port 443 (HTTPS), the base URL would be:
* --metabase-base-url https://metabase.customer.com
  For the previous case of docker commands, the base URL would be:
* --metabase-base-url http://localhost:3000

## Connector credentials
1. What credentials or information are needed to set up the connector? (For example, API key, client ID and secret, domain, etc.)

   Requires a base URL and an API Key. Args: --metabase-base-url, --metabase-api-key

   The required URL was defined in the connector requirements instructions
   To obtain the API key follow the next steps:
    1. In your Metabase address where the open source version was launched, click on the gear icon in the upper right section and click on admin settings:
       ![1-049.png](1-049.png)

    2. Click on authentication and API keys, then click on manage:
       ![2-049.png](2-049.png)

    3. Click on create API key:
       ![3-049.png](3-049.png)

    4. Fill in the required fields:
        * Key name: Enter any descriptive name (e.g. baton-connector).
        * Group: Select the administrators group as this will allow you to synchronize all connector resources with the API key.
        * Create the API key.
          ![4-049.png](4-049.png)

    5. Save your API key as you will not be able to view it again:
       ![5-049.png](5-049.png)

# Getting Started

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-metabase-v049
baton-metabase-v049
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_DOMAIN_URL=domain_url -e BATON_API_KEY=apiKey -e BATON_USERNAME=username ghcr.io/conductorone/baton-metabase-v049:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-metabase-v049/cmd/baton-metabase-v049@main

baton-metabase-v049

baton resources
```

# Data Model

`baton-metabase-v049` will pull down information about the following resources:
- Users

`baton-metabase-v049` does not specify supporting account provisioning or entitlement provisioning.

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually
building spreadsheets. We welcome contributions, and ideas, no matter how
small&mdash;our goal is to make identity and permissions sprawl less painful for
everyone. If you have questions, problems, or ideas: Please open a GitHub Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-metabase-v049` Command Line Usage

```
baton-metabase-v049

Usage:
  baton-metabase-v049 [flags]
  baton-metabase-v049 [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
      --metabase-base-url string     The base URL 2of the Metabase instance. e.g., https://metabase.customer.com ($METABASE_BASE_URL)
      --metabase-api-key string      API key generated in Metabase for the connector ($METABASE_API_KEY)
      --client-id string             The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string         The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
  -f, --file string                  The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                         help for baton-metabase-v049
      --log-format string            The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string             The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
  -p, --provisioning                 If this connector supports provisioning, this must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --ticketing                    This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                      version for baton-metabase-v049

Use "baton-metabase-v049 [command] --help" for more information about a command.
```
