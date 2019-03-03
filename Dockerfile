FROM golang:alpine
WORKDIR /root/gozhban
COPY ./ .
RUN go get github.com/corpix/uarand
RUN go get golang.org/x/net/html/charset
RUN go build zhban.go
RUN rm Dockerfile && rm docker-compose.yml
EXPOSE 3000/tcp
CMD ["./zhban"]