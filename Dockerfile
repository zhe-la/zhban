FROM golang:alpine as build-env
WORKDIR /root/zhban
COPY ./ .
RUN apk add --update --no-cache git
RUN go get github.com/corpix/uarand
RUN go get golang.org/x/net/html/charset
RUN go get google.golang.org/grpc
RUN go get github.com/hashicorp/consul/api
RUN go build zhban_server/zhban.go
RUN rm Dockerfile && rm docker-compose.yml && rm zhban.go && rm zhban_test.go && rm README.md

FROM alpine:latest as prod-env
COPY --from=build-env /root/zhban/zhban/zhban_server ./
EXPOSE 3000/tcp
CMD ["./zhban_server", "-http", "3000"]
