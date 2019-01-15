# Gladius Network Gateway

See the main [gladius-node](https://github.com/gladiusio/gladius-node) repository to see more.

## Docker

### Docker Hub

`docker run -it -v GLADIUS_PATH:/root/.gladius -p 7947:7947 -p 3001:3001 gladius-network-gateway:latest`

### Build from GitHub

`docker build --tag=gladius-network-gateway .`

`docker run -it -v $(pwd)/gladius:/root/.gladius -p 7947:7947 -p 3001:3001 gladius-network-gateway:latest`

* Runs the container mapping the local gladius folder in this directory to the Docker container
* Sets both used ports to the relevant machine ports

## Cross compile

To compile for other systems you need to install [xgo](https://github.com/karalabe/xgo).
This is because of the Ethereum CGO bindings.

Run `make dependencies`

Then run `xgo --targets="windows/*,darwin/*,linux/*" --dest="./build/" ./cmd/gladius-controld`
from the project root. You can change the target to be whatever system you want.

Optionally, you can install and run linting tools:

```sh
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
    alpine linux) to be able to run. This is due to the C bindings that Ethereum 
    has.

## References
