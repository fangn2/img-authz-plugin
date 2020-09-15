FROM golang as plugin-builder

WORKDIR /opt

COPY . .

RUN make --makefile=Makefile.src && \
    make --makefile=Makefile.src install

# install Notary and a pre-requisite
ENV GO111MODULE=on

RUN git clone https://github.com/theupdateframework/notary.git && \
    cd notary && \
    go get github.com/theupdateframework/notary && \
    go install -tags pkcs11 github.com/theupdateframework/notary/cmd/notary

#---#

FROM alpine@sha256:c929c5ca1d3f793bfdd2c6d6d9210e2530f1184c0f488f514f1bb8080bb1e82b

RUN apk update && apk add --no-cache libc6-compat

COPY --from=plugin-builder /usr/libexec/img-authz-plugin /usr/libexec/img-authz-plugin
COPY --from=plugin-builder /go/bin/notary /go/bin/notary

ENV PATH=${PATH}:/go/bin \
    CGO_ENABLED=0

ENTRYPOINT /usr/libexec/img-authz-plugin
