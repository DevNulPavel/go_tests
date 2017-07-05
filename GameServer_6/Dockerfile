# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /GameServer_6

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN cd /GameServer_6; go build

# Run the outyet command by default when the container starts.
ENTRYPOINT /GameServer_6/GameServer_6

# Document that the service listens on port 8080.
EXPOSE 9999