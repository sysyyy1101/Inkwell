# ============================================================================
# 把 init.sql 预先塞进官方 MySQL 镜像的入口目录
# ----------------------------------------------------------------------------
# 有了它, 只用镜像的人不用下载仓库、也不用再挂载 init.sql: 容器第一次启动(数据目录为空)
# 就会自动建库建表, 并灌入 60 个用户 / 12 个版块 / 120 篇帖子 / 300 条评论。
#
# 官方 entrypoint 只在数据目录为空时执行 /docker-entrypoint-initdb.d/ 下的脚本,
# 所以不会像 mysqld --init-file 那样每次重启都把四张表 TRUNCATE 一遍。
#
# 构建上下文是仓库根目录:
#   docker build -f deploy/mysql/Dockerfile -t ghcr.io/sysyyy1101/inkwell-mysql:latest .
# (CI 里由 .github/workflows/docker-publish.yml 自动构建推送)
# ============================================================================
FROM mysql:8.0.19

# 时区: 容器默认是 UTC, 不设的话时间列和日志都会差 8 小时
ENV TZ=Asia/Shanghai \
    MYSQL_DATABASE=inkwell_pre \
    MYSQL_ROOT_HOST=%

# 这两个启动参数必须保留(和 inkwell_backend/docker-compose.yml 里那套一致):
#   --default-authentication-plugin=mysql_native_password  MySQL 8 默认的 caching_sha2_password
#                                                          和项目里的 go-sql-driver/mysql v1.4.0
#                                                          不兼容, 去掉之后后端连不上数据库
#   utf8mb4                                                中文/emoji 不乱码
# (排序规则和 inkwell_backend/docker-compose.yml 里保持一致: utf8mb4_general_ci,
#  两边不一样的话同名表在两边建出来的索引/排序行为可能不同)
CMD ["mysqld", "--default-authentication-plugin=mysql_native_password", \
     "--character-set-server=utf8mb4", "--collation-server=utf8mb4_general_ci"]

# 文件名加 00- 前缀: 以后要加别的初始化脚本时, 靠前缀决定执行顺序
# (MYSQL_ROOT_PASSWORD 这类密钥不写进镜像, 由 compose 在运行时传进来)
COPY inkwell_backend/init.sql /docker-entrypoint-initdb.d/00-init.sql
