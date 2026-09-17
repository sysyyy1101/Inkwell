# Inkwell 前端 (Vue 3 + Vite + TypeScript)

这是 inkwell 社区的前端单页应用, 调用的接口全部来自 `../inkwell_backend`(前缀 `/api/v1`)。

## 快速开始

前置条件: 后端已经在 `http://127.0.0.1:8081` 启动(见根目录 README 的 5.5 节), Node.js 18+。

```bash
cd inkwell_frontend
npm install
npm run dev
```

打开 `http://127.0.0.1:8080` 即可。开发服务器会把 `/api/v1` 与 `/swagger` 都代理到 `http://127.0.0.1:8081`, 所以浏览器里不存在跨域问题, 页脚的"接口文档"链接(相对路径 `/swagger/index.html`)在开发环境也能直接打开。

## npm 脚本

| 命令 | 作用 |
| --- | --- |
| `npm run dev` | 启动开发服务器(8080 端口, 带接口代理) |
| `npm run serve` | `dev` 的别名 |
| `npm run build` | 先 `vue-tsc --noEmit` 做类型检查, 再打包到 `dist/` |
| `npm run build:only` | 只打包, 不做类型检查 |
| `npm run preview` | 本地预览 `dist/`(8080 端口, 同样带接口代理) |
| `npm run type-check` | 只做类型检查 |

端口固定为 8080(`strictPort: true`), 被占用时会直接报错; 想换后端地址可以用环境变量:

```bash
$env:VITE_DEV_PROXY="http://127.0.0.1:9090"; npm run dev     # PowerShell
VITE_DEV_PROXY=http://127.0.0.1:9090 npm run dev             # bash
```

## 页面与接口对应关系

| 路由 | 页面 | 用到的接口 |
| --- | --- | --- |
| `/` | 首页(热门/最新/时间榜) | `GET /post?order=score\|time&page=N`、`GET /post2?page&size`、`GET /community`、`GET /ping` |
| `/community/:id` | 版块页 | `GET /community/:id`、`GET /community/:id/post?order&page` |
| `/post/:id` | 帖子详情 + 评论区 | `GET /post/:id`、`GET /comment?post_id=`、`GET /comment?ids=`、`POST /comment`、`POST /vote` |
| `/user/:id` | 用户主页(点顶栏/帖子/评论里的头像进入; 发帖记录是朋友圈式时间轴, 左侧按天显示日期) | `GET /user/:id`、`GET /user/:id/post` |
| `/publish` | 发帖(需登录) | `GET /community`、`POST /post` |
| `/login` | 登录 | `POST /login`、`GET /refresh_token`(拦截器自动调用) |
| `/signup` | 注册 | `POST /signup` |
| `*` | 404 | `GET /ping` |

### 布局约定

- 首页、版块页、帖子详情页统一是「左侧版块导航 + 右侧内容」两栏: 左栏固定 240px, 内容列吃掉剩余宽度(容器最宽 1400px), 所以列表会一直延伸到右侧。
- 左栏顶部是「全部」入口, 与下面的版块项同款样式, 当前所在页会高亮; 版块项前面有一个按名称生成的稳定色点。
- 窄屏(小于 1024px)时左栏自动变成一行可横向滑动的胶囊, 不再占用竖向空间, 也不会把页面撑宽。
- 发帖、登录、注册、404 保持居中的单列布局。
- 站点没有图形 logo, 顶栏只保留导航与操作区。

### 字体与视觉

字体全部自托管(`@fontsource-variable/*`), 浏览器只会下载页面真正用到的 unicode-range 分片:

| 用途 | 字体 | 说明 |
| --- | --- | --- |
| 正文与 UI(拉丁) | Inter Variable | 数字统一 `tabular-nums`, 表格/计数不会跳动 |
| 正文与 UI(中文) | Noto Sans SC Variable(思源黑体) | 不再依赖系统里的微软雅黑, 任何机器上观感一致 |
| 大标题(拉丁) | Source Serif 4 Variable | 只用在 hero、页面标题、帖子详情标题、404 |
| 大标题(中文) | Noto Serif SC Variable(思源宋体) | 与 Source Serif 4 同一气质, 中文标题有阅读感 |
| 等宽 | JetBrains Mono Variable | 用户 ID、字数统计等 |

分离方式是在需要展示字体的标题上挂 `.font-display`(见 `src/styles/base.scss`), 其它地方一律走 `--font-sans`。

- **字重只有 5 档**(定义在 `src/styles/tokens.scss`): `400` 正文 / `500` 标签与次要强调 / `550` 按钮与导航 / `600` 卡片标题 / `650` 页面大标题, `700` 只留给 404 那种超大数字; 组件里不再出现写死的 `font-weight: 600`。思源黑体/宋体也是可变字体, 所以 `550`/`650` 这种中间值是真实字重, 不是浏览器伪粗。
- **行高与字距**同样走变量(`--leading-*` / `--tracking-*`): 大字标题收紧到 `-0.03em`, 中文正文保持 0, 正文行高 1.78。
- **背景**: 固定贴在视口上的独立图层 `.app-bg`(`--bg-wash`), 两层很淡的光晕 + 自上而下的过渡, 所以滚到页面任何位置都能看到; 底色放在 `html` 上, `body` 保持透明。
- **配色**: 版块色点与头像共用 6 个中低饱和色相, 定义在 `src/utils/format.ts` 的 `AVATAR_TONES`, 想更鲜艳或更素改这一处即可。
- **构建产物**: 带了两套中文字体后 `dist/static` 会有约 216 个字体分片(合计约 10MB), 但浏览器只下载当前页面用到的那几片(通常 100~300KB)。如果在意产物体积, 去掉思源宋体那一套(大标题回落到思源黑体)体积会减半。

### 顶栏与分段控件

- 顶栏是 **iOS 26 那种"渐变模糊"样式**: 没有 `border-bottom`、没有 `box-shadow`、自身背景透明, 分割感全部由 `.header__glass` 提供——一层比内容高 20px 的 `backdrop-filter: blur(20px) saturate(180%)`, 再用 `mask-image` 让模糊和底色一起向下淡出, 所以内容滚到顶栏下面是被逐渐雾化的, 不会出现一条硬边(实测界面上下 100px 的亮度变化都在 1~4 级以内, 没有台阶)。
- 雾的浓度由 `--header-bg` / `--header-bg-scrolled` 两个 token 控制(刻意压得很浅, 深色下是冷调薄雾而不是纯黑块), 滚动超过 8px 时加深一档。
- 顶栏的首页与主题切换共用 `.icon-button`(42×42、圆形描边、`::after` 把点击区域外扩 3px), 保证两个圆框完全一致; 首页按钮在当前路由高亮为主题色。
- 首页的热门/最新/时间榜是 `FeedTabs`: 选中底色是一个真实滑动的滑块(`transform` 动画 320ms), 用 `--segmented-*` 四个 token 控制底色、滑块、描边和阴影, 深色模式下滑块比容器亮一档。

### 动效约定

只动 `transform` / `opacity`, 时长 140~360ms, 全部受 `prefers-reduced-motion` 控制(base.scss 里的全局兜底会把动效压到 0.001ms):

- 列表入场: `.enter-rise` + `--enter-index`, 帖子卡片和评论按 35ms 错位淡入上移;
- 投票: 按钮按下缩放, 票数变化时数字轻微弹一下(`count-bump`);
- 顶栏: 滚动后加深并浮起阴影; 主题图标旋转交叉淡入切换;
- 版块导航项悬浮时右移 2px, 页面切换进场 260ms / 退场 160ms(退场更快)。

## 目录结构

```
src
├── api                 # 接口层
│   ├── http.ts         # axios 实例: 注入 token、1006/1008 自动刷新并重放、统一拆包与错误转换
│   ├── session.ts      # 登录态单一数据源(拦截器与 store 共用), 持久化到 localStorage
│   ├── types.ts        # 与后端约定的数据结构与响应码
│   └── auth/post/community/comment/system.ts
├── stores              # Pinia: auth / theme / toast / votes / community
├── router/index.ts     # 路由与登录守卫(meta.requireAuth / meta.guestOnly)
├── composables         # usePostFeed(分页加载)、useClickOutside(浮层关闭)
├── components          # 顶栏、页脚、版块导航、帖子卡片、投票控件、评论区、骨架屏、提示等
├── views               # 7 个页面
├── styles              # tokens.scss(设计变量/字体/深色主题)、base.scss(重置与通用类)
└── utils/format.ts     # 时间、计数、头像配色等
```

## 几个和后端强相关的实现细节

1. **64 位 ID 一律按字符串处理**: `post_id`/`user_id`/`comment_id` 是雪花 ID, 超过 JS 安全整数范围, 前端不做 `Number()` 转换。
2. **分页**: Redis 榜单(`/post`、`/community/:id/post`)每页固定 20 条且不支持 `size`; `/post2` 支持 `page/size`。两个接口都不返回总数, 所以按"本页是否取满"判断有没有下一页。
3. **榜单为空**: 榜单数据在 Redis 里, 帖子必须发帖/投票后才会写入。首页发现榜单为空时会自动切到"最新"(MySQL)并提示原因。
4. **投票是三态**: `direction` 为 `1/0/-1`, 重复投同一方向会被后端拒绝; 前端把"我对某帖投过什么票"记在 `localStorage` 的 `inkwell.votes` 里(`userID:postID -> direction`), 作者本人的帖子初始视为已赞成(后端发帖时会自动投赞成票)。如果本地记录与服务端不一致, 后端返回"已经投过票了"时前端会把本地状态同步过去。
5. **投票窗口**: 后端限制发帖超过 7 天不能投票, 前端会提前把按钮置灰并说明原因。
6. **评论树**: 后端返回的是带 `parent_id` 的扁平列表, 前端按 `parent_id` 组织成两级缩进的楼中楼, 并显示"回复 xxx"; 发表评论后调用 `GET /comment?ids=` 取回服务端数据再插入列表, 不在前端拼假数据。
7. **提示与跳转的顺序**: 提示(Toast)和路由跳转如果在同一个事件里先后触发, 跳转会吃掉这次 DOM 更新导致提示看不见, 所以统一用 `toast.notifyThenRedirect()`, 先渲染提示再跳转。
