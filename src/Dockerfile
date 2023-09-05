# Use Ubuntu as our builder as it's more full featured
FROM ubuntu:22.04 AS builder
RUN apt-get update && apt-get -y install golang-go
WORKDIR /tmp/build
RUN mkdir ./src ./bin
COPY app/*.go ./src/
RUN go build -o bin/letsv-certchecker src/letsv-certchecker.go


# Drop to alpine, the smallest Linux distro we can find, as Go binaries are statically
#	linked and all we need is a Kernel
FROM alpine:3.18
WORKDIR /opt/letsvalidate
COPY --from=builder /tmp/build/bin/letsv-certchecker .
ENTRYPOINT ["./letsv-certchecker"]
