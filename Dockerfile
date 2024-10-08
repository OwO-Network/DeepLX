FROM golang:1.23.2 AS builder
WORKDIR /go/src/github.com/OwO-Network/DeepLX
COPY . .
RUN go get -d -v ./
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o deeplx .

FROM alpine:latest

ENV TZ Asia/Shanghai
RUN apk add --no-cache tzdata \
    && ln -snf /usr/share/zoneinfo/$TZ /etc/localtime \
    && echo $TZ > /etc/timezone \

EXPOSE 1188

WORKDIR /app
COPY --from=builder /go/src/github.com/OwO-Network/DeepLX/deeplx /app/deeplx
CMD ["/app/deeplx"]
