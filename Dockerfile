# build stage
FROM golang:1.11.4 AS build-env
ADD . /src
RUN cd /src && go build -tags netgo -a -v -o gladius-network-gateway -ldflags '-w -extldflags "-static"' -i cmd/main.go

# final stage
FROM alpine
RUN apk --no-cache --update upgrade && apk --no-cache add ca-certificates
WORKDIR /app

VOLUME /root/.gladius
VOLUME /root/.gladius/wallet

COPY --from=build-env /src/gladius-network-gateway /app/
ENTRYPOINT ./gladius-network-gateway
