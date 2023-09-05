# Use Ubuntu as our builder as it's more full featured
FROM ubuntu:22.04 AS builder
RUN apt-get update && apt-get -y install ca-certificates golang-go
WORKDIR /tmp/build
RUN mkdir ./src ./bin
COPY app/*.go ./src/

# Install Go dependencies
RUN go mod init letsv-certchecker
RUN go get github.com/aws/aws-sdk-go-v2/aws
RUN go get github.com/aws/aws-sdk-go-v2/config
RUN go get github.com/aws/aws-sdk-go-v2/service/ssm

# Build our app
RUN go build -o bin/letsv-certchecker src/letsv-certchecker.go


# Drop to alpine, the smallest Linux distro we can find, as Go binaries are statically
#	linked and all we need is a Kernel
FROM alpine:3.18
WORKDIR /opt/letsvalidate
COPY --from=builder /tmp/build/bin/letsv-certchecker .
ENTRYPOINT ["./letsv-certchecker"]

