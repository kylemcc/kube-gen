FROM golang:alpine as builder
LABEL maintainer "Kyle McCullough <kylemcc@gmail.com>"

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN apk add --no-cache \
		ca-certificates

COPY . /go/src/github.com/kylemcc/kube-gen

RUN set -x \
		&& apk add --no-cache --virtual .build-deps \
			git \
			gcc \
			libc-dev \
			libgcc \
		&& cd /go/src/github.com/kylemcc/kube-gen/cmd/kube-gen \
		&& CGO_ENABLED=0 go build -ldflags "-extldflags -static" \
		&& mv kube-gen /usr/bin/kube-gen \
		&& apk del .build-deps \
		&& rm -rf /go \
		&& echo "Build complete."

FROM scratch

COPY --from=builder /usr/bin/kube-gen /usr/bin/kube-gen
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

ENTRYPOINT ["kube-gen"]
