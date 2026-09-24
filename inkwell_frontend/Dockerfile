# ============================================================================
# 第一阶段: 构建(builder) —— 在 Node 里跑 vite build, 只把 dist/ 带进运行镜像
# ============================================================================
# 钉住 Node 大版本: vite 6 要求 Node 18+, 用 node:alpine 这种滚动 tag 容易被换成不兼容的版本
FROM node:22-alpine AS builder

# 打包时的接口前缀: 前端默认就是相对路径 /api/v1(交给 nginx 反代), 一般不用改。
# 如果要把静态站点部署到别的地方、接口在另一个域名上, 构建时覆盖这个参数即可:
#   docker build --build-arg VITE_API_BASE=https://api.example.com/api/v1 ...
ARG VITE_API_BASE=/api/v1
ENV VITE_API_BASE=${VITE_API_BASE}

WORKDIR /build

# 先拷依赖清单再安装: 只要 package.json/package-lock.json 没变, 这一层就一直命中缓存。
# 用 npm ci 而不是 npm install: 严格按 lock 文件装, 本地/CI/镜像里的依赖版本一致
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# npm run build = vue-tsc --noEmit && vite build, 打包产物在 /build/dist
RUN npm run build

# ============================================================================
# 第二阶段: 运行(runtime) —— nginx 托管静态文件, 顺带把接口反代到后端
# ============================================================================
FROM nginx:1.27-alpine

ENV TZ=Asia/Shanghai

# tzdata: 基础镜像里只有 UTC, 装上之后 nginx 的访问日志时间才是东八区
RUN set -eux \
    && apk add --no-cache tzdata \
    && ln -snf "/usr/share/zoneinfo/${TZ}" /etc/localtime \
    && echo "${TZ}" > /etc/timezone

# 覆盖官方镜像自带的示例站点(静态文件 + /api/v1、/swagger 反代, 见 nginx.conf)
COPY nginx.conf /etc/nginx/conf.d/default.conf

COPY --from=builder /build/dist /usr/share/nginx/html

EXPOSE 80

# 探活: 首页能返回 200 就算健康(alpine 自带 busybox 的 wget)
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=5 \
    CMD wget -q -O /dev/null http://127.0.0.1/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
