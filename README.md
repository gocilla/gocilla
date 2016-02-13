# Gocilla

Continuous integration server, fully implemented in [Go](https://golang.org), focused on the whole process of [continuous delivery](https://en.wikipedia.org/wiki/Continuous_delivery). It is inspired by [travis](https://travis-ci.org), [jenkins](https://jenkins-ci.org/), and other tools.

## Architecture

![Architecture](docs/architecture.png)

## Installation

```
go get github.com/gocilla/gocilla
```

## Configuration

Gocilla is set up with a JSON configuration file. There is a default configuration at `${GOPATH}/src/github.com/gocilla/gocilla/config.json` that should be cloned to prepare custom settings.

Gocilla uses the environment variable `CONFIG_PATH` to locate the configuration file. If this variable is unset, then it is located at `${PWD}/config.json`.

## Start

### Initial requirements

#### mongoDB

Some information is stored in the database like the builds, the hooks, and access tokens to access GitHub API.

#### Developer application at GitHub

Register a new developer application at [GitHub](https://github.com/settings/applications/new) to enable Gocilla to interact with GitHub API. The authorization callback URL should be:

```http://{GOCILLA_HOST}:{GOCILLA_PORT}/login/callback```

This URL might be `http://localhost:3000/login/callback` in a development configuration. Note that this URL is not accessed by GitHub but by the user's browser after a HTTP redirection during the OAuth process.

The **Client ID** and **Client Secret** of the application correspond to configuration properties **oauth2.strategy.clientID** and **clientSecret** respectively.

### Launching Gocilla

```bash
export CONFIG_PATH=${HOME}/.gocilla/config.json
gocilla
```

**NOTE**: It is expected that `${GOPATH}/bin` is included in the `PATH`. It is also expected that the custom configuration for Gocilla is available at `${HOME}/.gocilla/config.json`.

You can access to Gocilla site with your web browser at [http://localhost:3000](http://localhost:3000).

## License

Copyright 2016 [Telefónica Investigación y Desarrollo, S.A.U](http://www.tid.es)

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
