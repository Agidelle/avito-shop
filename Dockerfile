FROM golang:1.23-alpine3.20 AS build

WORKDIR /app

RUN apk add git


ENV APP_NAME="avito-test"
ARG GITLAB_CREDS
ARG COMMIT_HASH="latest"
ARG COMMIT_TIME="latest"
ARG VERSION="dev"
ARG GOPRIVATE
ARG GOPROXY
ARG GIT_TOKEN
ARG GIT_USER
ARG GONOSUMDB

RUN mkdir /out
COPY . /app/

RUN go build  \
    -o /out/${APP_NAME}  \
    avito-shop


FROM alpine:3.20

WORKDIR /app

COPY --from=build /out/avito-test /app/
COPY --from=build /app/config/config.yaml /app/config/

EXPOSE 8080

CMD ["/app/avito-test", "serve"]
