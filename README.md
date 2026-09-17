# inkwell

一个基于 **Go + Gin + MySQL + Redis** 的社区后端, 带帖子榜单、投票、楼中楼评论和用户主页。

设计上刻意做了两件事: 业务数据(用户、帖子正文、评论)落 **MySQL**, 榜单与实时计数(票数、评论数)放 **Redis**, 两层用业务 ID 关联;
对外统一返回 `{"code":1000,"message":"success","data":{}}`, 由 `code` 表达业务结果。

| 层 | 技术 | 说明 |
| --- | --- | --- |
| 后端 | Go 1.20+ / Gin | 分层: router -> controller -> logic -> dao |
| 存储 | MySQL 8 / Redis 5 | MySQL 存实体数据, Redis 存榜单与实时计数 |
| 认证 | JWT (HS256) | 双 token(access 2 小时 / refresh 7 天) + 登录限流 |
| 文档 | swaggo/swag | 注解生成 Swagger, 访问 `/swagger/index.html` |
| 前端 | Vue 3 + Vite + TypeScript | 单页应用, 只调用 `/api/v1` |

## 1. 目录结构

```text
inkwell/
├── inkwell_backend/
│   ├── main.go / routers/     # 入口(配置->日志->JWT 密钥->MySQL/Redis->雪花 ID->路由)与路由注册
│   ├── cmd/seed/main.go       # 把 MySQL 的帖子灌进 Redis(初始化 / 丢数据后的重建工具)
│   ├── controller/ logic/     # 参数校验与错误码映射 / 业务编排(组合 MySQL 与 Redis)
│   ├── dao/mysql/ dao/redis/  # SQL / 榜单、投票、实时计数、回源重建
│   ├── models/ pkg/           # 请求响应模型与参数校验 / jwt、snowflake、ratelimit(都有单测)
│   ├── settings/ logger/ conf/  # 配置与热更新、zap 日志、两个环境的 yaml
│   ├── docs/                  # swag 生成的接口文档, 不要手改
│   ├── init.sql               # 建库建表 + 一套完整演示数据(会清空四张表)
│   └── Dockerfile / docker-compose.yml / wait-for.sh / .air.conf
└── inkwell_frontend/         # Vue3 + Vite 前端(见第 9 节)
```

## 2. 快速开始

### 2.1 依赖

| 依赖 | 版本 | 说明 |
| --- | --- | --- |
| Go | 1.20+ | 依赖要求, 开发环境用 go1.25 验证 |
| MySQL | 8.0(5.7 也可) | 业务数据 |
| Redis | 5.0+ | 榜单与实时计数 |
| Node.js / npm | 18+ / 9+ | 只在前端需要 |
| Docker | 可选 | 一键起 MySQL + Redis + 后端 |

### 2.2 初始化 MySQL

```bash
mysql -h127.0.0.1 -P3306 -uroot -p < inkwell_backend/init.sql
```

`init.sql` 会建库(`inkwell_pre`)、建表, 并灌入 60 个用户 / 12 个版块 / 120 篇帖子 / 300 条评论。
**它会先 TRUNCATE 这四张表**(脚本开头有说明), 属于"推倒重来"的初始化脚本; 只想建表就把那四行注释掉。

### 2.3 把数据灌进 Redis

榜单、票数、评论数都在 Redis 里, 只跑 `init.sql` 的话首页是空的, 需要再执行一次:

```bash
cd inkwell_backend
go run ./cmd/seed -conf ./conf/config.yaml
```

| 参数 | 默认 | 说明 |
| --- | --- | --- |
| `-clean` | `true` | 先删掉所有 `inkwell:` 前缀的 key(不影响同一个 db 里其它项目的 key) |
| `-votes` | `true` | 生成确定性的测试投票(同一篇帖子每次结果一样); 传 `false` 则只重建榜单结构 |
| `-conf` | `./conf/config.yaml` | 配置文件路径 |

它同时是**榜单重建工具**: Redis 被清空之后重新跑一遍就能把榜单恢复出来(`-votes=false` 时票数从已有的投票记录恢复)。

### 2.4 启动后端

```bash
cd inkwell_backend
go run .                            # 默认读 ./conf/config.yaml
go run . -conf ./conf/config.yaml   # 也可以显式指定
```

启动成功的标志: 日志里出现 `init logger success`、`init snowflake success` 以及 gin 打印的路由表。

### 2.5 启动前端(可选)

```bash
cd inkwell_frontend
npm install
npm run dev      # 打开 http://127.0.0.1:8080
```

开发服务器会把 `/api/v1` 和 `/swagger` 代理到 `http://127.0.0.1:8081`, 所以不存在跨域问题。

### 2.6 验证

```bash
curl http://127.0.0.1:8081/api/v1/ping          # {"code":1000,"message":"success","data":"pong"}
curl "http://127.0.0.1:8081/api/v1/post?page=1"  # 首页榜单(游客可访问), 每页 20 条
```

浏览器打开 `http://127.0.0.1:8081/swagger/index.html` 可以在线调试全部接口。

## 3. 配置说明

`conf/config.yaml`(本地开发):

```yaml
mode: "dev"           # dev 会额外把日志打到控制台; release 只写文件
port: 8081
machine_id: 1         # 雪花算法机器 ID(1-65535), 多实例部署时必须各不相同
jwt_secret: ""        # JWT 签名密钥; dev 模式留空用内置开发密钥, 其它模式必须显式配置
login_rate_limit:     # 登录接口的令牌桶限流(整个进程只有一个桶)
  rate: 2             # 每秒补充的令牌数
  burst: 10           # 桶容量(允许的突发次数)
# 其余是 log(级别/文件/切割)、mysql(连接与连接池)、redis(连接与连接池) 三段,
# 字段含义见 settings/settings.go 的 mapstructure tag
```

`conf/config.docker.yaml` 只有三处差异: `mode`/`log.level` 换成 `release`/`info`; `mysql.host`/`redis.host` 换成 compose 的服务名;
`jwt_secret` 给了一个占位值(**正式部署必须换成自己的随机密钥**)。

环境变量(优先级高于配置文件):

| 变量 | 说明 |
| --- | --- |
| `INKWELL_JWT_SECRET` | JWT 签名密钥。dev 模式留空会用内置开发密钥; **非 dev 模式留空、或直接用内置开发密钥, 启动都会失败** |
| `INKWELL_MACHINE_ID` | 覆盖 `machine_id` |

两个刻意的启动校验: `machine_id` 必须是 1-65535(用固定默认值会让多实例生成重复的业务 ID);
非 dev 模式必须显式配置 JWT 密钥、且不能等于代码里的内置开发密钥(内置密钥是公开的, 带上线等于任何人都能伪造 token)。

`settings.Init` 里注册了 `viper.OnConfigChange`, 改配置文件会自动刷新内存里的值, 但它**不会重建 MySQL/Redis 连接, 也不会重启 HTTP 服务**,
所以改端口或数据库地址仍然要重启进程。

## 4. 接口一览

通用约定:

- Base URL `http://<host>:8081/api/v1`; 请求与响应都是 `application/json; charset=utf-8`。
- 需要登录的接口带 `Authorization: Bearer <accessToken>`。
- 用户/帖子/评论 ID 是雪花 ID, JSON 里**一律用字符串**传输(超过 JS 安全整数), 版块 ID 用数字。
- 业务失败时 HTTP 状态码一般是 200, 要按 `code` 判断; 唯一例外是登录被限流(HTTP 429 + code 1010)。

公开接口(游客可访问):

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/signup` | 注册 |
| POST | `/login` | 登录(带令牌桶限流) |
| GET | `/refresh_token` | 用 refresh token 换新的 access token |
| GET | `/ping` | 健康检查 |
| GET | `/community` | 版块列表 |
| GET | `/community/:id` | 版块详情 |
| GET | `/community/:id/post` | 版块内榜单, 支持 `order`(`score`/`time`)与 `page` |
| GET | `/post` | 全站榜单, 支持 `order`/`page`, 每页固定 20 条 |
| GET | `/post2` | 直读 MySQL 的帖子列表, 支持 `page`/`size`(最大 50) |
| GET | `/post/:id` | 帖子详情(带 Redis 的实时票数与评论数) |
| GET | `/comment` | 评论列表, 支持 `post_id` 或 `ids`(可重复传) |
| GET | `/user/:id` | 用户主页信息(用户名、注册时间) |
| GET | `/user/:id/post` | 某个用户发布的帖子, 支持 `page`/`size` |

需要登录的接口:

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/post` | 发帖, 返回 `{"post_id":"..."}` |
| POST | `/vote` | 投票, 返回投票后的净票数 |
| POST | `/comment` | 发表评论或回复, 返回 `{"comment_id":"..."}` |

`POST /vote` 的请求体是 `{"post_id":"...","direction":1}`(`1` 赞成 / `0` 取消 / `-1` 反对), 成功时返回:

```json
{"code":1000,"message":"success","data":{"post_id":"3458764513820600001","vote_num":12}}
```

`vote_num` 是**净票数(赞成 - 反对, 可能为负)**, 客户端直接展示它就行, 不要自己按本地记录推算增量 ——
本机记录的投票方向可能缺失或过期(换设备、清缓存), 推算出来的值会和后端对不上。

响应码(定义在 `controller/code.go`):

| code | 含义 |
| --- | --- |
| 1000 | 成功 |
| 1001 | 请求参数错误(含"已经投过票了""已过投票时间""帖子不存在"这类业务拒绝) |
| 1002 | 用户名重复 |
| 1003 | 用户不存在 |
| 1004 | 用户名或密码错误(登录时两者不区分, 避免枚举用户名) |
| 1005 | 服务繁忙(服务端异常) |
| 1006 | 无效的 Token(前端据此触发刷新) |
| 1008 | 未登录(上下文里取不到用户 ID) |
| 1009 | 资源不存在(帖子/版块/用户不存在, 或路由不存在) |
| 1010 | 请求过于频繁(登录限流), 同时返回 HTTP 429 |

## 5. 关键实现

### 5.1 分层与依赖方向

`router -> controller -> logic -> dao(mysql / redis)`, 参数校验在 `models`。错误按"每层翻译一次"处理:
dao 返回 `mysql.ErrorUserExit` 这类错误, logic 用 `errors.Is` 判断后返回 `logic.ErrorUserExist`, controller 再映射成 `CodeUserExist`,
因此 controller 不依赖 dao 的具体错误。

### 5.2 一次发帖的链路

`JWTAuthMiddleware` 校验 token 并把 userID 放进 gin 上下文 -> `CreatePostHandler` 绑定参数(校验写在 `models.Post.UnmarshalJSON` 里)
-> `logic.CreatePost` 依次确认作者与版块存在、生成雪花 ID、写 MySQL、写 Redis 榜单 -> 返回 `{"post_id":"..."}`。

Redis 写失败只记日志、不影响接口结果: 帖子已经落库, 详情页仍然能打开; 返回失败反而会造成"提示发帖失败但帖子其实已经创建"的不一致。

### 5.3 投票: 净票数 + 热度分

`dao/redis/post.go` 的 `PostVote` 流程:

1. 校验 `direction` 只能是 `1/0/-1`;
2. 从时间榜取发帖时间(超过 7 天返回"已过投票时间"; 取不到说明帖子在 Redis 里没有记录,
   此时 `logic.Vote` 会回源 MySQL 确认帖子是否存在: 存在就用 MySQL 的数据重建榜单信息再重试一次投票,
   只有 MySQL 里也查不到才返回"帖子不存在或已过期");
3. 同一个方向重复投直接拒绝; 用 `TxPipeline` 写入新方向并给投票记录续期 7 天;
4. 用 `ZCount` 一次拿到赞成/反对票数, **净票数 = 赞成 - 反对** 写回帖子 Hash, 同时用 `Hot()` 重算热度分写回分数榜。

所以票数是净分: 点赞 +1、点踩 -1、赞成改成反对 -2、取消 ±1。热度分始终单独按赞成/反对计算 `Hot(ups, downs, date)`,
票数那部分取 `log10`(第一票收益最大, 之后越来越平), 时间部分按秒线性增长, 让新帖逐渐追上老帖。

### 5.4 评论: 计数以 MySQL 为准 + 丢失自愈

发表评论后不是 `HINCRBY +1`, 而是把 MySQL 的 `COUNT(*)`(权威值)写回 Redis, 这样历史上漂移过的计数会在下一次评论时被纠正。
如果发现帖子的 Redis Hash 不存在(被清空/被淘汰), 先用 `logic.RestorePostInfo` 回源 MySQL 重建榜单信息再写 ——
原来的实现在这种情况下会静默跳过, 评论数就永远停在 0 了。

### 5.5 Redis 数据结构

| Key | 类型 | 内容 |
| --- | --- | --- |
| `inkwell:post:{postID}` | Hash | 榜单里的冗余信息与计数: `title` `summary` `post:id` `user:id` `author:name` `community:id` `community:name` `time` `votes` `comments` |
| `inkwell:post:time` | ZSet | member = postID, score = 发帖时间(Unix 秒) |
| `inkwell:post:score` | ZSet | member = postID, score = Reddit 热度分 |
| `inkwell:post:voted:{postID}` | ZSet | member = userID, score = 投票方向(-1/0/1), 7 天过期 |
| `inkwell:community:{版块名}` | Set | member = postID, 版块内的帖子集合 |

榜单查询是 `ZRevRangeWithScores` 取一页(20 条) + 一次 pipeline 批量 `HGetAll` 拿详情, 避免 N 次网络往返;
版块榜先用 `ZInterStore` 求"版块集合 ∩ 榜单"的交集(带 60 秒缓存)再分页, 所以投票后它的排序最多滞后 1 分钟。

### 5.6 鉴权、刷新与登录限流

`pkg/jwt` 自定义 claims 并强制 HS256(解析时用 `ValidMethods` 限制, 防 `alg=none`), `iss = inkwell`: access token 2 小时并带 `user_id`,
refresh token 7 天、用户 ID 放在 `Subject`, 所以它不能当 access token 用; 密钥在启动时由 `pkg/jwt.Init` 确定。

`/refresh_token` 要求头部带(可能要过期的)access token、query 带 refresh token, 只有 access token 恰好过期才允许刷新(签名错误、格式错误一律拒绝);
前端在响应拦截器里遇到 1006/1008 会自动刷新并重放原请求, 并发请求只刷新一次。

登录接口还挂了一个令牌桶(`controller/ratelimit.go` + `pkg/ratelimit`): **整个进程只有一个桶**(不是按 IP 分桶), 参数来自
`login_rate_limit`, 令牌惰性补充(取的时候按流逝时间补, 不需要后台 goroutine), 被拦下来返回 HTTP 429 + code 1010。

### 5.7 日志

zap + lumberjack: JSON 编码、ISO8601 时间、级别大写、带短文件名与行号, 按 `max_size`/`max_age`/`max_backups` 切割保留;
`mode: dev` 时用 `zapcore.NewTee` 额外输出到控制台。业务代码统一用 `zap.L().Error(...)`, 失败日志里带 `post_id`、`author_id`、`sql` 这类上下文。
请求日志目前是 gin 默认格式; 想换成 zap 的 JSON 请求日志, 可以启用 `logger.GinLogger()` / `logger.GinRecovery(true)` 并把 `gin.Default()` 换成 `gin.New()`。

## 6. 测试数据

`init.sql` 里的数据是按"真实社区"的样子造的, 不是随机字符串堆出来的:

| 维度 | 规模 |
| --- | --- |
| 用户 | 60 个, 注册时间铺在一年里 |
| 版块 | 12 个: 编程语言与知识(Go、C/C++、Rust、Python、Java、MySQL、Linux 与云原生、JavaScript 前端、算法与数据结构)+ 运动(篮球、足球、跑步与健身) |
| 帖子 | 120 篇, 时间从 7 分钟前到一年前; 最近 7 天约 50 篇, 方便测投票 |
| 评论 | 300 条(含 15 条楼中楼回复); 时间都晚于所属帖子, `parent_id` 都指向同帖的评论 |
| 投票 | Redis 里每帖 2~55 人参与, 平均净票 23.6, 最高 +55, 有 7 篇是负票(用来验证负数展示与排序) |

帖子时间不是均匀分布的: 同一版块既有几分钟前的新帖也有一年前的老帖, 并且刻意让 `gopher_zhang`、`lisi_dev` 等用户在同一天发 2~3 篇,
用来验证用户主页按天分组的时间轴。`init.sql` 末尾会跑一遍自检(数量 + 引用完整性 + 时间分布), 期望值写在 SQL 里, 执行完可以直接对照。

测试账号(密码统一 `123456`):

| 用户名 | 说明 |
| --- | --- |
| `gopher_zhang` | 帖子最多, 有"同一天 3 篇"的记录 |
| `lisi_dev` / `wangwu_go` / `zhao_liu` | 普通用户 |
| `阿伟摸鱼` / `老王写代码` | 中文用户名 |

想重新灌一遍: 先跑 `init.sql`(TRUNCATE + 重新插入), 再跑 `go run ./cmd/seed`(`-clean` 会清掉旧榜单)。

## 7. Docker 部署

在 `inkwell_backend` 目录下执行 `docker-compose up --build`, 会起三个服务: `mysql8019`(23306:3306, root 密码 `root1234`)、
`redis507`(26379:6379)、`inkwell_app`(8081:8081, 用 `conf/config.docker.yaml` 启动)。

应用启动命令是 `./wait-for.sh redis507:6379 mysql8019:3306 -- ./inkwell -conf ./conf/config.docker.yaml`, 先等依赖端口通了再起服务。
容器里**不会**自动跑 `cmd/seed`, 首页榜单要自己灌一次; 镜像里也只有后端二进制, 前端 `dist/` 需要自己用 nginx 托管并把
`/api/v1`、`/swagger` 反代到后端。

## 8. 测试

后端单测不依赖 MySQL/Redis(测的是纯函数、参数解析和路由文档), 可以直接跑:

```bash
cd inkwell_backend
go vet ./...
go test ./...
```

覆盖点: JWT 签发/解析/刷新/拒绝非 HS256/密钥初始化策略、令牌桶的补充与并发、登录限流返回 429+1010、
ID 的字符串与数字解析、帖子与评论参数校验、投票参数解析、用户名密码长度校验、中英混排摘要截断、
Swagger 可访问且"所有路由都有文档"。

改了 controller 上的 swagger 注解之后要重新生成文档:

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.6
cd inkwell_backend && swag init -g main.go -o docs
```

## 9. 前端简介

`inkwell_frontend` 是 Vue 3 + Vite + TypeScript 单页应用, 只负责界面, 数据都来自上面的接口:

| 路由 | 页面 | 主要接口 |
| --- | --- | --- |
| `/` | 首页: 热门 / 最新 / 时间榜 | `GET /post`、`GET /post2` |
| `/community/:id` | 版块页(两榜切换 + 加载更多) | `GET /community/:id`、`GET /community/:id/post` |
| `/post/:id` | 帖子详情 + 评论区(楼中楼回复) | `GET /post/:id`、`GET /comment`、`POST /comment`、`POST /vote` |
| `/user/:id` | 用户主页: 朋友圈式时间轴(按天分组, 同一天只显示一次日期) | `GET /user/:id`、`GET /user/:id/post` |
| `/publish` | 发帖(需登录) | `GET /community`、`POST /post` |
| `/login`、`/signup` | 登录 / 注册 | `POST /login`、`POST /signup` |

```bash
npm run dev       # 开发(8080, 代理 /api/v1 与 /swagger)
npm run build     # vue-tsc 类型检查 + 打包到 dist/
npm run preview   # 预览 dist/(同样带代理)
```

登录态放在 `localStorage` 的 `inkwell.auth`, 由 `src/api/session.ts` 统一读写(axios 拦截器也要用, 所以没放进 pinia);
响应拦截器遇到 1006/1008 会自动刷新 token 并重放请求; 投票方向记在 `inkwell.votes`(后端没有"查我投过什么"的接口),
但**票数一律用接口返回的 `vote_num`**, 本地记录只用来决定点下去该发哪个方向。

## 10. 常见问题

**首页没有帖子, 但 `/post2` 有**: 榜单在 Redis 里, 还没跑 `go run ./cmd/seed`, 跑一次就有了。

**投票报"帖子不存在或已过期"**: 只有一种情况 —— 帖子在 MySQL 和 Redis 里都不存在, 或者发帖超过 7 天(设计如此)。
帖子在 Redis 里丢了(不管只是 Hash 丢了, 还是整个 Redis 被 flush)会自动回源 MySQL 重建: 投票和评论都会触发重建,
重建后这条帖子会重新进入榜单。整个 Redis 被 flush 时首页仍是空的(榜单不会再自动补全), 需要跑一次 `cmd/seed` 把全部帖子灌回去。

**启动报 `machine_id 必须是 1-65535` 或 `必须配置 JWT 密钥`**: 这是刻意做的启动校验, 见第 3 节。

**登录返回 429 / code 1010**: 触发了登录限流(默认 2 个/秒、容量 10, 全进程共享一个桶), 等一两秒即可; 压测时把 `login_rate_limit` 调大。

**接口返回 1005 服务繁忙**: 一般是 MySQL/Redis 连接问题或 SQL 错误, 看 `log/inkwell.log` 里的 `zap.Error` 上下文。

**改配置没生效**: 热更新只刷新内存里的值, 端口和 MySQL/Redis 连接需要重启进程。

**前端报端口被占用**: 开发服务器固定 8080(`strictPort: true`), 不会静默换端口; 换端口要同步改 `vite.config.ts`。

## 11. 已知限制

1. **登录限流只覆盖登录接口**, 而且不按 IP/用户分桶; 发帖、评论、投票都没有频率限制。
2. **投票的写票与重算不是一次原子操作**(先写投票记录, 再按 `ZCount` 写净票数与热度分); 极小概率下两步之间进程退出会造成短暂不一致, 再投一次票就会自动纠正。运维工具也只有 `cmd/seed` 这个全量重建, 没有增量修复与定时对账。
3. **帖子和评论只能创建, 不能编辑/删除**; 评论列表不分页(一次返回该帖全部评论), 也没有举报功能; 列表接口都不返回总数, 前端只能按"本页是否取满"判断下一页。
4. **榜单里如果有帖子的 Hash 丢了会被跳过**, 一页可能不足 20 条, 极端情况下前端会提前显示"到底了"。
5. **没有"查询我对某帖的投票"接口**, 前端只能把方向记在浏览器本地; 换设备后会收到"已经投过票了", 前端据此同步状态。
6. **refresh token 不可撤销**: 没有黑名单/一次性机制, 登出只能清本地; 另外 `/refresh_token` 没有校验 refresh token 与 access token 是否属于同一个用户, 建议后续绑定 `sub` 与 `user_id`。
7. **Redis 里的榜单 key 没有 TTL**(只有投票记录 7 天过期), 会随帖子数一直增长; 生产环境建议按热度裁剪或分库, 并留意 `maxmemory-policy`。
8. **前端没有自动化测试与 lint**(只有 `vue-tsc` 类型检查), 也没有 CI; 多实例部署需要自己保证 `machine_id` 不同, 帖子 `status` 字段预留但没有做下架与可见性控制。
