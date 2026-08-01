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
# 不再 COPY 生产 config.yaml,配置通过环境变量/挂载注入,避免镜像泄露密钥
RUN chmod +x gkube
EXPOSE 8080

ENTRYPOINT ["./gkube"]