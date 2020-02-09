FROM golang:latest AS builder

WORKDIR /go/src/geoip-server

ADD ./ /go/src/geoip-server
RUN go mod vendor && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o /geoip geoip.go

FROM alpine
COPY --from=builder /geoip /geoip
CMD /geoip
