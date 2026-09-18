# ============================================================================
# 第一阶段: 构建(builder) —— 在"大"镜像里编译, 只把编译产物带进运行镜像
# ============================================================================
# 钉住 Go 大版本: go.mod 要求 1.20+, 用 golang:alpine 这种滚动 tag 哪天被换成
# 不兼容的版本就麻烦了
FROM golang:1.25-alpine AS builder

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0

# 这里刻意不写死 GOOS/GOARCH: 交给 BuildKit 按目标平台构建。
# 写死成 linux/amd64 的话, 在 arm64 机器上会把 amd64 的二进制塞进 arm64 的
# 运行镜像, 启动直接 exec format error。
WORKDIR /build

# 先拷依赖清单再下载: 只要 go.mod/go.sum 没变, 这一层就一直命中缓存
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

# 一次编译两个可执行文件:
#   inkwell  后端服务
#   seed     把 MySQL 里的帖子灌进 Redis(榜单重建工具, 见 README 第 7 节)
RUN go build -trimpath -ldflags "-s -w" -o inkwell . \
    && go build -trimpath -ldflags "-s -w" -o seed ./cmd/seed

# ============================================================================
# 第二阶段: 运行(runtime) —— 只带二进制 + 配置 + 等待脚本
# ============================================================================
FROM debian:bullseye-slim

# 容器默认是 UTC, 而 MySQL 连接串用了 loc=Local: 不设时区的话, 容器里的日志时间
# 和入库时间都会差 8 小时
ENV TZ=Asia/Shanghai

COPY ./wait-for.sh /
COPY ./conf /conf
COPY --from=builder /build/inkwell /
COPY --from=builder /build/seed /

# netcat: wait-for.sh 用; tzdata: TZ 用; ca-certificates: 以后要出网访问 HTTPS 用
RUN set -eux \
    && apt-get update \
    && apt-get install -y --no-install-recommends netcat tzdata ca-certificates \
    && ln -snf "/usr/share/zoneinfo/${TZ}" /etc/localtime \
    && echo "${TZ}" > /etc/timezone \
    && rm -rf /var/lib/apt/lists/* \
    && chmod 755 /wait-for.sh /inkwell /seed

EXPOSE 8081

# 探活: 打项目自带的健康检查接口所在端口。
# netcat 只能判断"端口通了"(对单进程服务来说够了, 端口能监听就说明 MySQL/Redis/
# 密钥/雪花 ID 都已经初始化成功); 想连响应内容一起校验, 可以装 curl 换成
#   curl -fsS http://127.0.0.1:8081/api/v1/ping
HEALTHCHECK --interval=10s --timeout=3s --start-period=20s --retries=5 \
    CMD ["nc", "-z", "127.0.0.1", "8081"]

# 默认用本地配置启动, compose 里会覆盖成 -conf ./conf/config.docker.yaml
CMD ["/inkwell", "-conf", "./conf/config.yaml"]
