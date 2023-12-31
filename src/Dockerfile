# Use alpine-based builder image with Go already installed, so CGO works and we
# don't need all the weird CGO_ENABLED=0 pain
FROM golang:alpine AS builder
WORKDIR /tmp/build
RUN mkdir ./src ./bin
COPY app/*.go ./src/

# Create Go module so that "go get" words, and then install our Go dependencies 
RUN go mod init letsv-certchecker

# Was going to use lib/pq, but the maintainers are saying it's on ice:
#   https://github.com/lib/pq
# They point people towards pgx
#   https://github.com/jackc/pgx
RUN go get github.com/aws/aws-sdk-go-v2/aws github.com/aws/aws-sdk-go-v2/config \
    github.com/aws/aws-sdk-go-v2/service/ssm github.com/jackc/pgx/v5

# Build our app 
RUN go build -o bin/letsv-certchecker src/letsv-certchecker.go

# Start with a fresh alpine *without* any Go build tools installed. 
#   Alpine is the smallest Linux distro we can find, and since Go binaries are statically
#	linked, all we need is a Linux kernel
FROM alpine:3.18
WORKDIR /opt/letsvalidate
COPY --from=builder /tmp/build/bin/letsv-certchecker .
ENTRYPOINT ["./letsv-certchecker"]
#ENTRYPOINT ["/bin/sh"]
