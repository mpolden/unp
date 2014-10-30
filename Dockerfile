FROM golang:onbuild

# Install dependencies
RUN DEBIAN_FRONTEND=noninteractive \
    sed -i -e 's/main$/main contrib non-free/' /etc/apt/sources.list && \
    apt-get -y update && \
    apt-get -y install dtrx unrar

ENTRYPOINT ["/go/bin/app"]
