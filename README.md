# Gladius Network Gateway

See the main [gladius-node](https://github.com/gladiusio/gladius-node) repository to see more.

## Docker
Running the Network Gateway in a docker container

#### From Docker Hub

```bash
$ docker run -it -v YOUR_GLADIUS_PATH:/root/.gladius -p 7947:7947 \
    -p 3001:3001 ladiusio/network-gateway:latest
```

#### Build from GitHub

```bash
$ docker build --tag=gladiusio/network-gateway .

$ docker run -it -v $(pwd)/gladius:/root/.gladius -p 7947:7947 -p 3001:3001 \
    gladiusio/network-gateway:latest
```
* Runs the container mapping the local `./gladius` folder in this directory to the Docker containe
* Sets both used ports to the relevant machine ports

## Build from source

#### For your machine
Run `make`. The binary will be in `./build`

#### Cross compile
Check out the [gladius-node](https://github.com/gladiusio/gladius-node) repository for Dockerized cross compilation.

#### Run linter

Optionally, you can install and run linting tools:

```shell
go get gopkg.in/alecthomas/gometalinter.v2
gometalinter.v2 --install
make lint
```

## API Documentation

See the [API Documentation](./apidocs/APIDOCS.MD)

This document provides documentation for the Gladius Control Daemon to build interfaces on top of the Gladius Network with familiar REST API calls. If something needs more detail or explanation, please file an issue.

Throughout the document, you will see {{ETH_ADDRESS}}. This is a placeholder for either a node address or pool address in almost all cases.

## Known issues

-   You will need to install glibc on systems that don't have it by default (like
    alpine linux) to be able to run if the binary is dynamically linked. This is due to the C bindings that Ethereum 
    has. One way to fix this is to statically compile the Go binary with `-ldflags '-w -extldflags "-static"`
