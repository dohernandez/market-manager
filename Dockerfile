FROM golang:1.10-alpine AS builder

WORKDIR /go/src/github.com/dohernandez/market-manager

ARG GITHUB_TOKEN

ARG VERSION

COPY . .

RUN apk add --update bash make git

RUN make run

# Documentation builder
FROM mattjtodd/raml2html:7.0.0 AS docs-builder

COPY docs /docs

RUN raml2html  -i "/docs/raml/api.raml" -o "/docs/api.html"

# The final container
FROM alpine

COPY --from=builder /go/src/github.com/dohernandez/market-manager/build/market-manager /

COPY --from=docs-builder /docs/api.html /docs/documentation-raml/index.html

COPY ./resources/migrations /resources/migrations

# To be able to use https
RUN apk add --no-cache ca-certificates
RUN apk add --update curl && \
    rm -rf /var/cache/apk/*

EXPOSE 8000

ENTRYPOINT ["/market-manager-service"]
CMD ["http"]
