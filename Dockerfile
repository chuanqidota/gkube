FROM golang:1.23.0 AS builder
RUN go env -w GO111MODULE=on \
    && go env -w CGO_ENABLED=0 \
    && go env -w GOOS=linux \
    && go env -w GOPROXY=https://goproxy.cn,direct
RUN mkdir -p /opt
WORKDIR /opt
COPY . .
RUN go mod tidy
RUN go build -o gkube main.go

FROM alpine:3
RUN mkdir -p /opt
WORKDIR /opt
COPY --from=builder /opt/gkube /opt/gkube
COPY --from=builder /opt/config/config.yaml /opt/config/config.yaml
RUN chmod +x gkube
EXPOSE 8080

ENTRYPOINT ["./gkube"]