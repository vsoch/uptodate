FROM golang:alpine as builder

ARG version
LABEL maintainer="@vsoch"

LABEL "com.github.actions.name"="UpToDate Action"
LABEL "com.github.actions.description"="Check that repository assets are up to date"
LABEL "com.github.actions.icon"="activity"
LABEL "com.github.actions.color"="blue"

# Install module dependencies
WORKDIR /code
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest!
COPY . .

# Install some dependencies
RUN apk add --no-cache binutils build-base linux-headers

# Build the action
RUN make

# Multistage build to only copy over the binary
FROM alpine:latest@sha256:e1c082e3d3c45cccac829840a25941e679c25d438cc8412c2fa221cf1a824e6a

RUN apk add --no-cache git bash
WORKDIR /code
COPY --from=builder /code/uptodate /code/entrypoint.sh /code/
ENV PATH /code:$PATH
ENTRYPOINT ["/bin/bash", "/code/entrypoint.sh"]
