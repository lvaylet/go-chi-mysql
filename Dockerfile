# Accept the Go version for the image as a build argument.
# Default to Go 1.12
ARG GO_VERSION=1.12

FROM golang:${GO_VERSION}-alpine AS builder

# Create a working directory, not necessarily under $GOPATH/src
# thanks to Go modules being used to handle the dependencies.
WORKDIR /usr/src/app

# Install git to let `go mod download` fetch dependencies
RUN apk add --no-cache git

# Download all the dependencies specified in the `go.{mod,sum}`
# files. Because of how the layer caching system works in Docker,
# the `go mod download` command will *only* be executed when one
# of the `go.{mod,sum}` files changes (or when another Docker
# instruction is added before this line). As these files do not
# change frequently (unless you are updating the dependencies),
# they can be simply cached to speed up the build.
COPY go.mod .
COPY go.sum .
RUN go mod download

# Bundle source code to working directory
COPY main.go .
COPY model.go .
COPY app.go .

# Cross compile the application
#
# Note the use of `CGO_ENABLED` to let the Go compiler link the
# libraries on the system. It is enabled by default for native
# build in order to reduce the binary size. This time we use
# `scratch` as our base image. It is a special Docker image with
# nothing in it (not even libraries). We need to disable the CGO
# parameter to let the compiler package all the libraries required
# by the application into the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /go/bin/main

# ---

FROM scratch AS production
COPY --from=builder /go/bin/main /go/bin/main
ENTRYPOINT ["/go/bin/main"]
