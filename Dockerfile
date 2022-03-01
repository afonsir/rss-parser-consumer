FROM golang:1.17
WORKDIR /go/src/github.com/afonsir/rss-parser-consumer
COPY main.go go.mod go.sum ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/afonsir/rss-parser-consumer/app .
CMD ["./app"]
