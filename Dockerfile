FROM golang AS builder

WORKDIR /go/src/github.com/mdlayher/unifi_exporter/
COPY . .

RUN go get -d ./... && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./cmd/unifi_exporter

FROM scratch

COPY --from=builder /go/src/github.com/mdlayher/unifi_exporter/unifi_exporter ./

EXPOSE 9130
ENTRYPOINT ["./unifi_exporter"]
