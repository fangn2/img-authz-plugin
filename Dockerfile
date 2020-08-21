# Docker Image Authorization Plugin
# Build tools image
FROM ubuntu:20.04

RUN export DEBIAN_FRONTEND=noninteractive; \
    ln -fs /usr/share/zoneinfo/Europe/Paris /etc/localtime && \
    apt-get update && \
    apt-get install -y tzdata && \
    dpkg-reconfigure --frontend noninteractive tzdata
RUN apt-get update && \
     apt-get install -y vim make golang git
