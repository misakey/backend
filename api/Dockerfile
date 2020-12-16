FROM golang:1.14 as builder

WORKDIR /go/src/app
COPY . .

RUN GO111MODULE=on go mod vendor

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -mod=vendor \
    -installsuffix 'static' \
    -o /bin/api .

ENTRYPOINT ["/bin/api"]

FROM alpine:3.11

# install tls tooling
RUN until apk --no-cache add ca-certificates~=20191127-r2; do sleep 1; done; rm -rf /var/cache/apk/*

# copy mailing template files
COPY ./src/templates /etc/templates

# add configuration file as configuration dev file to be integrated directly in our images
# in production, the file must be served via a mounted volume
COPY ./config/api.toml /etc/api-config.dev.toml

# get binary from build image
COPY --from=builder /bin/api /bin/api

# set the version env variable used in the /version endpoint
ARG VERSION=unset
ENV VERSION $VERSION

# add all migrations to corresponding modules dir
COPY ./src/box/db/migration /etc/box-migration
COPY ./src/sso/db/migration /etc/sso-migration
ENV MIGRATION_DIR_BOX /etc/box-migration
ENV MIGRATION_DIR_SSO /etc/sso-migration

EXPOSE 5000

ENTRYPOINT ["/bin/api"]
