FROM golang:onbuild

# Install dependencies
RUN sed -i -e 's/main$/main contrib non-free/' /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y dtrx unrar && \
    rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/go/bin/app"]
