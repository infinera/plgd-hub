FROM golang:1.23.9-alpine AS build
ARG DIRECTORY
ARG NAME
ARG VERSION
ARG COMMIT_DATE
ARG SHORT_COMMIT
ARG DATE
ARG RELEASE_URL
RUN apk add --no-cache build-base curl git
WORKDIR $GOPATH/src/github.com/plgd-dev/hub
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /usr/local/go
RUN ( patch -p1 < "$GOPATH/src/github.com/plgd-dev/hub/tools/docker/patches/shrink_tls_conn.patch" )
WORKDIR $GOPATH/src/github.com/plgd-dev/hub/tools/mongodb/admin-tool
RUN CGO_ENABLED=0 go build \
    -ldflags "-X github.com/plgd-dev/hub/v2/pkg/build.CommitDate=$COMMIT_DATE \
    -X github.com/plgd-dev/hub/v2/pkg/build.CommitHash=$SHORT_COMMIT \
    -X github.com/plgd-dev/hub/v2/pkg/build.BuildDate=$DATE \
    -X github.com/plgd-dev/hub/v2/pkg/build.Version=$VERSION \
    -X github.com/plgd-dev/hub/v2/pkg/build.ReleaseURL=$RELEASE_URL" \
    -o /go/bin/mongodb-admin-tool \
    ./

FROM alpine:3.22 AS security-provider
RUN apk add -U --no-cache ca-certificates \
    && addgroup -S nonroot \
    && adduser -S nonroot -G nonroot

FROM alpine:3.22 AS service
RUN apk add -U --no-cache bash
COPY --from=security-provider /etc/passwd /etc/passwd
COPY --from=security-provider /etc/group /etc/group
COPY --from=security-provider /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/bin/mongodb-admin-tool /usr/local/bin/mongodb-admin-tool
USER nonroot
ENTRYPOINT [ "/bin/bash" ]

