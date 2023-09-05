# Use Ubuntu as our builder as it's more full featured
FROM ubuntu:22.04 AS builder
RUN apt-get update && apt-get -y install golang-go



# Drop to alpine, the smallest Linux distro we can find, as Go binaries are statically
#	linked and all we need is a Kernel
#FROM alpine:3.18

