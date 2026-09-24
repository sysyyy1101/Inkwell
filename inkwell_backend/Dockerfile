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
# 用 bookworm(bullseye 已经进入 LTS 收尾阶段, 换一个还在常规支持期的基础镜像);
# 二进制是 CGO_ENABLED=0 编出来的静态文件, 换基础镜像不影响它怎么跑
FROM debian:bookworm-slim

# 镜像元信息: 别人 docker inspect 就能看到这个镜像对应哪次提交。
# CI(见 .github/workflows/docker-publish.yml)会传入这两个参数, 本地构建就用默认值
ARG VERSION=dev
ARG VCS_REF=unknown
LABEL org.opencontainers.image.title="inkwell-backend" \
      org.opencontainers.image.description="Inkwell 社区后端(Go + Gin + MySQL + Redis)" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.source="https://github.com/sysyyy1101/Inkwell" \
      org.opencontainers.image.licenses="MIT"

# 容器默认是 UTC, 而 MySQL 连接串用了 loc=Local: 不设时区的话, 容器里的日志时间
# 和入库时间都会差 8 小时
ENV TZ=Asia/Shanghai

COPY ./wait-for.sh /
COPY ./conf /conf
COPY --from=builder /build/inkwell /
COPY --from=builder /build/seed /

# 在 Windows 上 clone 出来的代码, 只要 git 的 core.autocrlf 开着, wait-for.sh 就会是 CRLF
# 行尾, 而 "#!/bin/bash\r" 这种 shebang 在 Linux 里会直接报 bad interpreter(容器起不来)。
# 这里统一把 \r 清掉; 仓库根目录的 .gitattributes 也钉了 *.sh 用 LF, 算是双保险
RUN sed -i 's/\r$//' /wait-for.sh

# netcat-openbsd: wait-for.sh 和 HEALTHCHECK 里的 nc; tzdata: TZ 用;
# ca-certificates: 以后要出网访问 HTTPS 用
RUN set -eux \
    && apt-get update \
    && apt-get install -y --no-install-recommends netcat-openbsd tzdata ca-certificates \
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
