FROM golang:alpine AS builder

# 为我们的镜像设置必要的环境变量
ENV GO111MODULE=on \
GOPROXY=https://goproxy.cn,direct \
CGO_ENABLED=0 \
GOOS=linux \
GOARCH=amd64

# 移动到工作目录：/build
WORKDIR /build

# 复制项目中的 go.mod 和 go.sum文件并下载依赖信息
COPY go.mod .
COPY go.sum .
RUN go mod download

# 将代码复制到容器中
COPY . .

# 将我们的代码编译成二进制可执行文件 inkwell
RUN go build -o inkwell .

###################
# 接下来创建一个小镜像
###################
# stretch 已经停止维护, apt 源失效会导致下面的 apt-get update 直接失败, 改用 bullseye
FROM debian:bullseye-slim

COPY ./wait-for.sh /
COPY ./conf /conf

# 从builder镜像中把可执行文件拷贝到当前目录
COPY --from=builder /build/inkwell /

RUN set -eux \
    && apt-get update \
    && apt-get install -y --no-install-recommends netcat \
    && rm -rf /var/lib/apt/lists/* \
    && chmod 755 wait-for.sh

# 需要运行的命令
# 配置文件路径可以通过 -conf 指定, 例如容器里使用 conf/config.docker.yaml
#ENTRYPOINT ["/inkwell", "-conf", "conf/config.yaml"]
