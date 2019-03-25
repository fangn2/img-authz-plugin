# Docker Image Authorization Plugin
# Build tools image
FROM centos:7

MAINTAINER Chaitanya Prakash N <cpdevws@gmail.com>

RUN yum install -y epel-release && \
       yum --enablerepo=epel-testing install -y git make golang
