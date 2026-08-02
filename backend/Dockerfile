FROM docker.1ms.run/library/golang:1.26 AS builder
RUN go env -w GO111MODULE=on \
    && go env -w CGO_ENABLED=0 \
    && go env -w GOOS=linux \
    && go env -w GOARCH=amd64 \
    && go env -w GOPROXY=https://goproxy.cn,direct
WORKDIR /opt
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o gkube main.go

FROM docker.1ms.run/library/alpine:3.22
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /opt
COPY --from=builder /opt/gkube .
# 不 COPY 生产 config.yaml,配置通过挂载 config.docker.yaml 注入,避免镜像泄露密钥
RUN chmod +x gkube
EXPOSE 8080

ENTRYPOINT ["./gkube"]
