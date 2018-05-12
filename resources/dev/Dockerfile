FROM golang:1.10

ENV APP_DIR $GOPATH/src/github.com/dohernandez/market-manager

COPY . ${APP_DIR}
WORKDIR ${APP_DIR}

ARG GITHUB_TOKEN=${GITHUB_TOKEN}

RUN make deps-dev

RUN make run
