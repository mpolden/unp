FROM ubuntu:14.04

# Time zone
ENV DEBIAN_FRONTEND noninteractive
RUN echo "Europe/Oslo" > /etc/timezone
RUN dpkg-reconfigure tzdata

# Enable multiverse
RUN echo 'deb http://no.archive.ubuntu.com/ubuntu/ trusty multiverse' >> /etc/apt/sources.list
RUN echo 'deb-src http://no.archive.ubuntu.com/ubuntu/ trusty multiverse' >> /etc/apt/sources.list
RUN echo 'deb http://no.archive.ubuntu.com/ubuntu/ trusty-updates multiverse' >> /etc/apt/sources.list
RUN echo 'deb-src http://no.archive.ubuntu.com/ubuntu/ trusty-updates multiverse' >> /etc/apt/sources.list

# Install dependencies
RUN apt-get -y update
RUN apt-get -y install dtrx unrar

# Add app
RUN mkdir /app
ADD bin/gounpack /app/gounpack
RUN chmod 0755 /app/gounpack
ENTRYPOINT ["/app/gounpack"]
