FROM golang

ARG registries

WORKDIR /opt

COPY . .

RUN make --makefile=Makefile.src && \
    make --makefile=Makefile.src install && \
    make --makefile=Makefile.src clean

# install Notary and a pre-requisite
ENV GO111MODULE=on

RUN git clone https://github.com/theupdateframework/notary.git && \
    cd notary && \
    go get github.com/theupdateframework/notary && \
    go install -tags pkcs11 github.com/theupdateframework/notary/cmd/notary

# empty unless the image is specifically built with it
# the docker plugin install command will set this later if needed
ENV REGISTRIES=${registries} \
    PATH=${PATH};/go/bin

ENTRYPOINT /usr/libexec/img-authz-plugin
