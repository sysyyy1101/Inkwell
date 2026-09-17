-- ============================================================================
-- inkwell 数据库初始化脚本(建库 + 建表 + 灌入一套完整的演示数据)
-- ----------------------------------------------------------------------------
-- 一、这个文件做什么
--   1. 建库 inkwell_pre 与四张表(结构与线上一致);
--   2. **清空** user / community / post / comment 四张表(TRUNCATE), 再灌入:
--        60 个用户 / 12 个版块 / 120 篇帖子 / 300 条评论(含 20 条楼中楼回复);
--      12 个版块: Go、C/C++、篮球、足球、Rust、Python、Java、MySQL、
--                  Linux 与云原生、JavaScript 前端、算法与数据结构、跑步与健身
--      (既有主流编程语言与工程知识, 也有篮球/足球/跑步健身这些运动版块)
--   3. 帖子与评论的时间是"相对脚本执行时刻"随机铺开的, 最早大约一年前,
--      最新的在几分钟前, 并且故意让一部分帖子集中在同一天(用来验证按天分组的时间轴)。
--
--   注意: TRUNCATE 会删掉这四张表里的**全部**数据(包括你自己注册的账号),
--   这是一份"推倒重来"的脚本。只想建库建表的话, 把下面的 TRUNCATE 四行注释掉即可。
--
-- 二、怎么跑
--   命令行:   mysql -h127.0.0.1 -P3306 -uroot -p < inkwell_backend/init.sql
--   客户端:   use inkwell_pre;  source /path/to/init.sql;
--   docker:   这个文件被 docker-compose 挂载成 MySQL 容器的 --init-file,
--             容器第一次启动时会自动执行(见 docker-compose.yml)。
--
-- 三、数据只有 MySQL 一半, 还要灌 Redis
--   榜单、票数、评论数都在 Redis 里, 只跑这个文件的话:
--     GET /api/v1/post     首页榜单是空的(Redis 里没有榜单)
--     GET /api/v1/post2    能看到全部 120 篇(这个接口直读 MySQL)
--     GET /api/v1/post/:id 能打开帖子, 但票数/评论数是 0
--   把下面的数据灌进 Redis(会按帖子生成一份确定性的、带正负票的投票记录):
--     cd inkwell_backend && go run ./cmd/seed -conf ./conf/config.yaml
--   加 -votes=false 则只重建榜单结构、不生成测试投票(票数从已有的投票记录里恢复)。
--
-- 四、测试账号
--   密码统一是 123456(password 列是真实的 bcrypt 哈希, 可以直接登录), 例如:
--     gopher_zhang / lisi_dev / wangwu_go / zhao_liu / sun_qi / 阿伟摸鱼
--   用户名列表见下面的 INSERT, 一共 60 个, 覆盖了从 2024 年到 2025 年注册的用户。
--
-- 五、ID 约定
--   用户/帖子/评论 ID 都是雪花量级的 64 位整数(约 3.46e18), 远大于 JS 的安全整数
--   上限 2^53-1(9007199254740991), 用来验证"64 位 ID 必须用字符串传输"这条约定。
--   版块 ID 沿用原来设计里的小整数 1~12。
-- ============================================================================

SET NAMES utf8mb4;

CREATE DATABASE IF NOT EXISTS `inkwell_pre` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE `inkwell_pre`;

-- ----------------------------------------------------------------------------
-- 1. 建表(与线上一致, 已存在就跳过)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `user` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `user_id` bigint(20) NOT NULL,
    `username` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
    `password` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
    `email` varchar(64) COLLATE utf8mb4_general_ci,
    `gender` tinyint(4) NOT NULL DEFAULT '0',
    `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_username` (`username`) USING BTREE,
    UNIQUE KEY `idx_user_id` (`user_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `community` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `community_id` int(10) unsigned NOT NULL,
  `community_name` varchar(128) COLLATE utf8mb4_general_ci NOT NULL,
  `introduction` varchar(256) COLLATE utf8mb4_general_ci NOT NULL,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_community_id` (`community_id`),
  UNIQUE KEY `idx_community_name` (`community_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `post` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `post_id` bigint(20) NOT NULL COMMENT '帖子id',
  `title` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT '标题',
  `content` varchar(8192) COLLATE utf8mb4_general_ci NOT NULL COMMENT '内容',
  `author_id` bigint(20) NOT NULL COMMENT '作者的用户id',
  `community_id` bigint(20) NOT NULL COMMENT '所属社区',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '帖子状态',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_post_id` (`post_id`),
  KEY `idx_author_id` (`author_id`),
  KEY `idx_community_id` (`community_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `comment` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `comment_id` bigint(20) unsigned NOT NULL,
  `content` text COLLATE utf8mb4_general_ci NOT NULL,
  `post_id` bigint(20) NOT NULL,
  `author_id` bigint(20) NOT NULL,
  `parent_id` bigint(20) NOT NULL DEFAULT '0',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_comment_id` (`comment_id`),
  KEY `idx_author_Id` (`author_id`),
  KEY `idx_post_id` (`post_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------------------------------------------------------
-- 2. 清空旧数据(推倒重来; 不想删数据就把这 4 行注释掉)
-- ----------------------------------------------------------------------------
TRUNCATE TABLE `comment`;
TRUNCATE TABLE `post`;
TRUNCATE TABLE `community`;
TRUNCATE TABLE `user`;

-- 时间基准: 后面所有时间都写成"距脚本执行时刻多少分钟", 这样不管哪天执行,
-- 数据看起来都是"最近一年持续有人在发帖", 而不是一堆写死的日期。
SET @t_now := NOW();

-- ----------------------------------------------------------------------------
-- 3. 用户(60 个)
--    密码统一 123456, 下面几个 bcrypt 哈希轮换使用(DefaultCost=10, 60 字节)。
-- ----------------------------------------------------------------------------
INSERT INTO `user` (user_id, username, password, create_time) VALUES
(3458764513820541001, 'gopher_zhang',   '$2a$10$jtWYHLoo2DcOahoNIQjXvuxPnXP9k09PE93YLMME/33MqZj1ORSNa', @t_now - INTERVAL 391 DAY),
(3458764513820541002, 'lisi_dev',       '$2a$10$SrG0M/pC9eHkgkNiLXFLNuwnbT7imlTOsM.50iYy5PZhL/JtDNEki', @t_now - INTERVAL 386 DAY),
(3458764513820541003, 'wangwu_go',      '$2a$10$wspDbTJWmntUZcHrVwsTfeHnzO8UtkAo2D9n2rZ63ZcR0owrWMKeO', @t_now - INTERVAL 378 DAY),
(3458764513820541004, 'zhao_liu',       '$2a$10$LjHOli0cEypdGtENHmg9y.hEpyWPr0zEyBS60g.cU2DwE8f5/6FEW', @t_now - INTERVAL 372 DAY),
(3458764513820541005, 'sun_qi',         '$2a$10$jtWYHLoo2DcOahoNIQjXvuxPnXP9k09PE93YLMME/33MqZj1ORSNa', @t_now - INTERVAL 366 DAY),
(3458764513820541006, 'zhou_ba',        '$2a$10$SrG0M/pC9eHkgkNiLXFLNuwnbT7imlTOsM.50iYy5PZhL/JtDNEki', @t_now - INTERVAL 360 DAY),
(3458764513820541007, 'wu_jiu',         '$2a$10$wspDbTJWmntUZcHrVwsTfeHnzO8UtkAo2D9n2rZ63ZcR0owrWMKeO', @t_now - INTERVAL 351 DAY),
(3458764513820541008, 'zheng_shi',      '$2a$10$LjHOli0cEypdGtENHmg9y.hEpyWPr0zEyBS60g.cU2DwE8f5/6FEW', @t_now - INTERVAL 344 DAY),
(3458764513820541009, 'qian_yi',        '$2a$10$jtWYHLoo2DcOahoNIQjXvuxPnXP9k09PE93YLMME/33MqZj1ORSNa', @t_now - INTERVAL 338 DAY),
(3458764513820541010, 'shen_er',        '$2a$10$SrG0M/pC9eHkgkNiLXFLNuwnbT7imlTOsM.50iYy5PZhL/JtDNEki', @t_now - INTERVAL 331 DAY),
(3458764513820541011, 'han_san',        '$2a$10$wspDbTJWmntUZcHrVwsTfeHnzO8UtkAo2D9n2rZ63ZcR0owrWMKeO', @t_now - INTERVAL 325 DAY),
(3458764513820541012, 'yang_si',        '$2a$10$LjHOli0cEypdGtENHmg9y.hEpyWPr0zEyBS60g.cU2DwE8f5/6FEW', @t_now - INTERVAL 318 DAY),
(3458764513820541013, 'zhu_xi',         '$2a$10$jtWYHLoo2DcOahoNIQjXvuxPnXP9k09PE93YLMME/33MqZj1ORSNa', @t_now - INTERVAL 312 DAY),
(3458764513820541014, 'qin_wu',         '$2a$10$SrG0M/pC9eHkgkNiLXFLNuwnbT7imlTOsM.50iYy5PZhL/JtDNEki', @t_now - INTERVAL 305 DAY),
(3458764513820541015, 'you_liu',        '$2a$10$wspDbTJWmntUZcHrVwsTfeHnzO8UtkAo2D9n2rZ63ZcR0owrWMKeO', @t_now - INTERVAL 297 DAY),
(3458764513820541016, 'xu_qi',          '$2a$10$LjHOli0cEypdGtENHmg9y.hEpyWPr0zEyBS60g.cU2DwE8f5/6FEW', @t_now - INTERVAL 289 DAY),
(3458764513820541017, 'he_ba',          '$2a$10$jtWYHLoo2DcOahoNIQjXvuxPnXP9k09PE93YLMME/33MqZj1ORSNa', @t_now - INTERVAL 280 DAY),
(3458764513820541018, 'lv_jiu',         '$2a$10$SrG0M/pC9eHkgkNiLXFLNuwnbT7imlTOsM.50iYy5PZhL/JtDNEki', @t_now - INTERVAL 271 DAY),
(3458764513820541019, 'shi_shi',        '$2a$10$wspDbTJWmntUZcHrVwsTfeHnzO8UtkAo2D9n2rZ63ZcR0owrWMKeO', @t_now - INTERVAL 262 DAY),
(3458764513820541020, 'tang_yi',        '$2a$10$LjHOli0cEypdGtENHmg9y.hEpyWPr0zEyBS60g.cU2DwE8f5/6FEW', @t_now - INTERVAL 253 DAY),
(3458764513820541021, 'moyu_wang',      '$2a$10$jtWYHLoo2DcOahoNIQjXvuxPnXP9k09PE93YLMME/33MqZj1ORSNa', @t_now - INTERVAL 244 DAY),
(3458764513820541022, 'byte_dancer',    '$2a$10$SrG0M/pC9eHkgkNiLXFLNuwnbT7imlTOsM.50iYy5PZhL/JtDNEki', @t_now - INTERVAL 236 DAY),
(3458764513820541023, 'caffeine_max',   '$2a$10$wspDbTJWmntUZcHrVwsTfeHnzO8UtkAo2D9n2rZ63ZcR0owrWMKeO', @t_now - INTERVAL 228 DAY),
(3458764513820541024, 'midnight_push',  '$2a$10$LjHOli0cEypdGtENHmg9y.hEpyWPr0zEyBS60g.cU2DwE8f5/6FEW', @t_now - INTERVAL 220 DAY),
(3458764513820541025, 'rubber_duck',    '$2a$10$jtWYHLoo2DcOahoNIQjXvuxPnXP9k09PE93YLMME/33MqZj1ORSNa', @t_now - INTERVAL 212 DAY),
(3458764513820541026, 'null_pointer',   '$2a$10$SrG0M/pC9eHkgkNiLXFLNuwnbT7imlTOsM.50iYy5PZhL/JtDNEki', @t_now - INTERVAL 205 DAY),
(3458764513820541027, 'stack_overflow', '$2a$10$wspDbTJWmntUZcHrVwsTfeHnzO8UtkAo2D9n2rZ63ZcR0owrWMKeO', @t_now - INTERVAL 197 DAY),
(3458764513820541028, 'hello_world_go', '$2a$10$LjHOli0cEypdGtENHmg9y.hEpyWPr0zEyBS60g.cU2DwE8f5/6FEW', @t_now - INTERVAL 190 DAY),
(3458764513820541029, 'goroutine_leak', '$2a$10$jtWYHLoo2DcOahoNIQjXvuxPnXP9k09PE93YLMME/33MqZj1ORSNa', @t_now - INTERVAL 183 DAY),
(3458764513820541030, 'ctx_deadline',   '$2a$10$SrG0M/pC9eHkgkNiLXFLNuwnbT7imlTOsM.50iYy5PZhL/JtDNEki', @t_now - INTERVAL 176 DAY),
(3458764513820541031, 'mutex_holder',   '$2a$10$wspDbTJWmntUZcHrVwsTfeHnzO8UtkAo2D9n2rZ63ZcR0owrWMKeO', @t_now - INTERVAL 169 DAY),
(3458764513820541032, 'channel_buf',    '$2a$10$LjHOli0cEypdGtENHmg9y.hEpyWPr0zEyBS60g.cU2DwE8f5/6FEW', @t_now - INTERVAL 162 DAY),
(3458764513820541033, 'defer_stacker',  '$2a$10$jtWYHLoo2DcOahoNIQjXvuxPnXP9k09PE93YLMME/33MqZj1ORSNa', @t_now - INTERVAL 155 DAY),
(3458764513820541034, 'interface_nil',  '$2a$10$SrG0M/pC9eHkgkNiLXFLNuwnbT7imlTOsM.50iYy5PZhL/JtDNEki', @t_now - INTERVAL 148 DAY),
(3458764513820541035, 'slice_grower',   '$2a$10$wspDbTJWmntUZcHrVwsTfeHnzO8UtkAo2D9n2rZ63ZcR0owrWMKeO', @t_now - INTERVAL 141 DAY),
(3458764513820541036, 'map_racer',      '$2a$10$LjHOli0cEypdGtENHmg9y.hEpyWPr0zEyBS60g.cU2DwE8f5/6FEW', @t_now - INTERVAL 134 DAY),
(3458764513820541037, 'gc_watcher',     '$2a$10$jtWYHLoo2DcOahoNIQjXvuxPnXP9k09PE93YLMME/33MqZj1ORSNa', @t_now - INTERVAL 127 DAY),
(3458764513820541038, 'pprof_hunter',   '$2a$10$SrG0M/pC9eHkgkNiLXFLNuwnbT7imlTOsM.50iYy5PZhL/JtDNEki', @t_now - INTERVAL 120 DAY),
(3458764513820541039, 'trace_reader',   '$2a$10$wspDbTJWmntUZcHrVwsTfeHnzO8UtkAo2D9n2rZ63ZcR0owrWMKeO', @t_now - INTERVAL 113 DAY),
(3458764513820541040, 'bench_master',   '$2a$10$LjHOli0cEypdGtENHmg9y.hEpyWPr0zEyBS60g.cU2DwE8f5/6FEW', @t_now - INTERVAL 106 DAY),
(3458764513820541041, 'prod_incident',  '$2a$10$jtWYHLoo2DcOahoNIQjXvuxPnXP9k09PE93YLMME/33MqZj1ORSNa', @t_now - INTERVAL 99 DAY),
(3458764513820541042, 'oncall_hero',    '$2a$10$SrG0M/pC9eHkgkNiLXFLNuwnbT7imlTOsM.50iYy5PZhL/JtDNEki', @t_now - INTERVAL 92 DAY),
(3458764513820541043, 'yaml_indent',    '$2a$10$wspDbTJWmntUZcHrVwsTfeHnzO8UtkAo2D9n2rZ63ZcR0owrWMKeO', @t_now - INTERVAL 85 DAY),
(3458764513820541044, 'pod_crasher',    '$2a$10$LjHOli0cEypdGtENHmg9y.hEpyWPr0zEyBS60g.cU2DwE8f5/6FEW', @t_now - INTERVAL 78 DAY),
(3458764513820541045, 'svc_mesh_noob',  '$2a$10$jtWYHLoo2DcOahoNIQjXvuxPnXP9k09PE93YLMME/33MqZj1ORSNa', @t_now - INTERVAL 71 DAY),
(3458764513820541046, 'redis_ttl',      '$2a$10$SrG0M/pC9eHkgkNiLXFLNuwnbT7imlTOsM.50iYy5PZhL/JtDNEki', @t_now - INTERVAL 64 DAY),
(3458764513820541047, 'slow_log_man',   '$2a$10$wspDbTJWmntUZcHrVwsTfeHnzO8UtkAo2D9n2rZ63ZcR0owrWMKeO', @t_now - INTERVAL 57 DAY),
(3458764513820541048, 'index_builder',  '$2a$10$LjHOli0cEypdGtENHmg9y.hEpyWPr0zEyBS60g.cU2DwE8f5/6FEW', @t_now - INTERVAL 50 DAY),
(3458764513820541049, 'explain_plan',   '$2a$10$jtWYHLoo2DcOahoNIQjXvuxPnXP9k09PE93YLMME/33MqZj1ORSNa', @t_now - INTERVAL 43 DAY),
(3458764513820541050, 'tx_isolation',   '$2a$10$SrG0M/pC9eHkgkNiLXFLNuwnbT7imlTOsM.50iYy5PZhL/JtDNEki', @t_now - INTERVAL 36 DAY),
(3458764513820541051, 'vue_react_fan',  '$2a$10$wspDbTJWmntUZcHrVwsTfeHnzO8UtkAo2D9n2rZ63ZcR0owrWMKeO', @t_now - INTERVAL 30 DAY),
(3458764513820541052, 'css_battle',     '$2a$10$LjHOli0cEypdGtENHmg9y.hEpyWPr0zEyBS60g.cU2DwE8f5/6FEW', @t_now - INTERVAL 25 DAY),
(3458764513820541053, 'ts_strict',      '$2a$10$jtWYHLoo2DcOahoNIQjXvuxPnXP9k09PE93YLMME/33MqZj1ORSNa', @t_now - INTERVAL 20 DAY),
(3458764513820541054, 'vite_speed',     '$2a$10$SrG0M/pC9eHkgkNiLXFLNuwnbT7imlTOsM.50iYy5PZhL/JtDNEki', @t_now - INTERVAL 15 DAY),
(3458764513820541055, 'leetcode_daily', '$2a$10$wspDbTJWmntUZcHrVwsTfeHnzO8UtkAo2D9n2rZ63ZcR0owrWMKeO', @t_now - INTERVAL 11 DAY),
(3458764513820541056, 'dp_struggler',   '$2a$10$LjHOli0cEypdGtENHmg9y.hEpyWPr0zEyBS60g.cU2DwE8f5/6FEW', @t_now - INTERVAL 8 DAY),
(3458764513820541057, 'graph_walker',   '$2a$10$jtWYHLoo2DcOahoNIQjXvuxPnXP9k09PE93YLMME/33MqZj1ORSNa', @t_now - INTERVAL 5 DAY),
(3458764513820541058, '阿伟摸鱼',       '$2a$10$SrG0M/pC9eHkgkNiLXFLNuwnbT7imlTOsM.50iYy5PZhL/JtDNEki', @t_now - INTERVAL 3 DAY),
(3458764513820541059, '老王写代码',     '$2a$10$wspDbTJWmntUZcHrVwsTfeHnzO8UtkAo2D9n2rZ63ZcR0owrWMKeO', @t_now - INTERVAL 2 DAY),
(3458764513820541060, 'xiaomei_dev',    '$2a$10$LjHOli0cEypdGtENHmg9y.hEpyWPr0zEyBS60g.cU2DwE8f5/6FEW', @t_now - INTERVAL 1 DAY);

-- ----------------------------------------------------------------------------
-- 4. 版块(12 个)
-- ----------------------------------------------------------------------------
INSERT INTO `community` (community_id, community_name, introduction, create_time, update_time) VALUES
(1,  'Go',                 'Go 语言、并发模型与工程实践',            '2016-11-01 08:10:10', '2016-11-01 08:10:10'),
(2,  'C/C++',              '指针、内存管理与 STL',                  '2015-06-01 09:00:00', '2015-06-01 09:00:00'),
(3,  '篮球',               '投篮、战术与野球场经验',                 '2020-03-01 10:00:00', '2020-03-01 10:00:00'),
(4,  '足球',               '传接球、阵型与看球心得',                 '2020-03-01 10:10:00', '2020-03-01 10:10:00'),
(5,  'Rust',               '所有权、生命周期与无畏并发',             '2021-03-15 10:00:00', '2021-03-15 10:00:00'),
(6,  'Python',             '人生苦短, 我用 Python',                 '2017-05-20 14:20:00', '2017-05-20 14:20:00'),
(7,  'Java',               'Java 后端、JVM 与中间件',               '2016-09-01 09:00:00', '2016-09-01 09:00:00'),
(8,  'MySQL',              'SQL、索引、事务与性能优化',              '2019-04-12 11:30:00', '2019-04-12 11:30:00'),
(9,  'Linux 与云原生',      'Linux、容器与线上运维',                 '2020-07-08 15:45:00', '2020-07-08 15:45:00'),
(10, 'JavaScript 前端',     'JavaScript、浏览器与前端工程化',         '2018-11-11 10:10:00', '2018-11-11 10:10:00'),
(11, '算法与数据结构',       '从红黑树到一致性哈希',                   '2019-08-19 16:00:00', '2019-08-19 16:00:00'),
(12, '跑步与健身',          '跑步、力量训练与运动恢复',               '2020-02-02 20:00:00', '2020-02-02 20:00:00');

-- ----------------------------------------------------------------------------
-- 5. 帖子(120 篇)
--    时间是"距今多少分钟", 从 7 分钟前一直铺到一年前, 并且刻意让不同版块交错出现,
--    同一个版块既有一年内最老的帖子也有几小时前的新帖; 有几天会出现"同一天 2~3 篇"
--    的情况(用来验证用户主页按天分组的时间轴)。
--    作者在 60 个用户里轮换, 每人写 2 篇。
-- ----------------------------------------------------------------------------
INSERT INTO `post` (post_id, title, content, author_id, community_id, status, create_time) VALUES
(3458764513820600001, 'sync.Pool 在高并发接口里的实测收益', '把一个每秒要构造几万个临时 buffer 的接口改成 sync.Pool 复用之后, GC 的 STW 从 3ms 降到 0.4ms。注意 Pool 里的对象随时可能被 GC 清掉, 不能当缓存用。', 3458764513820541001, 1, 1, @t_now - INTERVAL 7 MINUTE),
(3458764513820600002, 'Vue3 的 shallowRef 什么时候必须用', '大数组只做整体替换时用 shallowRef 能省掉一层深度代理, 我们的两万行表格从卡顿变得流畅; 但改数组内部元素不会触发更新, 这点要提前和同事说清楚。', 3458764513820541002, 10, 1, @t_now - INTERVAL 26 MINUTE),
(3458764513820600003, '慢查询日志里最容易被忽略的三类 SQL', '隐式类型转换、order by 里带函数、以及 limit 偏移量很大的深分页, 我们线上 P99 从 800ms 降到 90ms 主要靠前两类。', 3458764513820541001, 8, 1, @t_now - INTERVAL 53 MINUTE),
(3458764513820600004, 'pandas 处理千万行时的内存控制', '读 CSV 时用 usecols 和 dtype 指定列类型, 内存能降到三分之一; 分块读加上及时 del, 比直接换机器划算得多。', 3458764513820541004, 6, 1, @t_now - INTERVAL 92 MINUTE),
(3458764513820600005, 'vector 扩容为什么会让指针失效', '扩容时会申请新的内存块并把元素搬过去, 原来的指针和迭代器全部失效; 知道容量不够就提前 reserve, 能省掉这中间的拷贝和风险。', 3458764513820541002, 2, 1, @t_now - INTERVAL 154 MINUTE),
(3458764513820600006, '为什么 Rc 不是 Send', 'Rc 的引用计数不是原子操作, 跨线程共享会让计数错乱, 所以多线程场景必须换 Arc, 代价是每次克隆多一次原子操作。', 3458764513820541006, 5, 1, @t_now - INTERVAL 233 MINUTE),
(3458764513820600007, '线程池参数里最容易设错的两个', '队列容量设成无界会让 maximumPoolSize 形同虚设; 拒绝策略用 CallerRunsPolicy 又可能把压力打回上游。这两个参数要放在一起调。', 3458764513820541007, 7, 1, @t_now - INTERVAL 311 MINUTE),
(3458764513820600008, 'requests 和 limits 不一致带来的坑', 'requests 给 0.5 核 limits 给 2 核, 结果是节点上塞进太多 Pod, 高负载时互相抢 CPU, 所有服务的延迟一起抖。', 3458764513820541001, 9, 1, @t_now - INTERVAL 402 MINUTE),
(3458764513820600009, '一致性哈希为什么要引入虚拟节点', '物理节点少的时候, 哈希环上的分布会非常不均匀; 每个节点配上百个虚拟节点之后, 负载偏差能压到 5% 以内。', 3458764513820541009, 11, 1, @t_now - INTERVAL 517 MINUTE),
(3458764513820600010, '第一次跑 5 公里应该怎么配速', '起步比目标配速慢 15 秒, 前两公里只用鼻子呼吸, 最后一点五公里再提速; 一上来就冲的人基本会在三公里处走回去。', 3458764513820541010, 12, 1, @t_now - INTERVAL 638 MINUTE),
(3458764513820600011, 'context.WithValue 不要用来传业务参数', 'WithValue 是给请求域元数据(比如 traceID)用的, 塞业务参数会让函数签名失去意义, 也不好做静态检查。', 3458764513820541011, 1, 1, @t_now - INTERVAL 745 MINUTE),
(3458764513820600012, 'vite 冷启动从 12 秒优化到 2 秒', '主要是把 barrel 文件(index.ts 里 export * 一大堆)拆掉, 另外把几个体积大的图表库改成异步组件, 启动时间立刻下来了。', 3458764513820541011, 10, 1, @t_now - INTERVAL 866 MINUTE),
(3458764513820600013, '投篮手型怎么固定下来', '手肘对准篮筐、出手后手腕自然下压, 每次投篮都保持同一个结束动作; 先在中距离空位练到连续命中 20 个, 再往三分线退。', 3458764513820541012, 3, 1, @t_now - INTERVAL 1005 MINUTE),
(3458764513820600014, '业余球队怎么练传接球', '两个人一组, 跑动中一脚出球、接球前先观察身后; 每周练二十分钟, 比赛里的失误会明显减少。', 3458764513820541014, 4, 1, @t_now - INTERVAL 1122 MINUTE),
(3458764513820600015, 'asyncio 里混进阻塞调用怎么排查', '用 debug=True 打开事件循环的慢回调日志, 再配合 py-spy dump 看栈, 基本十分钟就能定位到是哪个同步库在拖后腿。', 3458764513820541015, 6, 1, @t_now - INTERVAL 1287 MINUTE),
(3458764513820600016, '组合索引最左前缀的边界情况', '范围查询之后的列用不上索引, 但等值条件写在范围列后面依然有效, 这条规则值得在评审里反复强调。', 3458764513820541016, 8, 1, @t_now - INTERVAL 1403 MINUTE),
(3458764513820600017, '组件库按需引入到底省了多少', '一个中型项目里, 按需引入把打包体积从 1.8MB 降到 620KB, 首屏可用时间少了将近一秒, 但要注意样式文件也要一起按需。', 3458764513820541017, 10, 1, @t_now - INTERVAL 1591 MINUTE),
(3458764513820600018, 'pprof 火焰图应该怎么看', '先看宽度最大的几块(时间占比), 再顺着调用栈往下找自己的代码; 不要一上来就盯着最深的栈帧, 那往往是标准库。', 3458764513820541018, 1, 1, @t_now - INTERVAL 1720 MINUTE),
(3458764513820600019, 'malloc 和 new 的区别不只是要不要构造', 'malloc 只拿内存, new 还会调用构造函数; 反过来说 free 不会析构, delete 会。混用(比如 new 出来用 free)在带资源的类型上一定出问题。', 3458764513820541019, 2, 1, @t_now - INTERVAL 1886 MINUTE),
(3458764513820600020, '布隆过滤器的误判率怎么估', '误判率大概是 (1 - e^(-kn/m))^k, 工程上一般先定误判率再反推位数组大小, 别凭感觉拍一个 m。', 3458764513820541020, 11, 1, @t_now - INTERVAL 2015 MINUTE),
(3458764513820600021, 'Helm values 分环境管理的两种做法', '一种是一份 values 加多份 override, 另一种是每个环境一个完整文件。团队小的时候后者更省心, 前者更容易保证一致。', 3458764513820541021, 9, 1, @t_now - INTERVAL 2233 MINUTE),
(3458764513820600022, 'Option 链式处理让代码短了一半', 'map_or_else 加上 ok_or 之后, 一堆嵌套 match 变成了流水线, 可读性提升比想象中大, 但别把链子拉得太长。', 3458764513820541022, 5, 1, @t_now - INTERVAL 2401 MINUTE),
(3458764513820600023, 'Spring 事务失效的几种常见场景', '同类内部方法调用、方法不是 public、异常被自己 catch 掉、以及传播行为设成 NOT_SUPPORTED, 这几种最容易踩。', 3458764513820541023, 7, 1, @t_now - INTERVAL 2612 MINUTE),
(3458764513820600024, '跑步膝疼是姿势问题还是量的问题', '八成是量的问题: 周跑量涨得太快、跑前没热身、加上鞋没缓震; 先把每周增量控制在一成以内, 再谈姿势。', 3458764513820541024, 12, 1, @t_now - INTERVAL 2805 MINUTE),
(3458764513820600025, 'errgroup 比手写 WaitGroup 好在哪', '第一个错误就能取消整个 context, 不用自己写 err 通道和 cancel, 配合 WithContext 还能顺手控制超时。', 3458764513820541025, 1, 1, @t_now - INTERVAL 3006 MINUTE),
(3458764513820600026, '前端如何优雅地处理接口 401', '在 axios 响应拦截器里统一刷新 token 并重放请求, 用一个共享的 promise 保证并发只刷一次, 业务代码完全不用感知。', 3458764513820541026, 10, 1, @t_now - INTERVAL 3218 MINUTE),
(3458764513820600027, '深分页的三种优化方案', '记录上次的最大 id(游标)、延迟关联先查主键再回表、以及业务上限制最大页数, 前两种最通用。', 3458764513820541027, 8, 1, @t_now - INTERVAL 3440 MINUTE),
(3458764513820600028, '给项目加上 mypy 之后踩的坑', '第三方库缺类型存根的时候要么装 types-xxx 要么写 ignore, 否则 CI 会一直红; 建议先把严格模式关掉再逐步打开。', 3458764513820541028, 6, 1, @t_now - INTERVAL 3661 MINUTE),
(3458764513820600029, '指针和引用在接口设计里怎么选', '需要表达"一定不为空且不换对象"就用引用; 允许为空、可以重新指向别处就用指针。把这条规定写进代码规范, review 时能少吵很多架。', 3458764513820541029, 2, 1, @t_now - INTERVAL 3894 MINUTE),
(3458764513820600030, '生命周期标注什么时候可以省', '只有一个输入引用时输出引用自动跟随它, 不用手写; 一旦有两个以上输入引用, 编译器就没法猜了, 必须显式标注。', 3458764513820541030, 5, 1, @t_now - INTERVAL 4113 MINUTE),
(3458764513820600031, 'JVM 调优我一般只看这三个参数', '堆大小、新生代比例、以及 GC 类型(低延迟场景直接上 G1 或 ZGC), 其他参数在业务代码没写坏之前基本不用动。', 3458764513820541031, 7, 1, @t_now - INTERVAL 4360 MINUTE),
(3458764513820600032, 'Pod 一直 Pending 的排查顺序', '先 describe 看 Events, 大概率是资源不足、节点亲和性不满足、或者 PVC 没绑定; 顺序别搞反, 不然会绕很久。', 3458764513820541032, 9, 1, @t_now - INTERVAL 4588 MINUTE),
(3458764513820600033, '红黑树和 AVL 树的取舍', 'AVL 更平衡所以查得快, 红黑树旋转少所以插入删除便宜; 工程里读多写少用 AVL, 通用容器基本都是红黑树。', 3458764513820541033, 11, 1, @t_now - INTERVAL 4802 MINUTE),
(3458764513820600034, '力量训练一周练几次比较合适', '新手隔天练一次就够了, 同一肌群至少间隔 48 小时; 练得太勤恢复不过来, 重量反而上不去。', 3458764513820541034, 12, 1, @t_now - INTERVAL 5031 MINUTE),
(3458764513820600035, '运球过人: 先练变向还是先练背后运球', '先把体前变向练到左右手都能压住球, 再练背后和胯下; 顺序反了很容易变成动作很花但是过不了人。', 3458764513820541035, 3, 1, @t_now - INTERVAL 5273 MINUTE),
(3458764513820600036, 'pinia 和 vuex 到底怎么选', '新项目直接 pinia, 组合式写法加完整类型推导; 老项目如果 vuex 用得很稳, 没必要为了新而换。', 3458764513820541036, 10, 1, @t_now - INTERVAL 5510 MINUTE),
(3458764513820600037, '挡拆之后的三分出手时机', '掩护人还没站定就出手基本是白投, 等防守人绕过来那一瞬间才是真正的空位; 如果对方选择换防, 就直接打错位, 不要勉强出手。', 3458764513820541037, 3, 1, @t_now - INTERVAL 5762 MINUTE),
(3458764513820600038, '442 和 433 阵型的区别', '442 更强调两条线的紧凑和边路传中, 433 靠中场三人的跑动覆盖; 业余球队体能有限, 442 通常更好执行。', 3458764513820541038, 4, 1, @t_now - INTERVAL 6003 MINUTE),
(3458764513820600039, 'virtualenv 和 conda 怎么选', '纯 Python 项目用 venv 就够了; 需要装 CUDA、GDAL 这类带系统依赖的包, conda 省事很多。', 3458764513820541039, 6, 1, @t_now - INTERVAL 6255 MINUTE),
(3458764513820600040, '事务隔离级别与幻读的真实关系', 'MySQL 的 RR 用间隙锁挡住了大部分幻读, 但快照读和当前读混用时结论会不一样, 面试和线上都要分清楚。', 3458764513820541040, 8, 1, @t_now - INTERVAL 6512 MINUTE),
(3458764513820600041, '表格虚拟滚动的实现要点', '固定行高可以直接算偏移, 不定行高就得维护前缀高度数组; 无论哪种, 都要处理滚动时白屏那一帧。', 3458764513820541041, 10, 1, @t_now - INTERVAL 6777 MINUTE),
(3458764513820600042, '一次线上 goroutine 泄漏的复盘', '上游超时后我们没有取消下游请求, 导致每次调用都留下一个阻塞的 goroutine, 一天涨到十万。修法是把 context 一路传下去。', 3458764513820541042, 1, 1, @t_now - INTERVAL 7040 MINUTE),
(3458764513820600043, 'unique_ptr 与 shared_ptr 的使用边界', '默认用 unique_ptr, 只有确实需要多处共享所有权时才升级成 shared_ptr; shared_ptr 的引用计数是原子操作, 在热点路径上并不便宜。', 3458764513820541043, 2, 1, @t_now - INTERVAL 7311 MINUTE),
(3458764513820600044, 'TopK 问题为什么用堆而不是排序', 'n 很大 k 很小时, 维护一个大小为 k 的堆是 O(n log k), 而全排序是 O(n log n), 数据量大时差距很明显。', 3458764513820541044, 11, 1, @t_now - INTERVAL 7588 MINUTE),
(3458764513820600045, '探针配置不当把服务打挂的经历', 'liveness 探针的超时和失败阈值设得太激进, 一次 GC 抖动就被判定为不健康, 然后 Pod 被反复重启, 雪崩。', 3458764513820541045, 9, 1, @t_now - INTERVAL 7869 MINUTE),
(3458764513820600046, '所有权规则帮我避免的一次 bug', '把可变引用借出去之后再读原变量, 编译器直接报错; 换到别的语言里, 这就是一个很难复现的数据竞争。', 3458764513820541046, 5, 1, @t_now - INTERVAL 8152 MINUTE),
(3458764513820600047, 'HashMap 扩容对性能的影响', '初始容量没设好会触发多次扩容和 rehash, 已知数据量时直接按负载因子反推初始容量, 能省掉这部分抖动。', 3458764513820541047, 7, 1, @t_now - INTERVAL 8441 MINUTE),
(3458764513820600048, '晨跑和夜跑哪个更容易坚持', '晨跑打卡率高但要早起, 夜跑方便却容易被加班冲掉; 关键是固定一个时间点, 别今天早明天晚。', 3458764513820541048, 12, 1, @t_now - INTERVAL 8733 MINUTE),
(3458764513820600049, '接口返回 nil 切片引发的坑', 'nil 切片序列化成 null, 空切片序列化成 [], 前端如果不做兼容就会在 forEach 上报错, 统一返回空切片更省事。', 3458764513820541049, 1, 1, @t_now - INTERVAL 9028 MINUTE),
(3458764513820600050, '图片懒加载踩过的坑', '用 IntersectionObserver 时忘了处理元素已经进入视口的情况, 结果首屏图片一直不加载; 另外要记得 unobserve 释放。', 3458764513820541050, 10, 1, @t_now - INTERVAL 9325 MINUTE),
(3458764513820600051, 'count(*) 和 count(1) 有区别吗', 'InnoDB 里两者都是统计行数, 优化器处理方式一致, 别再纠结这个; 真正要关心的是能不能用上更小的索引。', 3458764513820541051, 8, 1, @t_now - INTERVAL 9712 MINUTE),
(3458764513820600052, 'requirements.txt 锁版本的正确姿势', '用 pip-compile 从 requirements.in 生成带完整依赖树的锁文件, 直接 pip freeze 会把间接依赖也写进去, 升级时很难受。', 3458764513820541052, 6, 1, @t_now - INTERVAL 10450 MINUTE),
(3458764513820600053, 'memcpy 与 memmove 到底差在哪', 'memcpy 假设源和目标不重叠, 重叠时行为未定义; memmove 会先判断方向再拷贝, 所以更慢一点但更安全。自己写内存搬移时优先用 memmove。', 3458764513820541053, 2, 1, @t_now - INTERVAL 11203 MINUTE),
(3458764513820600054, 'Rust 智能指针选型速查', '独占用 Box, 单线程共享用 Rc, 多线程共享用 Arc, 需要内部可变再套 RefCell 或 Mutex, 先想清楚共享范围再动手。', 3458764513820541054, 5, 1, @t_now - INTERVAL 12061 MINUTE),
(3458764513820600055, 'JVM 内存模型与 volatile', 'volatile 保证可见性和有序性但不保证原子性, i++ 这种复合操作还是要靠锁或原子类。', 3458764513820541055, 7, 1, @t_now - INTERVAL 13012 MINUTE),
(3458764513820600056, 'HPA 按 CPU 扩容为什么总是滞后', 'CPU 指标本身就是滞后指标, 再加上采集和扩容的延迟, 突发流量根本来不及; 关键服务建议用 QPS 这类自定义指标。', 3458764513820541056, 9, 1, @t_now - INTERVAL 14207 MINUTE),
(3458764513820600057, 'KMP 的 next 数组怎么推导', '先把前缀函数定义成"最长相等前后缀长度", 再用两个指针自己匹配自己, 想清楚回退那一步之后就很好写了。', 3458764513820541057, 11, 1, @t_now - INTERVAL 15530 MINUTE),
(3458764513820600058, '增肌期要不要做有氧', '要, 每周两三次低强度有氧能维持心肺和食欲; 但别和力量训练挨着做, 会影响恢复。', 3458764513820541058, 12, 1, @t_now - INTERVAL 16902 MINUTE),
(3458764513820600059, 'defer 与命名返回值的执行顺序', 'return 先把返回值赋给命名返回值, 再执行 defer, 所以 defer 里能改到最终返回值; 这个特性用好了很优雅, 用不好会让人看不懂。', 3458764513820541059, 1, 1, @t_now - INTERVAL 18455 MINUTE),
(3458764513820600060, 'provide/inject 的响应式陷阱', '注入一个解构出来的普通对象会丢失响应式, 要注入 ref 本身用 toRefs 包一层, 否则子组件永远不会更新。', 3458764513820541060, 10, 1, @t_now - INTERVAL 20130 MINUTE),
(3458764513820600061, '野球场防守站位的三个原则', '永远站在进攻人和篮筐之间、保持一臂距离、重心比进攻人低; 做到这三条, 就算脚步慢一点也不会被轻易过掉。', 3458764513820541001, 3, 1, @t_now - INTERVAL 21904 MINUTE),
(3458764513820600062, '边路突破之后该传中还是内切', '先看禁区里有几个人, 只有一个人就别传中; 内切能带走防守人给边后卫让出下底的空间, 两种选择要交替用。', 3458764513820541002, 4, 1, @t_now - INTERVAL 23876 MINUTE),
(3458764513820600063, '日志里为什么要带上 trace 上下文', '一次请求跨好几个服务时, 没有 traceID 只能靠时间戳猜; 加上之后排查时间经常能从半小时压到五分钟。', 3458764513820541003, 6, 1, @t_now - INTERVAL 25990 MINUTE),
(3458764513820600064, '主从延迟怎么看与怎么缓解', '先看 Seconds_Behind_Master 和 GTID 差距, 大事务是最常见的原因; 拆分大事务加并行复制之后基本能接受。', 3458764513820541004, 8, 1, @t_now - INTERVAL 28214 MINUTE),
(3458764513820600065, '响应式数据里不要放 class 实例', 'Vue 会把实例的属性递归代理一遍, 方法里的 this 也会错乱; 要么用 markRaw 标记, 要么存普通对象。', 3458764513820541005, 10, 1, @t_now - INTERVAL 30661 MINUTE),
(3458764513820600066, '接口设计: 传结构体还是传参数', '参数超过四个或者有明显语义分组时就应该换成结构体, 否则调用方永远记不住顺序。', 3458764513820541006, 1, 1, @t_now - INTERVAL 33245 MINUTE),
(3458764513820600067, 'const 修饰指针的四种写法', 'const 在星号左边修饰的是指向的内容, 在右边修饰的是指针本身; 写 const int * const p 的时候两样都不能改, 记住位置比记口诀靠谱。', 3458764513820541007, 2, 1, @t_now - INTERVAL 36012 MINUTE),
(3458764513820600068, '跳表为什么能替代平衡树', '实现简单、并发友好, 范围查询也方便; Redis 的有序集合就是跳表加哈希表。', 3458764513820541008, 11, 1, @t_now - INTERVAL 38990 MINUTE),
(3458764513820600069, 'ConfigMap 更新后应用没生效', '挂载成文件会自动更新, 注入成环境变量不会; 很多框架还只在启动时读一次配置, 改完记得触发一次滚动重启。', 3458764513820541009, 9, 1, @t_now - INTERVAL 42135 MINUTE),
(3458764513820600070, 'Cargo workspace 拆分经验', '按"能被独立复用"来拆, 而不是按目录结构拆; 拆得太碎之后编译时间反而变长。', 3458764513820541010, 5, 1, @t_now - INTERVAL 45503 MINUTE),
(3458764513820600071, 'CompletableFuture 的异常处理', 'thenApply 里抛异常会直接短路, 想要兜底得用 handle 或 exceptionally, 否则错误会静默丢掉。', 3458764513820541011, 7, 1, @t_now - INTERVAL 49012 MINUTE),
(3458764513820600072, '跑鞋该怎么选: 缓震还是支撑', '先看自己的足弓和落地方式, 扁平足优先考虑支撑型; 跑量不大的时候, 尺码合脚比参数重要。', 3458764513820541012, 12, 1, @t_now - INTERVAL 52770 MINUTE),
(3458764513820600073, '跟腱不舒服还能不能打球', '只要不是撕裂就先停球两周, 换成骑车和力量训练维持状态; 疼痛消失后也要从半场开始, 别一上来就全速跑跳。', 3458764513820541013, 3, 1, @t_now - INTERVAL 56715 MINUTE),
(3458764513820600074, '组合式 API 里怎么组织请求逻辑', '把 loading、error、data 和重试封成一个 useRequest, 页面里只留业务分支, 比在每个组件里手写一遍靠谱。', 3458764513820541014, 10, 1, @t_now - INTERVAL 60990 MINUTE),
(3458764513820600075, 'online DDL 与锁的取舍', '加索引基本不锁表, 但改列类型、加非空默认值在老版本上会重建表; 上线前先在同数据量的预发环境验证。', 3458764513820541015, 8, 1, @t_now - INTERVAL 65435 MINUTE),
(3458764513820600076, '用 FastAPI 起一个内部服务', 'pydantic 做入参校验、依赖注入管理连接池, 半天就能写完一个带文档的内部工具, 很适合做运维小平台。', 3458764513820541016, 6, 1, @t_now - INTERVAL 70112 MINUTE),
(3458764513820600077, '结构体内存对齐与 padding', '编译器会按最大成员的对齐要求补空洞, 所以结构体大小往往不等于字段之和; 把大字段放前面、小的凑一起, 能省下不少内存。', 3458764513820541017, 2, 1, @t_now - INTERVAL 75060 MINUTE),
(3458764513820600078, '错误处理用 thiserror 还是 anyhow', '库对外暴露的错误用 thiserror 定义具体类型, 应用层内部流程用 anyhow 加 context 更省事。', 3458764513820541018, 5, 1, @t_now - INTERVAL 80235 MINUTE),
(3458764513820600079, 'MyBatis 缓存要不要开', '一级缓存默认开着, 同一次会话里可能读到旧数据; 二级缓存跨会话共享, 分布式环境下不建议开。', 3458764513820541019, 7, 1, @t_now - INTERVAL 85635 MINUTE),
(3458764513820600080, 'Ingress 与 Gateway API 的差异', 'Ingress 只解决了最简单的七层转发, Gateway API 把角色拆开、表达能力更强, 但迁移成本要提前评估。', 3458764513820541020, 9, 1, @t_now - INTERVAL 91310 MINUTE),
(3458764513820600081, '动态规划的状态定义技巧', '状态尽量设计成"处理到第 i 个位置且满足某条件"的形态, 这样转移方程通常只有两三种, 不用硬凑。', 3458764513820541021, 11, 1, @t_now - INTERVAL 97270 MINUTE),
(3458764513820600082, '跑前热身和跑后拉伸怎么做', '跑前做动态热身(高抬腿、开合跳)五分钟, 跑后做静态拉伸; 反过来做容易拉伤。', 3458764513820541022, 12, 1, @t_now - INTERVAL 103512 MINUTE),
(3458764513820600083, '切片扩容规则与预分配', '容量不足时按 1.25 倍左右增长(小切片翻倍), 已知长度就直接 make 带容量, 能省掉多次拷贝。', 3458764513820541023, 1, 1, @t_now - INTERVAL 110030 MINUTE),
(3458764513820600084, '自定义指令实现按钮级权限', '把权限校验写在指令的 mounted 里直接移除节点, 比在每个按钮外面套一层 v-if 清爽很多。', 3458764513820541024, 10, 1, @t_now - INTERVAL 116855 MINUTE),
(3458764513820600085, '三对三和五对五的节奏差异', '三对三空间大、回合快, 无球掩护和挡拆比单打更有效; 五对五要先落位再推进, 不然很容易变成四个人看一个人打。', 3458764513820541025, 3, 1, @t_now - INTERVAL 123990 MINUTE),
(3458764513820600086, '看球时应该关注哪些细节', '别看球在哪里, 看没有球的人在跑什么位置; 一次进攻的成败往往在接球之前就决定了。', 3458764513820541026, 4, 1, @t_now - INTERVAL 131412 MINUTE),
(3458764513820600087, '装饰器保留原函数签名的细节', '不加 functools.wraps 会让文档工具和类型检查都失效, 顺带把函数元信息丢掉, 属于必写的两行。', 3458764513820541027, 6, 1, @t_now - INTERVAL 139115 MINUTE),
(3458764513820600088, '在线加索引的注意事项', '低峰期执行、注意磁盘余量、大表分多次加; 加完记得 analyze table, 否则优化器可能还在用旧统计信息。', 3458764513820541028, 8, 1, @t_now - INTERVAL 147120 MINUTE),
(3458764513820600089, '单元测试从哪几个组件开始写', '先覆盖纯函数和格式化工具, 再写表单校验和请求封装, 最后才是需要挂载的复杂组件, 收益递减。', 3458764513820541029, 10, 1, @t_now - INTERVAL 155440 MINUTE),
(3458764513820600090, 'singleflight 合并重复请求', '同一个 key 的并发请求只放一个到下游, 其余的等结果; 用在缓存击穿的场景上效果最明显。', 3458764513820541030, 1, 1, @t_now - INTERVAL 164090 MINUTE),
(3458764513820600091, 'constexpr 能做哪些编译期计算', '常量表达式、编译期数组大小、模板参数都能用 constexpr; 但它只是"尽量在编译期算", 运行期调用的重载还是可能走到, 别把它当成零成本保证。', 3458764513820541031, 2, 1, @t_now - INTERVAL 173010 MINUTE),
(3458764513820600092, '哈希冲突的三种解决方式', '链地址法、开放寻址、再加一层哈希, 工程里最常见的是第一种; 开放寻址对缓存更友好但删除比较麻烦。', 3458764513820541032, 11, 1, @t_now - INTERVAL 182220 MINUTE),
(3458764513820600093, '命名空间与资源配额的实践', '给每个团队一个 namespace 并配上 ResourceQuota, 能有效防止某个服务把整个集群的资源吃光。', 3458764513820541033, 9, 1, @t_now - INTERVAL 191700 MINUTE),
(3458764513820600094, '闭包捕获与 move 关键字', '不加 move 时闭包按引用捕获, 生命周期受限于外部变量; 交给线程时必须 move, 否则编译器直接拦下来。', 3458764513820541034, 5, 1, @t_now - INTERVAL 201460 MINUTE),
(3458764513820600095, '接口幂等性的几种实现', '唯一索引、状态机、以及基于 token 的防重表, 三种可以叠加使用, 支付类接口基本都是这么做的。', 3458764513820541035, 7, 1, @t_now - INTERVAL 211510 MINUTE),
(3458764513820600096, '久坐的人怎么安排一天的训练', '早上十分钟拉伸激活, 中午散步二十分钟, 晚上力量训练四十分钟; 比一次练两小时更容易坚持。', 3458764513820541036, 12, 1, @t_now - INTERVAL 221850 MINUTE),
(3458764513820600097, 'time.Ticker 不 Stop 的后果', '占用一个定时器并且让 goroutine 无法回收, 长时间运行的服务里会慢慢堆积; defer ticker.Stop() 是标配。', 3458764513820541037, 1, 1, @t_now - INTERVAL 232490 MINUTE),
(3458764513820600098, '路由守卫里做权限校验', '全局守卫里判断登录态和角色, 页面级 meta 里声明需要的权限, 按钮级的再配合自定义指令, 三层各管一段。', 3458764513820541038, 10, 1, @t_now - INTERVAL 243440 MINUTE),
(3458764513820600099, 'explain 里最重要的三列', 'type、rows、以及 Extra; 出现 ALL 和 Using filesort 就该警惕, Using index condition 通常是个好信号。', 3458764513820541039, 8, 1, @t_now - INTERVAL 254690 MINUTE),
(3458764513820600100, '生成器与列表推导的内存差异', '列表推导会一次性把结果全部放进内存, 生成器只在迭代时产出, 处理大文件时差别非常明显。', 3458764513820541040, 6, 1, @t_now - INTERVAL 266250 MINUTE),
(3458764513820600101, 'STL 迭代器失效的几种场景', 'vector 扩容、erase 之后、以及 map 的 erase 都会让某些迭代器失效; 改成用返回值接住新的迭代器, 或者先用下标再操作, 是最省心的写法。', 3458764513820541041, 2, 1, @t_now - INTERVAL 278120 MINUTE),
(3458764513820600102, 'trait object 与泛型的取舍', '泛型是编译期单态化, 没有运行时开销但会让二进制变大; trait object 走动态分发, 适合插件式设计。', 3458764513820541042, 5, 1, @t_now - INTERVAL 290290 MINUTE),
(3458764513820600103, '日志级别与异步 Appender', '高并发下同步写盘会阻塞业务线程, 换异步 Appender 并设置丢弃策略之后, 吞吐提升很明显。', 3458764513820541043, 7, 1, @t_now - INTERVAL 302740 MINUTE),
(3458764513820600104, '本地开发用 kind 还是 minikube', '只看调度和网络行为用 kind 更轻, 需要 dashboard 和持久卷的完整链路就用 minikube。', 3458764513820541044, 9, 1, @t_now - INTERVAL 315470 MINUTE),
(3458764513820600105, '前缀和与差分数组', '前缀和解决区间求和, 差分解决区间加; 两者互为逆运算, 看到"多次区间修改最后一次查询"就该想到差分。', 3458764513820541045, 11, 1, @t_now - INTERVAL 328460 MINUTE),
(3458764513820600106, '跑步心率区间怎么算', '最大心率可以用 220 减年龄估算, 有氧区间大概在 65%~75%; 更实用的验证方法是说话法: 能完整说出句子就说明强度合适。', 3458764513820541046, 12, 1, @t_now - INTERVAL 341690 MINUTE),
(3458764513820600107, '看 NBA 学战术动作真的有用吗', '能学思路但别照搬细节, 业余场地的身体条件和吹罚尺度完全不同; 学跑位和传球时机, 比学某个球星的动作更有用。', 3458764513820541047, 3, 1, @t_now - INTERVAL 355150 MINUTE),
(3458764513820600108, '大表单的状态管理', '字段超过三十个之后, 用 schema 驱动渲染加统一校验, 比一个个写 v-model 和 rules 好维护得多。', 3458764513820541048, 10, 1, @t_now - INTERVAL 368820 MINUTE),
(3458764513820600109, '和比自己强的对手打球怎么找位置', '少持球多跑动, 把注意力放在防守和篮板上; 接到球就传出去再切进来, 比硬打成功率高得多。', 3458764513820541049, 3, 1, @t_now - INTERVAL 382680 MINUTE),
(3458764513820600110, '五人制足球的跑位要点', '场地小、触球次数多, 所以要多做交叉跑动和撞墙配合; 站着等球基本等于放弃进攻。', 3458764513820541050, 4, 1, @t_now - INTERVAL 396780 MINUTE),
(3458764513820600111, '写一个计时的上下文管理器', '用 contextmanager 装饰器加 try/finally, 既能统计耗时也能保证异常时打印; 比在函数里手动记时间干净。', 3458764513820541051, 6, 1, @t_now - INTERVAL 411120 MINUTE),
(3458764513820600112, '删数据的正确姿势', '先加软删标记观察一个版本, 确认没人查询之后再物理删除; 大批量删除要分批并在低峰期执行, 避免长事务。', 3458764513820541052, 8, 1, @t_now - INTERVAL 425700 MINUTE),
(3458764513820600113, '骨架屏怎么做才不闪', '骨架的尺寸和真实内容保持一致、用同一个圆角和间距, 否则数据回来的一瞬间会有明显的跳动。', 3458764513820541053, 10, 1, @t_now - INTERVAL 440500 MINUTE),
(3458764513820600114, '结构体比较的注意事项', '结构体可比较的前提是所有字段都可比较, 含切片或 map 时编译不过; 这时候应该自己写 Equal 方法。', 3458764513820541054, 1, 1, @t_now - INTERVAL 455540 MINUTE),
(3458764513820600115, '内联函数为什么不能递归', '内联是把函数体摊到调用点, 递归展开会无限膨胀, 所以编译器遇到递归会直接放弃内联; 想内联的工具函数最好写成没有循环的小函数。', 3458764513820541055, 2, 1, @t_now - INTERVAL 470820 MINUTE),
(3458764513820600116, '复杂度分析的常见误区', '只数循环层数不看实际执行次数、忽略哈希的均摊、以及把递归的空间算漏, 这三个最容易在面试里翻车。', 3458764513820541056, 11, 1, @t_now - INTERVAL 486350 MINUTE),
(3458764513820600117, '发布策略: 蓝绿与金丝雀', '蓝绿切换快但资源翻倍, 金丝雀多一步流量控制但更稳; 有状态服务尽量别用蓝绿。', 3458764513820541057, 9, 1, @t_now - INTERVAL 502120 MINUTE),
(3458764513820600118, '迭代器适配器的惰性求值', 'map、filter 都是惰性的, 真正触发计算的是最后的 collect 或 for 循环; 不理解这一点很容易把性能优化写反。', 3458764513820541058, 5, 1, @t_now - INTERVAL 518140 MINUTE),
(3458764513820600119, '类加载顺序与静态初始化', '父类静态块、子类静态块、再是实例初始化和构造方法; 静态块里做耗时操作会把启动时间拖得很长。', 3458764513820541059, 7, 1, @t_now - INTERVAL 534400 MINUTE),
(3458764513820600120, '坚持跑步一年之后的真实变化', '静息心率从 72 降到 58, 睡眠明显变好, 体重只掉了四公斤; 变化最大的其实是情绪, 不是身材。', 3458764513820541060, 12, 1, @t_now - INTERVAL 545000 MINUTE);

-- 帖子的 update_time 和 create_time 保持一致(插入时 update_time 取的是当前时间)
UPDATE `post` SET `update_time` = `create_time`;

-- ----------------------------------------------------------------------------
-- 6. 评论(300 条, 其中一部分是楼中楼回复)
--    时间都晚于对应的帖子(写成"距今多少分钟", 所以数值比帖子的更小);
--    parent_id 指向同一个帖子下已经存在的评论, 前端才能正确挂到楼中楼里。
-- ----------------------------------------------------------------------------
INSERT INTO `comment` (comment_id, content, post_id, author_id, parent_id, create_time) VALUES
(3458764513820700001, '这个场景我们也在跑, 收益确实明显, 但要注意对象里有指针的时候别忘了重置', 3458764513820600001, 3458764513820541012, 0, @t_now - INTERVAL 4 MINUTE),
(3458764513820700002, '沙发, 正想找这方面的数据', 3458764513820600002, 3458764513820541009, 0, @t_now - INTERVAL 23 MINUTE),
(3458764513820700003, '请问 shallowRef 配合 watch 监听内部变化是不是就失效了', 3458764513820600002, 3458764513820541031, 3458764513820700002, @t_now - INTERVAL 12 MINUTE),
(3458764513820700004, '隐式类型转换那个我们排查了整整一下午, 字段是 varchar 结果传了数字', 3458764513820600003, 3458764513820541015, 0, @t_now - INTERVAL 49 MINUTE),
(3458764513820700005, '深分页那个建议再加一个业务上限制最大页数, 双保险', 3458764513820600003, 3458764513820541042, 0, @t_now - INTERVAL 33 MINUTE),
(3458764513820700006, 'dtype 指定成 category 也能省不少内存, 适合低基数的列', 3458764513820600004, 3458764513820541024, 0, @t_now - INTERVAL 88 MINUTE),
(3458764513820700007, '分块读的时候要注意最后一块要单独处理, 很容易漏掉', 3458764513820600004, 3458764513820541037, 0, @t_now - INTERVAL 71 MINUTE),
(3458764513820700008, '回复楼上: 对, 我们现在统一用 chunksize 加最后补齐的方式', 3458764513820600004, 3458764513820541004, 3458764513820700007, @t_now - INTERVAL 54 MINUTE),
(3458764513820700009, 'reserve 这一条太重要了, 我们一个循环里 append 十万次, 加上预分配快了一倍', 3458764513820600005, 3458764513820541056, 0, @t_now - INTERVAL 149 MINUTE),
(3458764513820700010, '还有一点是扩容之后原来的引用也会失效, 尤其是把引用存到成员变量里的场景', 3458764513820600005, 3458764513820541018, 0, @t_now - INTERVAL 122 MINUTE),
(3458764513820700011, '回复楼上: 对, 存引用比存指针更容易忘, 我们用 shared_ptr 反而安全一些', 3458764513820600005, 3458764513820541029, 3458764513820700010, @t_now - INTERVAL 96 MINUTE),
(3458764513820700012, 'Arc 的克隆是原子操作, 热点路径上频繁克隆还是会有开销', 3458764513820600006, 3458764513820541054, 0, @t_now - INTERVAL 228 MINUTE),
(3458764513820700013, '所以能用引用传参的地方就别克隆, 这条在 Rust 里特别重要', 3458764513820600006, 3458764513820541010, 0, @t_now - INTERVAL 190 MINUTE),
(3458764513820700014, 'RefCell 加 Rc 在单线程里能实现内部可变性, 但也容易把借用检查推迟到运行时', 3458764513820600006, 3458764513820541046, 0, @t_now - INTERVAL 141 MINUTE),
(3458764513820700015, '我们用 CallerRunsPolicy 之后上游开始超时, 后来改成自定义策略加降级', 3458764513820600007, 3458764513820541043, 0, @t_now - INTERVAL 305 MINUTE),
(3458764513820700016, 'corePoolSize 也要结合 QPS 和单任务耗时算, 不能拍脑袋', 3458764513820600007, 3458764513820541020, 0, @t_now - INTERVAL 262 MINUTE),
(3458764513820700017, '回复楼上: 我们是按 目标QPS 乘以 平均耗时 再留一点余量', 3458764513820600007, 3458764513820541007, 3458764513820700016, @t_now - INTERVAL 205 MINUTE),
(3458764513820700018, 'requests 太小还会导致调度器把它当成小任务, 频繁触发驱逐', 3458764513820600008, 3458764513820541044, 0, @t_now - INTERVAL 396 MINUTE),
(3458764513820700019, '建议再加上 LimitRange 兜底, 防止有人不写 requests', 3458764513820600008, 3458764513820541016, 0, @t_now - INTERVAL 341 MINUTE),
(3458764513820700020, '我们踩过的坑是 requests 给得比实际用量大很多, 集群利用率上不去', 3458764513820600008, 3458764513820541021, 0, @t_now - INTERVAL 268 MINUTE),
(3458764513820700021, '虚拟节点数量也不是越多越好, 太多会让内存里的环变大', 3458764513820600009, 3458764513820541019, 0, @t_now - INTERVAL 508 MINUTE),
(3458764513820700022, '这个结论和 Redis Cluster 的 16384 个槽位其实是一个思路', 3458764513820600009, 3458764513820541031, 0, @t_now - INTERVAL 432 MINUTE),
(3458764513820700023, '回复楼上: 对, 都是为了让分布更均匀', 3458764513820600009, 3458764513820541009, 3458764513820700021, @t_now - INTERVAL 349 MINUTE),
(3458764513820700024, '鼻子呼吸这个提示很实用, 一喘不上气就说明前面冲快了', 3458764513820600010, 3458764513820541025, 0, @t_now - INTERVAL 630 MINUTE),
(3458764513820700025, '第一次跑别追求配速, 能不停下来走完就算成功', 3458764513820600010, 3458764513820541052, 0, @t_now - INTERVAL 552 MINUTE),
(3458764513820700026, '最后一点五公里提速这个安排很像比赛里的负分割, 效果确实好', 3458764513820600010, 3458764513820541047, 0, @t_now - INTERVAL 441 MINUTE),
(3458764513820700027, 'traceID 配上日志采集, 排查跨服务问题效率提升非常明显', 3458764513820600011, 3458764513820541042, 0, @t_now - INTERVAL 738 MINUTE),
(3458764513820700028, '注意不要用字符串拼接 traceID, 直接用日志库的字段', 3458764513820600011, 3458764513820541011, 0, @t_now - INTERVAL 651 MINUTE),
(3458764513820700029, 'barrel 文件确实是重灾区, 我们项目里有三千行的 index.ts', 3458764513820600012, 3458764513820541054, 0, @t_now - INTERVAL 858 MINUTE),
(3458764513820700030, '异步组件加上路由懒加载, 首屏能再省几百毫秒', 3458764513820600012, 3458764513820541051, 0, @t_now - INTERVAL 762 MINUTE),
(3458764513820700031, '回复楼上: 对, 我们是两个一起做的', 3458764513820600012, 3458764513820541002, 3458764513820700030, @t_now - INTERVAL 611 MINUTE),
(3458764513820700032, '手型的关键其实是出手那一下的跟随动作, 很多人投完手就撤了', 3458764513820600013, 3458764513820541037, 0, @t_now - INTERVAL 997 MINUTE),
(3458764513820700033, '我一般先在篮下把动作定型, 再一步一步往后退, 效果比直接投三分好', 3458764513820600013, 3458764513820541050, 0, @t_now - INTERVAL 880 MINUTE),
(3458764513820700034, '接球前的观察习惯最难养成, 但作用最大', 3458764513820600014, 3458764513820541026, 0, @t_now - INTERVAL 1113 MINUTE),
(3458764513820700035, '我们队练过一段时间三角传球, 比赛里出球速度快了很多', 3458764513820600014, 3458764513820541014, 0, @t_now - INTERVAL 981 MINUTE),
(3458764513820700036, 'py-spy 真是神器, 线上不用改代码就能看栈', 3458764513820600015, 3458764513820541015, 0, @t_now - INTERVAL 1278 MINUTE),
(3458764513820700037, '另外 asyncio 里千万别直接调 requests, 必须换 httpx 的异步客户端', 3458764513820600015, 3458764513820541040, 0, @t_now - INTERVAL 1121 MINUTE),
(3458764513820700038, '我们还遇到过 logging 里做 IO 导致事件循环卡住', 3458764513820600015, 3458764513820541028, 0, @t_now - INTERVAL 916 MINUTE),
(3458764513820700039, '回复楼上: 日志这块建议换成异步 handler', 3458764513820600015, 3458764513820541005, 3458764513820700038, @t_now - INTERVAL 703 MINUTE),
(3458764513820700040, '范围查询后面的列用不上索引, 这条规则我给新人讲过很多遍', 3458764513820600016, 3458764513820541048, 0, @t_now - INTERVAL 1395 MINUTE),
(3458764513820700041, '等值写在最前面, 范围写在最后, 顺序其实很重要', 3458764513820600016, 3458764513820541049, 0, @t_now - INTERVAL 1224 MINUTE),
(3458764513820700042, '还有一点是索引列上不要做运算, 否则也会失效', 3458764513820600016, 3458764513820541030, 0, @t_now - INTERVAL 1001 MINUTE),
(3458764513820700043, '按需引入之后记得检查 CSS 有没有跟着按需, 我们漏过一次样式丢了一半', 3458764513820600017, 3458764513820541051, 0, @t_now - INTERVAL 1582 MINUTE),
(3458764513820700044, '体积优化最好配一个分析工具, 不然容易做无用功', 3458764513820600017, 3458764513820541053, 0, @t_now - INTERVAL 1398 MINUTE),
(3458764513820700045, '回复楼上: 我们用 rollup 的可视化插件, 一眼就能看出谁最大', 3458764513820600017, 3458764513820541017, 3458764513820700044, @t_now - INTERVAL 1147 MINUTE),
(3458764513820700046, '火焰图我一般是先看有没有特别宽的平顶, 那通常就是瓶颈', 3458764513820600018, 3458764513820541038, 0, @t_now - INTERVAL 1711 MINUTE),
(3458764513820700047, '加上 -http 起个页面比看文本直观多了', 3458764513820600018, 3458764513820541039, 0, @t_now - INTERVAL 1512 MINUTE),
(3458764513820700048, '采样频率别调太高, 生产环境还是要控制开销', 3458764513820600018, 3458764513820541029, 0, @t_now - INTERVAL 1246 MINUTE),
(3458764513820700049, '混用 new 和 free 这个坑我们是靠 ASAN 扫出来的, 平时根本看不出来', 3458764513820600019, 3458764513820541055, 0, @t_now - INTERVAL 1877 MINUTE),
(3458764513820700050, '带虚析构的类型更要注意, 基类析构不是虚的还会内存泄漏', 3458764513820600019, 3458764513820541057, 0, @t_now - INTERVAL 1655 MINUTE),
(3458764513820700051, '现在写新代码基本用智能指针, 裸 new 已经很难见到了', 3458764513820600019, 3458764513820541019, 0, @t_now - INTERVAL 1362 MINUTE),
(3458764513820700052, '误判率公式建议直接记结论, 现场推导很容易出错', 3458764513820600020, 3458764513820541008, 0, @t_now - INTERVAL 2006 MINUTE),
(3458764513820700053, 'Redis 的布隆过滤器模块可以直接用, 不用自己实现', 3458764513820600020, 3458764513820541046, 0, @t_now - INTERVAL 1765 MINUTE),
(3458764513820700054, '注意它不支持删除元素, 需要删除的场景要换计数型布隆', 3458764513820600020, 3458764513820541023, 0, @t_now - INTERVAL 1450 MINUTE),
(3458764513820700055, 'override 文件多了之后容易互相覆盖, 建议用 helmfile 管理', 3458764513820600021, 3458764513820541021, 0, @t_now - INTERVAL 2224 MINUTE),
(3458764513820700056, '小团队其实一份完整的 values 更不容易搞错', 3458764513820600021, 3458764513820541057, 0, @t_now - INTERVAL 1956 MINUTE),
(3458764513820700057, '回复楼上: 对, 关键是要有 lint 兜底', 3458764513820600021, 3458764513820541010, 3458764513820700056, @t_now - INTERVAL 1608 MINUTE),
(3458764513820700058, '链子太长也不好读, 我一般控制在三步以内', 3458764513820600022, 3458764513820541054, 0, @t_now - INTERVAL 2392 MINUTE),
(3458764513820700059, 'ok_or_else 比 unwrap 好太多, 起码不会直接崩', 3458764513820600022, 3458764513820541018, 0, @t_now - INTERVAL 2103 MINUTE),
(3458764513820700060, '这个特性配合 ? 运算符写起来非常舒服', 3458764513820600022, 3458764513820541022, 0, @t_now - INTERVAL 1729 MINUTE),
(3458764513820700061, '同类内部调用失效这个坑, 我们是把方法抽到另一个 bean 才解决的', 3458764513820600023, 3458764513820541023, 0, @t_now - INTERVAL 2603 MINUTE),
(3458764513820700062, '还要注意事务方法里不要做远程调用, 事务会拉得很长', 3458764513820600023, 3458764513820541043, 0, @t_now - INTERVAL 2287 MINUTE),
(3458764513820700063, '异常被 catch 住不抛出是最隐蔽的一种, 日志里看不出来', 3458764513820600023, 3458764513820541013, 0, @t_now - INTERVAL 1880 MINUTE),
(3458764513820700064, '跑量按每周一成递增这条铁律, 我是疼过之后才记住的', 3458764513820600024, 3458764513820541058, 0, @t_now - INTERVAL 2796 MINUTE),
(3458764513820700065, '跑前热身真的很关键, 我原来直接开跑, 膝盖一直不舒服', 3458764513820600024, 3458764513820541059, 0, @t_now - INTERVAL 2458 MINUTE),
(3458764513820700066, '回复楼上: 加五分钟动态热身之后基本没再疼过', 3458764513820600024, 3458764513820541024, 3458764513820700065, @t_now - INTERVAL 2020 MINUTE),
(3458764513820700067, 'errgroup 最舒服的是配合 signal.NotifyContext, 优雅退出一套就齐了', 3458764513820600025, 3458764513820541025, 0, @t_now - INTERVAL 2997 MINUTE),
(3458764513820700068, '注意默认只返回第一个错误, 需要收集全部错误要额外处理', 3458764513820600025, 3458764513820541016, 0, @t_now - INTERVAL 2634 MINUTE),
(3458764513820700069, 'SetLimit 也能用来控制并发度, 不用自己写信号量', 3458764513820600025, 3458764513820541029, 0, @t_now - INTERVAL 2164 MINUTE),
(3458764513820700070, '并发刷新只刷一次这个点太重要了, 我们之前就是并发刷新导致 token 互相覆盖', 3458764513820600026, 3458764513820541026, 0, @t_now - INTERVAL 3209 MINUTE),
(3458764513820700071, '记得给刷新请求本身也加上防重入标记, 否则会死循环', 3458764513820600026, 3458764513820541051, 0, @t_now - INTERVAL 2823 MINUTE),
(3458764513820700072, '回复楼上: 我们是用一个 retried 字段标记重放请求的', 3458764513820600026, 3458764513820541002, 3458764513820700071, @t_now - INTERVAL 2319 MINUTE),
(3458764513820700073, '延迟关联这招对大表特别有效, 回表次数少了一个数量级', 3458764513820600027, 3458764513820541027, 0, @t_now - INTERVAL 3431 MINUTE),
(3458764513820700074, '业务上限制最大页数是最省事的, 反正也没人翻到一百页', 3458764513820600027, 3458764513820541049, 0, @t_now - INTERVAL 3015 MINUTE),
(3458764513820700075, '游标分页要注意排序字段不能有重复值, 最好加上主键做二级排序', 3458764513820600027, 3458764513820541020, 0, @t_now - INTERVAL 2478 MINUTE),
(3458764513820700076, '类型存根缺失的时候也可以自己写一个 .pyi, 比 ignore 干净', 3458764513820600028, 3458764513820541028, 0, @t_now - INTERVAL 3652 MINUTE),
(3458764513820700077, '建议 CI 里跑 mypy 但不要阻塞合并, 先当成提示', 3458764513820600028, 3458764513820541016, 0, @t_now - INTERVAL 3210 MINUTE),
(3458764513820700078, 'pydantic 的类型推导其实也依赖这些注解, 收益是双份的', 3458764513820600028, 3458764513820541052, 0, @t_now - INTERVAL 2637 MINUTE),
(3458764513820700079, '引用这条规则我们写进规范了, review 时确实少了很多讨论', 3458764513820600029, 3458764513820541029, 0, @t_now - INTERVAL 3885 MINUTE),
(3458764513820700080, '补充一点: 返回内部数据的接口千万别返回引用, 生命周期会被调用方搞错', 3458764513820600029, 3458764513820541056, 0, @t_now - INTERVAL 3413 MINUTE),
(3458764513820700081, '回复楼上: 对, 这种还是老老实实返回值或者 const 引用加文档说明', 3458764513820600029, 3458764513820541055, 3458764513820700080, @t_now - INTERVAL 2804 MINUTE),
(3458764513820700082, '生命周期标注这块我建议直接看编译器提示, 它会告诉你在哪加', 3458764513820600030, 3458764513820541030, 0, @t_now - INTERVAL 4104 MINUTE),
(3458764513820700083, '结构体里存引用的时候基本都要标, 存所有权就不用', 3458764513820600030, 3458764513820541054, 0, @t_now - INTERVAL 3604 MINUTE),
(3458764513820700084, '回复楼上: 对, 能存所有权就尽量存, 简单很多', 3458764513820600030, 3458764513820541006, 3458764513820700083, @t_now - INTERVAL 2961 MINUTE),
(3458764513820700085, '堆大小我一般先设成容器内存上限的 70%, 再根据 GC 日志微调', 3458764513820600031, 3458764513820541031, 0, @t_now - INTERVAL 4351 MINUTE),
(3458764513820700086, 'ZGC 的停顿确实小, 但吞吐和内存占用要权衡', 3458764513820600031, 3458764513820541043, 0, @t_now - INTERVAL 3821 MINUTE),
(3458764513820700087, '不要忘了给容器也配上 -XX:MaxRAMPercentage', 3458764513820600031, 3458764513820541041, 0, @t_now - INTERVAL 3140 MINUTE),
(3458764513820700088, 'describe 的 Events 一定要看, 我见过因为镜像拉取失败 Pending 两小时的', 3458764513820600032, 3458764513820541032, 0, @t_now - INTERVAL 4579 MINUTE),
(3458764513820700089, '还有一种是节点打满了 taint, 调度器直接不选它', 3458764513820600032, 3458764513820541044, 0, @t_now - INTERVAL 4020 MINUTE),
(3458764513820700090, 'PVC 没绑定的时候也会一直 Pending, 但状态上不太明显', 3458764513820600032, 3458764513820541016, 0, @t_now - INTERVAL 3303 MINUTE),
(3458764513820700091, '红黑树在标准库里的实现细节也值得看一下, 旋转那套很巧妙', 3458764513820600033, 3458764513820541033, 0, @t_now - INTERVAL 4793 MINUTE),
(3458764513820700092, '面试里经常问为什么不用 AVL, 答案就是写多读少的场景更多', 3458764513820600033, 3458764513820541057, 0, @t_now - INTERVAL 4208 MINUTE),
(3458764513820700093, '跳表其实也是这个思路的另一个方向, 实现更简单', 3458764513820600033, 3458764513820541008, 0, @t_now - INTERVAL 3458 MINUTE),
(3458764513820700094, '同一肌群隔天练这个安排, 我之前天天练反而一直不长', 3458764513820600034, 3458764513820541034, 0, @t_now - INTERVAL 5022 MINUTE),
(3458764513820700095, '睡眠和吃不够的话, 练得再勤也没效果', 3458764513820600034, 3458764513820541024, 0, @t_now - INTERVAL 4407 MINUTE),
(3458764513820700096, '左手也要练, 只会右手变向的话防守人太好猜了', 3458764513820600035, 3458764513820541035, 0, @t_now - INTERVAL 5264 MINUTE),
(3458764513820700097, '先练低重心再练花活, 重心高的话什么动作都过不了人', 3458764513820600035, 3458764513820541042, 0, @t_now - INTERVAL 4620 MINUTE),
(3458764513820700098, '每天原地运球十分钟, 两个星期之后手感明显不一样', 3458764513820600035, 3458764513820541029, 0, @t_now - INTERVAL 3798 MINUTE),
(3458764513820700099, 'pinia 的 setup 写法配合 TypeScript 体验确实好', 3458764513820600036, 3458764513820541036, 0, @t_now - INTERVAL 5501 MINUTE),
(3458764513820700100, '老项目迁移最麻烦的是 modules 的映射, 建议一个模块一个模块来', 3458764513820600036, 3458764513820541051, 0, @t_now - INTERVAL 4828 MINUTE),
(3458764513820700101, '挡拆之后如果防守人从下面绕, 也可以直接顺下打篮下', 3458764513820600037, 3458764513820541037, 0, @t_now - INTERVAL 5753 MINUTE),
(3458764513820700102, '掩护质量比时机更重要, 站不住的掩护基本都是进攻犯规', 3458764513820600037, 3458764513820541050, 0, @t_now - INTERVAL 5048 MINUTE),
(3458764513820700103, '回复楼上: 对, 我们野球场十次挡拆有三次是因为掩护没站住', 3458764513820600037, 3458764513820541003, 3458764513820700102, @t_now - INTERVAL 4151 MINUTE),
(3458764513820700104, '业余比赛阵型能保持住就已经赢一半了', 3458764513820600038, 3458764513820541038, 0, @t_now - INTERVAL 5994 MINUTE),
(3458764513820700105, '我们试过 433 结果两个边路来回跑, 半场就跑不动了', 3458764513820600038, 3458764513820541026, 0, @t_now - INTERVAL 5262 MINUTE),
(3458764513820700106, 'venv 最大的好处是简单, 出问题好排查', 3458764513820600039, 3458764513820541039, 0, @t_now - INTERVAL 6246 MINUTE),
(3458764513820700107, 'conda 的环境隔离更彻底, 但依赖解析有时候很慢', 3458764513820600039, 3458764513820541015, 0, @t_now - INTERVAL 5481 MINUTE),
(3458764513820700108, '我们现在统一用 uv 了, 装包速度快一个数量级', 3458764513820600039, 3458764513820541052, 0, @t_now - INTERVAL 4508 MINUTE),
(3458764513820700109, 'RR 下的快照读和当前读确实容易混, 建议画个时间线理解', 3458764513820600040, 3458764513820541040, 0, @t_now - INTERVAL 6503 MINUTE),
(3458764513820700110, '间隙锁的范围和索引类型有关, 没有索引的时候会锁很多行', 3458764513820600040, 3458764513820541048, 0, @t_now - INTERVAL 5704 MINUTE),
(3458764513820700111, '面试被问过这个, 当时只答了隔离级别没提间隙锁', 3458764513820600040, 3458764513820541050, 0, @t_now - INTERVAL 4692 MINUTE),
(3458764513820700112, '不定行高维护前缀高度数组那个方案, 用二分定位起点, 性能可以接受', 3458764513820600041, 3458764513820541041, 0, @t_now - INTERVAL 6766 MINUTE),
(3458764513820700113, '白屏那一帧建议用 overscan 多渲染几行, 体验会好很多', 3458764513820600041, 3458764513820541024, 0, @t_now - INTERVAL 6582 MINUTE),
(3458764513820700114, '回复楼上: 对, 我们上下各多渲染 5 行就基本看不出来了', 3458764513820600041, 3458764513820541021, 3458764513820700113, @t_now - INTERVAL 6238 MINUTE),
(3458764513820700115, 'context 一路传下去这句话说起来简单, 落地要靠 code review 坚持', 3458764513820600042, 3458764513820541042, 0, @t_now - INTERVAL 7028 MINUTE),
(3458764513820700116, '我们后来加了 goroutine 数量的监控, 超过阈值就告警', 3458764513820600042, 3458764513820541029, 0, @t_now - INTERVAL 6831 MINUTE),
(3458764513820700117, 'http.Client 也要设超时, 不然默认是没有超时的', 3458764513820600042, 3458764513820541038, 0, @t_now - INTERVAL 6472 MINUTE),
(3458764513820700118, 'shared_ptr 的引用计数是原子的, 我们有一次在热点循环里拷贝它, 直接成了瓶颈', 3458764513820600043, 3458764513820541043, 0, @t_now - INTERVAL 7299 MINUTE),
(3458764513820700119, '所以传参用 const 引用, 只有确实要共享的时候才拷贝所有权', 3458764513820600043, 3458764513820541055, 0, @t_now - INTERVAL 7098 MINUTE),
(3458764513820700120, 'weak_ptr 也能顺手提一下, 解决循环引用就靠它', 3458764513820600043, 3458764513820541019, 0, @t_now - INTERVAL 6724 MINUTE),
(3458764513820700121, '小顶堆求最大的 k 个, 这个方向别搞反', 3458764513820600044, 3458764513820541044, 0, @t_now - INTERVAL 7576 MINUTE),
(3458764513820700122, '数据量不大的时候其实 quickselect 更快', 3458764513820600044, 3458764513820541056, 0, @t_now - INTERVAL 7372 MINUTE),
(3458764513820700123, '海量数据还可以先分片再各自取 TopK 归并', 3458764513820600044, 3458764513820541045, 0, @t_now - INTERVAL 6988 MINUTE),
(3458764513820700124, '启动探针确实要单独配, 我们加了 startupProbe 之后重启次数降了两个数量级', 3458764513820600045, 3458764513820541045, 0, @t_now - INTERVAL 7857 MINUTE),
(3458764513820700125, 'failureThreshold 至少给 3, 给 1 就是在赌运气', 3458764513820600045, 3458764513820541044, 0, @t_now - INTERVAL 7649 MINUTE),
(3458764513820700126, 'readiness 和 liveness 用同一个接口也很容易出问题', 3458764513820600045, 3458764513820541032, 0, @t_now - INTERVAL 7256 MINUTE),
(3458764513820700127, '所以 Rust 里能靠类型解决的问题就不要留到运行时', 3458764513820600046, 3458764513820541046, 0, @t_now - INTERVAL 8140 MINUTE),
(3458764513820700128, '刚上手的时候天天和借用检查器打架, 一个月之后就习惯了', 3458764513820600046, 3458764513820541054, 0, @t_now - INTERVAL 7928 MINUTE),
(3458764513820700129, 'Cell 和 RefCell 的区别也可以顺便讲一下, 面试常问', 3458764513820600046, 3458764513820541022, 0, @t_now - INTERVAL 7531 MINUTE),
(3458764513820700130, '初始容量按 预估元素数 除以 0.75 再向上取整就行', 3458764513820600047, 3458764513820541047, 0, @t_now - INTERVAL 8429 MINUTE),
(3458764513820700131, 'JDK8 之后链表转红黑树的阈值是 8, 这个也可以提一下', 3458764513820600047, 3458764513820541023, 0, @t_now - INTERVAL 8213 MINUTE),
(3458764513820700132, '回复楼上: 对, 而且树化之后扩容又会退化回链表', 3458764513820600047, 3458764513820541013, 3458764513820700131, @t_now - INTERVAL 7811 MINUTE),
(3458764513820700133, '晨跑最大的好处是没人打扰, 一天里最安静的一小时', 3458764513820600048, 3458764513820541048, 0, @t_now - INTERVAL 8721 MINUTE),
(3458764513820700134, '夜跑的话建议选有路灯的路线, 安全第一', 3458764513820600048, 3458764513820541059, 0, @t_now - INTERVAL 8501 MINUTE),
(3458764513820700135, '固定时间这一点最重要, 身体会形成习惯', 3458764513820600048, 3458764513820541058, 0, @t_now - INTERVAL 8094 MINUTE),
(3458764513820700136, '我们也是统一返回空切片, 前端省了很多判空', 3458764513820600049, 3458764513820541049, 0, @t_now - INTERVAL 9016 MINUTE),
(3458764513820700137, 'slice 的 json tag 上加 omitempty 反而会让空数组消失, 要注意', 3458764513820600049, 3458764513820541016, 0, @t_now - INTERVAL 8792 MINUTE),
(3458764513820700138, '这个坑在对接第三方的时候遇到过, 对方直接把 null 当异常处理了', 3458764513820600049, 3458764513820541041, 0, @t_now - INTERVAL 8380 MINUTE),
(3458764513820700139, 'unobserve 这一步确实容易忘, 特别是列表会被销毁重建的场景', 3458764513820600050, 3458764513820541050, 0, @t_now - INTERVAL 9313 MINUTE),
(3458764513820700140, '也可以用 rootMargin 提前一点加载, 体验更顺', 3458764513820600050, 3458764513820541051, 0, @t_now - INTERVAL 9085 MINUTE),
(3458764513820700141, '首屏的图片建议直接设成 eager, 别都交给懒加载', 3458764513820600050, 3458764513820541054, 0, @t_now - INTERVAL 8668 MINUTE),
(3458764513820700142, '这个我专门测过, 执行计划完全一样, 确实不用纠结', 3458764513820600051, 3458764513820541051, 0, @t_now - INTERVAL 9700 MINUTE),
(3458764513820700143, 'count 加 where 条件的时候索引选择才是关键', 3458764513820600051, 3458764513820541048, 0, @t_now - INTERVAL 9468 MINUTE),
(3458764513820700144, 'MyISAM 时代两者确实有区别, 现在可以忘了', 3458764513820600051, 3458764513820541040, 0, @t_now - INTERVAL 9046 MINUTE),
(3458764513820700145, 'pip-compile 生成的锁文件记得一起提交, 不然 CI 和本地会不一致', 3458764513820600052, 3458764513820541052, 0, @t_now - INTERVAL 10438 MINUTE),
(3458764513820700146, '加上 --generate-hashes 还能做完整性校验', 3458764513820600052, 3458764513820541016, 0, @t_now - INTERVAL 10201 MINUTE),
(3458764513820700147, '间接依赖冲突的时候 pip-compile 的报错信息比 pip 清楚得多', 3458764513820600052, 3458764513820541015, 0, @t_now - INTERVAL 9771 MINUTE),
(3458764513820700148, 'memmove 稍微慢一点, 但重叠场景下用 memcpy 真的是随机行为', 3458764513820600053, 3458764513820541053, 0, @t_now - INTERVAL 11191 MINUTE),
(3458764513820700149, '编译器有时会把 memcpy 优化成更宽的指令, 光看源码看不出差别', 3458764513820600053, 3458764513820541029, 0, @t_now - INTERVAL 10949 MINUTE),
(3458764513820700150, '回复楼上: 对, 所以压测时最好看反汇编确认一下', 3458764513820600053, 3458764513820541005, 3458764513820700149, @t_now - INTERVAL 10514 MINUTE),
(3458764513820700151, '这个速查表很实用, 我一般再加上一个 Cow 的场景', 3458764513820600054, 3458764513820541054, 0, @t_now - INTERVAL 12049 MINUTE),
(3458764513820700152, 'Arc 套 Mutex 是最常见的组合, 但锁粒度要控制好', 3458764513820600054, 3458764513820541046, 0, @t_now - INTERVAL 11803 MINUTE),
(3458764513820700153, '单线程里能用 Rc 就别用 Arc, 省下的原子操作很可观', 3458764513820600054, 3458764513820541042, 0, @t_now - INTERVAL 11363 MINUTE),
(3458764513820700154, 'volatile 不保证原子性这点面试几乎必问', 3458764513820600055, 3458764513820541055, 0, @t_now - INTERVAL 13000 MINUTE),
(3458764513820700155, 'long 的写入在 32 位机器上还可能撕裂, 更要注意', 3458764513820600055, 3458764513820541023, 0, @t_now - INTERVAL 12749 MINUTE),
(3458764513820700156, '建议顺便看一下内存屏障, 理解会更透彻', 3458764513820600055, 3458764513820541013, 0, @t_now - INTERVAL 12304 MINUTE),
(3458764513820700157, '自定义指标要接适配器, 链路确实长, 但值得做', 3458764513820600056, 3458764513820541056, 0, @t_now - INTERVAL 14195 MINUTE),
(3458764513820700158, '扩容的冷却时间也别设太短, 不然会反复伸缩', 3458764513820600056, 3458764513820541045, 0, @t_now - INTERVAL 13939 MINUTE),
(3458764513820700159, '我们最后是提前扩容加定时伸缩, 比 HPA 稳', 3458764513820600056, 3458764513820541043, 0, @t_now - INTERVAL 13489 MINUTE),
(3458764513820700160, '前缀函数的定义一定要先统一, 不同教材写法不一样', 3458764513820600057, 3458764513820541057, 0, @t_now - INTERVAL 15518 MINUTE),
(3458764513820700161, '自己匹配自己那一步想通了就很简单了', 3458764513820600057, 3458764513820541055, 0, @t_now - INTERVAL 15257 MINUTE),
(3458764513820700162, '实际工程里字符串匹配基本都用现成库, 但原理还是要懂', 3458764513820600057, 3458764513820541011, 0, @t_now - INTERVAL 14800 MINUTE),
(3458764513820700163, '有氧安排在力量训练之后或者隔天, 影响会小很多', 3458764513820600058, 3458764513820541058, 0, @t_now - INTERVAL 16890 MINUTE),
(3458764513820700164, '我一般用走路代替, 半小时快走也能到有氧区间', 3458764513820600058, 3458764513820541059, 0, @t_now - INTERVAL 16624 MINUTE),
(3458764513820700165, '增肌期最怕的是有氧做太多, 热量缺口拉着就练不动了', 3458764513820600058, 3458764513820541060, 0, @t_now - INTERVAL 16161 MINUTE),
(3458764513820700166, '命名返回值加 defer 这个组合我第一次看到的时候也觉得很神奇', 3458764513820600059, 3458764513820541059, 0, @t_now - INTERVAL 18443 MINUTE),
(3458764513820700167, '不过团队里最好统一风格, 不然读代码的人会绕晕', 3458764513820600059, 3458764513820541030, 0, @t_now - INTERVAL 18172 MINUTE),
(3458764513820700168, 'recover 写在 defer 里也是靠这个机制生效的', 3458764513820600059, 3458764513820541025, 0, @t_now - INTERVAL 17703 MINUTE),
(3458764513820700169, 'provide/inject 我一般只在插件和表单这类深层组件里用', 3458764513820600060, 3458764513820541060, 0, @t_now - INTERVAL 20118 MINUTE),
(3458764513820700170, '用 toRefs 包一层这个细节太关键了, 我们踩过一次', 3458764513820600060, 3458764513820541051, 0, @t_now - INTERVAL 19842 MINUTE),
(3458764513820700171, '回复楼上: 后来我们干脆改成注入一个 reactive 对象, 也行', 3458764513820600060, 3458764513820541053, 3458764513820700170, @t_now - INTERVAL 19368 MINUTE),
(3458764513820700172, '重心低这一点最难坚持, 累了之后第一个变形的就是防守姿势', 3458764513820600061, 3458764513820541058, 0, @t_now - INTERVAL 21892 MINUTE),
(3458764513820700173, '放投不放突在野球场特别有效, 大部分人三分其实不稳', 3458764513820600061, 3458764513820541050, 0, @t_now - INTERVAL 21611 MINUTE),
(3458764513820700174, '滑步练起来很枯燥, 但是真的能救命', 3458764513820600061, 3458764513820541019, 0, @t_now - INTERVAL 21131 MINUTE),
(3458764513820700175, '下底传中的落点比力度重要, 传到点球点附近最有威胁', 3458764513820600062, 3458764513820541059, 0, @t_now - INTERVAL 23864 MINUTE),
(3458764513820700176, '内切之前先做一次假动作, 不然很容易被边后卫卡住', 3458764513820600062, 3458764513820541014, 0, @t_now - INTERVAL 23578 MINUTE),
(3458764513820700177, '回复楼上: 对, 而且内切之后回传也是很好的选择', 3458764513820600062, 3458764513820541026, 3458764513820700176, @t_now - INTERVAL 23093 MINUTE),
(3458764513820700178, 'traceID 最好在网关生成并透传, 这样全链路都能串起来', 3458764513820600063, 3458764513820541003, 0, @t_now - INTERVAL 25978 MINUTE),
(3458764513820700179, '我们还在日志里带上了用户 ID, 排查具体用户问题很方便', 3458764513820600063, 3458764513820541042, 0, @t_now - INTERVAL 25712 MINUTE),
(3458764513820700180, '大事务拆小是第一步, 第二步是把 binlog 格式换成 row', 3458764513820600064, 3458764513820541004, 0, @t_now - INTERVAL 28202 MINUTE),
(3458764513820700181, '并行复制要开 slave_parallel_workers, 不然还是单线程回放', 3458764513820600064, 3458764513820541048, 0, @t_now - INTERVAL 27931 MINUTE),
(3458764513820700182, '读从库的场景最好做业务上的兜底, 比如写后读走主库', 3458764513820600064, 3458764513820541050, 0, @t_now - INTERVAL 27640 MINUTE),
(3458764513820700183, 'markRaw 这个 API 之前一直没用过, 学到了', 3458764513820600065, 3458764513820541005, 0, @t_now - INTERVAL 30649 MINUTE),
(3458764513820700184, '所以 store 里存实例是很容易出问题的, 建议只存普通数据', 3458764513820600065, 3458764513820541036, 0, @t_now - INTERVAL 30372 MINUTE),
(3458764513820700185, '参数用结构体还有个好处是以后加字段不用改所有调用点', 3458764513820600066, 3458764513820541006, 0, @t_now - INTERVAL 33233 MINUTE),
(3458764513820700186, '我们团队的做法是三行以上的参数一律用结构体', 3458764513820600066, 3458764513820541011, 0, @t_now - INTERVAL 32951 MINUTE),
(3458764513820700187, 'const 在星号右边修饰指针这点我记了很多次才记牢', 3458764513820600067, 3458764513820541007, 0, @t_now - INTERVAL 36000 MINUTE),
(3458764513820700188, '团队里建议统一用 east const, 至少读起来顺序一致', 3458764513820600067, 3458764513820541057, 0, @t_now - INTERVAL 35712 MINUTE),
(3458764513820700189, 'Redis 的 zset 就是这个结构, 面试聊到跳表可以顺便提一下', 3458764513820600068, 3458764513820541008, 0, @t_now - INTERVAL 38978 MINUTE),
(3458764513820700190, '跳表的层数是随机的, 所以时间复杂度是期望值', 3458764513820600068, 3458764513820541046, 0, @t_now - INTERVAL 38685 MINUTE),
(3458764513820700191, '环境变量不会自动更新这个坑我们踩过, 排查了很久', 3458764513820600069, 3458764513820541009, 0, @t_now - INTERVAL 42123 MINUTE),
(3458764513820700192, '也可以挂载之后在容器里加一个 reloader 侧车', 3458764513820600069, 3458764513820541044, 0, @t_now - INTERVAL 41825 MINUTE),
(3458764513820700193, '按可复用来拆这点很重要, 单纯按目录拆最后会互相依赖', 3458764513820600070, 3458764513820541010, 0, @t_now - INTERVAL 45491 MINUTE),
(3458764513820700194, 'workspace 里共享依赖版本也是它的一个好处', 3458764513820600070, 3458764513820541054, 0, @t_now - INTERVAL 45188 MINUTE),
(3458764513820700195, 'handle 里既能拿结果也能拿异常, 比 exceptionally 灵活', 3458764513820600071, 3458764513820541011, 0, @t_now - INTERVAL 49000 MINUTE),
(3458764513820700196, '记得最后统一 join, 否则主线程可能提前结束', 3458764513820600071, 3458764513820541023, 0, @t_now - INTERVAL 48692 MINUTE),
(3458764513820700197, '跑量不大的话确实不用迷信顶级缓震, 合脚最重要', 3458764513820600072, 3458764513820541012, 0, @t_now - INTERVAL 52758 MINUTE),
(3458764513820700198, '扁平足建议先去店里做个足型测试, 比看评测靠谱', 3458764513820600072, 3458764513820541024, 0, @t_now - INTERVAL 52445 MINUTE),
(3458764513820700199, '跟腱的问题我拖了半年, 后来老老实实停了两个月才彻底好', 3458764513820600073, 3458764513820541013, 0, @t_now - INTERVAL 56703 MINUTE),
(3458764513820700200, '换成游泳和骑车确实能维持体能, 复出之后不至于完全没状态', 3458764513820600073, 3458764513820541042, 0, @t_now - INTERVAL 56385 MINUTE),
(3458764513820700201, 'useRequest 这类封装一定要支持手动取消, 不然切页面会报错', 3458764513820600074, 3458764513820541014, 0, @t_now - INTERVAL 60978 MINUTE),
(3458764513820700202, '我们是用 AbortController 统一处理的', 3458764513820600074, 3458764513820541026, 0, @t_now - INTERVAL 60655 MINUTE),
(3458764513820700203, '预发用同等数据量验证这点特别重要, 我们吃过一次亏', 3458764513820600075, 3458764513820541015, 0, @t_now - INTERVAL 65423 MINUTE),
(3458764513820700204, 'pt-online-schema-change 也可以了解一下, 大表改结构很稳', 3458764513820600075, 3458764513820541048, 0, @t_now - INTERVAL 65095 MINUTE),
(3458764513820700205, 'FastAPI 的自动文档对内部工具来说太省事了', 3458764513820600076, 3458764513820541016, 0, @t_now - INTERVAL 70100 MINUTE),
(3458764513820700206, '建议再加一层鉴权, 内部工具也不能裸奔', 3458764513820600076, 3458764513820541042, 0, @t_now - INTERVAL 69767 MINUTE),
(3458764513820700207, '按对齐要求排字段这个习惯很好, 我们一个结构体从 64 字节压到了 40', 3458764513820600077, 3458764513820541017, 0, @t_now - INTERVAL 75048 MINUTE),
(3458764513820700208, '跨平台通信时还要注意 pack, 不然两边解析出来的偏移不一样', 3458764513820600077, 3458764513820541055, 0, @t_now - INTERVAL 74710 MINUTE),
(3458764513820700209, 'anyhow 的 context 链在打日志的时候特别好用', 3458764513820600078, 3458764513820541018, 0, @t_now - INTERVAL 80223 MINUTE),
(3458764513820700210, '对外暴露的库用 thiserror 是为了让调用方能匹配具体错误', 3458764513820600078, 3458764513820541022, 0, @t_now - INTERVAL 79880 MINUTE),
(3458764513820700211, '二级缓存在多实例部署下确实很容易读到脏数据', 3458764513820600079, 3458764513820541019, 0, @t_now - INTERVAL 85623 MINUTE),
(3458764513820700212, '一级缓存的问题可以用 localCacheScope 关掉', 3458764513820600079, 3458764513820541023, 0, @t_now - INTERVAL 85275 MINUTE),
(3458764513820700213, 'Gateway API 的角色拆分对多团队协作确实更友好', 3458764513820600080, 3458764513820541020, 0, @t_now - INTERVAL 91298 MINUTE),
(3458764513820700214, '迁移之前先看自己的 Ingress 用了哪些注解, 有些能力还没对齐', 3458764513820600080, 3458764513820541045, 0, @t_now - INTERVAL 90945 MINUTE),
(3458764513820700215, '状态定义成"处理到第 i 个"真的能省很多思考时间', 3458764513820600081, 3458764513820541021, 0, @t_now - INTERVAL 97258 MINUTE),
(3458764513820700216, '也可以先从记忆化搜索写起, 再改成递推', 3458764513820600081, 3458764513820541056, 0, @t_now - INTERVAL 96900 MINUTE),
(3458764513820700217, '滚动数组优化空间的时候要注意依赖顺序', 3458764513820600081, 3458764513820541055, 0, @t_now - INTERVAL 96531 MINUTE),
(3458764513820700218, '动态热身和静态拉伸分开这一点很多人搞反了', 3458764513820600082, 3458764513820541022, 0, @t_now - INTERVAL 103500 MINUTE),
(3458764513820700219, '跑后拉伸我一般做五分钟小腿和大腿后侧, 第二天轻松很多', 3458764513820600082, 3458764513820541058, 0, @t_now - INTERVAL 103137 MINUTE),
(3458764513820700220, '预分配确实能省掉不少拷贝, 尤其是循环里 append 的场景', 3458764513820600083, 3458764513820541023, 0, @t_now - INTERVAL 110018 MINUTE),
(3458764513820700221, '切片共享底层数组这个特性也要注意, 容易造成内存不释放', 3458764513820600083, 3458764513820541049, 0, @t_now - INTERVAL 109650 MINUTE),
(3458764513820700222, '指令里移除节点比 v-if 清爽, 但要注意权限变化后要重新渲染', 3458764513820600084, 3458764513820541024, 0, @t_now - INTERVAL 116843 MINUTE),
(3458764513820700223, '我们是用 v-if 加权限 hook, 感觉更好调试', 3458764513820600084, 3458764513820541051, 0, @t_now - INTERVAL 116470 MINUTE),
(3458764513820700224, '三对三里会传球的人真的很难防, 尤其是手递手之后直接投', 3458764513820600085, 3458764513820541025, 0, @t_now - INTERVAL 123978 MINUTE),
(3458764513820700225, '五对五最怕的就是不落位, 三个人挤在一侧要球', 3458764513820600085, 3458764513820541050, 0, @t_now - INTERVAL 123600 MINUTE),
(3458764513820700226, '我们打半场的时候会先约好谁做轴, 节奏顺很多', 3458764513820600085, 3458764513820541058, 0, @t_now - INTERVAL 123219 MINUTE),
(3458764513820700227, '看无球跑动这句话太对了, 以前只会盯着球看', 3458764513820600086, 3458764513820541026, 0, @t_now - INTERVAL 131400 MINUTE),
(3458764513820700228, '建议看一次回放只看一个球员, 收获比看全场多', 3458764513820600086, 3458764513820541014, 0, @t_now - INTERVAL 131017 MINUTE),
(3458764513820700229, 'functools.wraps 不加的话 IDE 跳转也会失效', 3458764513820600087, 3458764513820541027, 0, @t_now - INTERVAL 139103 MINUTE),
(3458764513820700230, '类装饰器也要注意保留元信息, 同样的道理', 3458764513820600087, 3458764513820541052, 0, @t_now - INTERVAL 138715 MINUTE),
(3458764513820700231, '磁盘余量这一点很容易被忽略, 加索引会占大量空间', 3458764513820600088, 3458764513820541028, 0, @t_now - INTERVAL 147108 MINUTE),
(3458764513820700232, '分多次加索引的时候可以按主键区间来切, 方便控制影响', 3458764513820600088, 3458764513820541048, 0, @t_now - INTERVAL 146715 MINUTE),
(3458764513820700233, '从工具函数开始写测试确实最容易见效', 3458764513820600089, 3458764513820541029, 0, @t_now - INTERVAL 155428 MINUTE),
(3458764513820700234, '组件测试建议优先覆盖复杂交互, 而不是渲染细节', 3458764513820600089, 3458764513820541053, 0, @t_now - INTERVAL 155030 MINUTE),
(3458764513820700235, '先跑通第一个测试比覆盖率数字更重要', 3458764513820600089, 3458764513820541060, 0, @t_now - INTERVAL 154627 MINUTE),
(3458764513820700236, 'singleflight 配合本地缓存能挡住大部分缓存击穿', 3458764513820600090, 3458764513820541030, 0, @t_now - INTERVAL 164078 MINUTE),
(3458764513820700237, '注意它只对同一进程内的调用有效, 多实例还要靠分布式锁', 3458764513820600090, 3458764513820541025, 0, @t_now - INTERVAL 163675 MINUTE),
(3458764513820700238, '返回的 shared 标记可以判断结果是不是被共享的, 挺好用', 3458764513820600090, 3458764513820541042, 0, @t_now - INTERVAL 163268 MINUTE),
(3458764513820700239, 'constexpr 配上 static_assert 做编译期校验特别舒服', 3458764513820600091, 3458764513820541031, 0, @t_now - INTERVAL 172900 MINUTE),
(3458764513820700240, '注意有些标准库函数在 C++20 才变成 constexpr, 老编译器上会报错', 3458764513820600091, 3458764513820541056, 0, @t_now - INTERVAL 172520 MINUTE),
(3458764513820700241, '开放寻址删除要用墓碑标记, 这点很容易忽略', 3458764513820600092, 3458764513820541032, 0, @t_now - INTERVAL 182110 MINUTE),
(3458764513820700242, '链地址法的链表在冲突多的时候也会退化成线性查找', 3458764513820600092, 3458764513820541011, 0, @t_now - INTERVAL 181725 MINUTE),
(3458764513820700243, 'ResourceQuota 上线之前要先统计一下现有用量, 不然会直接把服务卡住', 3458764513820600093, 3458764513820541033, 0, @t_now - INTERVAL 191590 MINUTE),
(3458764513820700244, '配上 LimitRange 给默认值, 忘记写 requests 的也能跑起来', 3458764513820600093, 3458764513820541044, 0, @t_now - INTERVAL 191200 MINUTE),
(3458764513820700245, 'move 之后原变量就不能用了, 这点和 C++ 的 move 语义不同', 3458764513820600094, 3458764513820541034, 0, @t_now - INTERVAL 201350 MINUTE),
(3458764513820700246, '跨线程传的时候还会要求 Send 约束, 编译器提示很明确', 3458764513820600094, 3458764513820541054, 0, @t_now - INTERVAL 200955 MINUTE),
(3458764513820700247, '唯一索引做幂等是最简单可靠的, 但要注意冲突时的错误处理', 3458764513820600095, 3458764513820541035, 0, @t_now - INTERVAL 211400 MINUTE),
(3458764513820700248, '状态机方案更适合有多个中间状态的业务', 3458764513820600095, 3458764513820541023, 0, @t_now - INTERVAL 211000 MINUTE),
(3458764513820700249, '中午散步二十分钟这一条我已经坚持半年了, 下午确实不困', 3458764513820600096, 3458764513820541036, 0, @t_now - INTERVAL 221740 MINUTE),
(3458764513820700250, '久坐最大的问题是髋屈肌紧张, 建议每天做几分钟拉伸', 3458764513820600096, 3458764513820541059, 0, @t_now - INTERVAL 221335 MINUTE),
(3458764513820700251, 'Ticker 忘 Stop 这个问题在长跑服务里特别隐蔽', 3458764513820600097, 3458764513820541037, 0, @t_now - INTERVAL 232380 MINUTE),
(3458764513820700252, '建议用 NewTimer 的场景就别用 Ticker, 用完即走更清晰', 3458764513820600097, 3458764513820541029, 0, @t_now - INTERVAL 231970 MINUTE),
(3458764513820700253, 'meta 里声明权限这个做法很清晰, 比在组件里判断好维护', 3458764513820600098, 3458764513820541038, 0, @t_now - INTERVAL 243330 MINUTE),
(3458764513820700254, '记得把白名单路由单独列出来, 不然容易把自己锁在外面', 3458764513820600098, 3458764513820541051, 0, @t_now - INTERVAL 242915 MINUTE),
(3458764513820700255, 'type 出现 ALL 的时候基本就是要加索引了', 3458764513820600099, 3458764513820541039, 0, @t_now - INTERVAL 254580 MINUTE),
(3458764513820700256, 'rows 是估算值, 但和实际差距太大的时候说明统计信息过期了', 3458764513820600099, 3458764513820541049, 0, @t_now - INTERVAL 254160 MINUTE),
(3458764513820700257, '处理大文件时生成器几乎是必选项, 内存差别太大了', 3458764513820600100, 3458764513820541040, 0, @t_now - INTERVAL 266140 MINUTE),
(3458764513820700258, 'yield from 还能把生成器串起来, 写起来更顺', 3458764513820600100, 3458764513820541015, 0, @t_now - INTERVAL 265715 MINUTE),
(3458764513820700259, 'erase 之后继续用旧迭代器这个 bug 我写过不止一次', 3458764513820600101, 3458764513820541041, 0, @t_now - INTERVAL 278010 MINUTE),
(3458764513820700260, 'C++20 的 erase_if 挺好用, 顺手能把返回值接住', 3458764513820600101, 3458764513820541055, 0, @t_now - INTERVAL 277580 MINUTE),
(3458764513820700261, '泛型单态化带来的编译时间和体积问题在大项目里很明显', 3458764513820600102, 3458764513820541042, 0, @t_now - INTERVAL 290180 MINUTE),
(3458764513820700262, '需要插件化的时候 trait object 几乎是唯一选择', 3458764513820600102, 3458764513820541046, 0, @t_now - INTERVAL 289745 MINUTE),
(3458764513820700263, '异步 Appender 的队列也要设上限, 不然 OOM 的时候日志先挂', 3458764513820600103, 3458764513820541043, 0, @t_now - INTERVAL 302630 MINUTE),
(3458764513820700264, '丢弃策略建议至少保留错误级别的日志', 3458764513820600103, 3458764513820541023, 0, @t_now - INTERVAL 302190 MINUTE),
(3458764513820700265, 'kind 启动快, 适合天天重启的开发流程', 3458764513820600104, 3458764513820541044, 0, @t_now - INTERVAL 315360 MINUTE),
(3458764513820700266, '需要模拟真实网络延迟的话, 还是用 minikube 加 tc 更方便', 3458764513820600104, 3458764513820541045, 0, @t_now - INTERVAL 314915 MINUTE),
(3458764513820700267, '多次区间加最后一次查询这个口诀我记了很久', 3458764513820600105, 3458764513820541045, 0, @t_now - INTERVAL 328350 MINUTE),
(3458764513820700268, '二维差分也是同样的思路, 就是容斥要写对', 3458764513820600105, 3458764513820541057, 0, @t_now - INTERVAL 327900 MINUTE),
(3458764513820700269, '说话法比看心率表实用, 尤其是刚开始跑的人', 3458764513820600106, 3458764513820541046, 0, @t_now - INTERVAL 341580 MINUTE),
(3458764513820700270, '我用储备心率法算出来的区间和这个差不多, 心里有数就行', 3458764513820600106, 3458764513820541058, 0, @t_now - INTERVAL 341125 MINUTE),
(3458764513820700271, '学无球跑位收益最大, 会跑位的人在哪都受欢迎', 3458764513820600107, 3458764513820541047, 0, @t_now - INTERVAL 355040 MINUTE),
(3458764513820700272, '职业球员的脚步细节在水泥地上根本做不出来, 学思路就够了', 3458764513820600107, 3458764513820541025, 0, @t_now - INTERVAL 354580 MINUTE),
(3458764513820700273, 'schema 驱动的表单在字段多的时候优势太明显了', 3458764513820600108, 3458764513820541048, 0, @t_now - INTERVAL 368710 MINUTE),
(3458764513820700274, '校验规则也建议抽出来复用, 前后端可以用同一份定义', 3458764513820600108, 3458764513820541053, 0, @t_now - INTERVAL 368245 MINUTE),
(3458764513820700275, '防守和篮板是最容易获得上场时间的, 尤其是水平参差的时候', 3458764513820600109, 3458764513820541049, 0, @t_now - INTERVAL 382570 MINUTE),
(3458764513820700276, '接球就传再切进去这一招我用了很多年, 成功率确实高', 3458764513820600109, 3458764513820541050, 0, @t_now - INTERVAL 382100 MINUTE),
(3458764513820700277, '小场地的第一脚触球质量决定一切, 停不好就没时间做下一个动作', 3458764513820600110, 3458764513820541050, 0, @t_now - INTERVAL 396670 MINUTE),
(3458764513820700278, '撞墙配合(二过一)是最简单也最有效的配合', 3458764513820600110, 3458764513820541026, 0, @t_now - INTERVAL 396195 MINUTE),
(3458764513820700279, 'contextmanager 写计时器比自己记时间靠谱多了', 3458764513820600111, 3458764513820541051, 0, @t_now - INTERVAL 411010 MINUTE),
(3458764513820700280, '记得用 perf_counter, 别用 time.time', 3458764513820600111, 3458764513820541028, 0, @t_now - INTERVAL 410530 MINUTE),
(3458764513820700281, '软删标记加过期清理是最稳的一条路', 3458764513820600112, 3458764513820541052, 0, @t_now - INTERVAL 425590 MINUTE),
(3458764513820700282, '分批删除的时候记得加上 sleep, 不然主从延迟会很难看', 3458764513820600112, 3458764513820541048, 0, @t_now - INTERVAL 425105 MINUTE),
(3458764513820700283, '骨架屏尺寸对齐这一点太关键了, 不然真的很跳', 3458764513820600113, 3458764513820541053, 0, @t_now - INTERVAL 440390 MINUTE),
(3458764513820700284, '也可以用固定的最小高度过渡一下, 减少布局抖动', 3458764513820600113, 3458764513820541054, 0, @t_now - INTERVAL 439900 MINUTE),
(3458764513820700285, '含 map 的结构体不能比较这个坑我踩过, 编译期就报错还算友好', 3458764513820600114, 3458764513820541054, 0, @t_now - INTERVAL 455430 MINUTE),
(3458764513820700286, '要比较的话可以用 reflect.DeepEqual, 但性能差很多', 3458764513820600114, 3458764513820541016, 0, @t_now - INTERVAL 454935 MINUTE),
(3458764513820700287, '工具函数写成一行反而更容易被内联, 长函数编译器基本会放弃', 3458764513820600115, 3458764513820541055, 0, @t_now - INTERVAL 470710 MINUTE),
(3458764513820700288, '现在一般交给编译器决定, 手写 inline 更多是表达意图', 3458764513820600115, 3458764513820541057, 0, @t_now - INTERVAL 470210 MINUTE),
(3458764513820700289, '递归空间这一点我经常忘, 谢谢提醒', 3458764513820600116, 3458764513820541056, 0, @t_now - INTERVAL 486240 MINUTE),
(3458764513820700290, '哈希的均摊分析建议结合扩容一起讲, 面试官一般会追问', 3458764513820600116, 3458764513820541057, 0, @t_now - INTERVAL 485735 MINUTE),
(3458764513820700291, '有状态服务用蓝绿确实痛苦, 数据迁移是个大工程', 3458764513820600117, 3458764513820541057, 0, @t_now - INTERVAL 502010 MINUTE),
(3458764513820700292, '金丝雀建议配合自动化回滚, 否则发现问题时已经全量了', 3458764513820600117, 3458764513820541044, 0, @t_now - INTERVAL 501500 MINUTE),
(3458764513820700293, '惰性求值这一点搞不清楚就会做出反向优化', 3458764513820600118, 3458764513820541058, 0, @t_now - INTERVAL 518030 MINUTE),
(3458764513820700294, 'collect 的时候预分配容量还能再快一点', 3458764513820600118, 3458764513820541054, 0, @t_now - INTERVAL 517515 MINUTE),
(3458764513820700295, '静态块里读配置是最常见的坏味道, 启动慢还不好排查', 3458764513820600119, 3458764513820541059, 0, @t_now - INTERVAL 534290 MINUTE),
(3458764513820700296, '初始化顺序问题可以用静态内部类延迟加载绕开', 3458764513820600119, 3458764513820541023, 0, @t_now - INTERVAL 533770 MINUTE),
(3458764513820700297, '静息心率下降这个指标最直观, 我也是半年之后就看到了', 3458764513820600120, 3458764513820541060, 0, @t_now - INTERVAL 544890 MINUTE),
(3458764513820700298, '体重没掉多少但围度变了, 这个体验很真实', 3458764513820600120, 3458764513820541029, 0, @t_now - INTERVAL 544365 MINUTE),
(3458764513820700299, '情绪变好这一点我最有共鸣, 跑完那一小时最放松', 3458764513820600120, 3458764513820541059, 0, @t_now - INTERVAL 543835 MINUTE),
(3458764513820700300, '回复楼上: 对, 有时候就是靠跑那半小时把一天的情绪消化掉', 3458764513820600120, 3458764513820541058, 3458764513820700298, @t_now - INTERVAL 543300 MINUTE);

-- 评论的 update_time 和 create_time 保持一致
UPDATE `comment` SET `update_time` = `create_time`;

-- ----------------------------------------------------------------------------
-- 7. 自检(只读查询, 不改数据)
--    "实际数量"要和"期望值"一致; 下面几条引用完整性检查期望值都是 0。
-- ----------------------------------------------------------------------------
SELECT '用户' AS `检查项`, COUNT(*) AS `实际数量`, 60 AS `期望值` FROM `user`;
SELECT '版块' AS `检查项`, COUNT(*) AS `实际数量`, 12 AS `期望值` FROM `community`;
SELECT '帖子' AS `检查项`, COUNT(*) AS `实际数量`, 120 AS `期望值` FROM `post`;
SELECT '评论' AS `检查项`, COUNT(*) AS `实际数量`, 300 AS `期望值` FROM `comment`;
SELECT '楼中楼评论' AS `检查项`, COUNT(*) AS `实际数量`, 20 AS `期望值` FROM `comment` WHERE `parent_id` <> 0;

SELECT '帖子引用了不存在的用户' AS `检查项`, COUNT(*) AS `实际数量`, 0 AS `期望值`
FROM `post` p LEFT JOIN `user` u ON u.user_id = p.author_id WHERE u.user_id IS NULL;
SELECT '帖子引用了不存在的版块' AS `检查项`, COUNT(*) AS `实际数量`, 0 AS `期望值`
FROM `post` p LEFT JOIN `community` c ON c.community_id = p.community_id WHERE c.community_id IS NULL;
SELECT '评论引用了不存在的帖子' AS `检查项`, COUNT(*) AS `实际数量`, 0 AS `期望值`
FROM `comment` c LEFT JOIN `post` p ON p.post_id = c.post_id WHERE p.post_id IS NULL;
SELECT '评论引用了不存在的用户' AS `检查项`, COUNT(*) AS `实际数量`, 0 AS `期望值`
FROM `comment` c LEFT JOIN `user` u ON u.user_id = c.author_id WHERE u.user_id IS NULL;
SELECT '回复了不存在的父评论或跨帖回复' AS `检查项`, COUNT(*) AS `实际数量`, 0 AS `期望值`
FROM `comment` c LEFT JOIN `comment` pc ON pc.comment_id = c.parent_id
WHERE c.parent_id <> 0 AND (pc.comment_id IS NULL OR pc.post_id <> c.post_id);
SELECT '评论时间早于所属帖子' AS `检查项`, COUNT(*) AS `实际数量`, 0 AS `期望值`
FROM `comment` c JOIN `post` p ON p.post_id = c.post_id WHERE c.create_time < p.create_time;

-- 帖子时间分布(信息性查询): 一眼看出数据是不是从最近铺到了一年前
SELECT '最近 1 天' AS `时间范围`, COUNT(*) AS `帖子数` FROM `post` WHERE create_time >= @t_now - INTERVAL 1 DAY
UNION ALL SELECT '最近 7 天(可投票)', COUNT(*) FROM `post` WHERE create_time >= @t_now - INTERVAL 7 DAY
UNION ALL SELECT '7~30 天', COUNT(*) FROM `post` WHERE create_time < @t_now - INTERVAL 7 DAY AND create_time >= @t_now - INTERVAL 30 DAY
UNION ALL SELECT '30~180 天', COUNT(*) FROM `post` WHERE create_time < @t_now - INTERVAL 30 DAY AND create_time >= @t_now - INTERVAL 180 DAY
UNION ALL SELECT '180 天以上', COUNT(*) FROM `post` WHERE create_time < @t_now - INTERVAL 180 DAY;

-- 每个版块的帖子数(信息性查询)
SELECT c.community_name AS `版块`, COUNT(p.post_id) AS `帖子数`
FROM `community` c LEFT JOIN `post` p ON p.community_id = c.community_id
GROUP BY c.community_id, c.community_name ORDER BY c.community_id;

-- 别忘了把这些数据灌进 Redis(首页榜单/票数/评论数都在 Redis 里):
--   cd inkwell_backend && go run ./cmd/seed -conf ./conf/config.yaml
