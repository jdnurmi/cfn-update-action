ARG ALPINE_VER=3.15
# Container image that runs your code
FROM golang:1.18-alpine${ALPINE_VER}

# Copies your code file from your action repository to the filesystem path `/` of the container
ADD . /usr/src/cfn-update
WORKDIR /usr/src/cfn-update
RUN go build

FROM alpine:${ALPINE_VER}
COPY --from=0 /usr/src/cfn-update/cfn-update /usr/bin/cfn-update
# Code file to execute when the docker container starts up (`entrypoint.sh`)
ENTRYPOINT ["/usr/bin/cfn-update"]
