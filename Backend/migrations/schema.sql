-- =============================================================================
--  校内失物招领 · 建表脚本
--  对应代码：Backend/storage/model.go、Backend/Tags/Tags.go、Backend/auth/token.go
--  适用：MySQL 8.0+ / MariaDB 10.6+
--
--  用法（任选一种）：
--    命令行  : mysql -u<用户名> -p <库名> < Backend/migrations/schema.sql
--    图形界面: 用 Navicat / DataGrip / DBeaver 打开本文件，整段执行
--
--  ⚠ 第一步：库本身要先存在。后端默认 DSN 指向 demo 库，先手工建库并授权：
--      CREATE DATABASE IF NOT EXISTS demo DEFAULT CHARSET utf8mb4;
--      CREATE USER IF NOT EXISTS 'user_1'@'%' IDENTIFIED BY '123456';
--      GRANT ALL PRIVILEGES ON demo.* TO 'user_1'@'%';
--      FLUSH PRIVILEGES;
--    （用 root 执行；本机实测 user_1 目前只有 USAGE 权限、demo 库也不存在，
--     不改的话后端启动会直接 Fatal: Unknown database 'demo'）
--
--  设计说明：
--    1. 全部使用 CREATE TABLE IF NOT EXISTS —— 可以重复执行，不会删数据、不会报错。
--       本文件刻意不写 DROP TABLE，避免手滑把已有数据清空。
--    2. 表名以「代码实际用到」为准。注意是 users（复数），不是 README 里的 user。
--       原因：GORM 对 User 结构体默认推断复数表名 users，
--             且 auth/token.go:67 的 SQL 里硬编码了 JOIN users u —— 两处都要求 users。
--    3. 字符集统一 utf8mb4，否则中文能存但 emoji（前端标签图标）会报错。
--    4. 不建外键：项目里所有关联都是应用层维护（GORM 未开启迁移外键），
--       加外键反而会让「先建帖子再建评论」这类批量操作更容易失败。
-- =============================================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;


-- -----------------------------------------------------------------------------
-- 1. users —— 用户表
--    对应 storage.User（storage/model.go:8）
--
--    ⚠ creat_at 是代码里的拼写错误（不是 created_at），
--      Go 字段叫 CreatAt，GORM 推断出的列名就是 creat_at。
--      这里必须将错就错，改成 created_at 会直接导致注册时 Unknown column。
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `users` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `account_id`    VARCHAR(36)     DEFAULT NULL COMMENT 'UUID，注册时生成',
  `username`      VARCHAR(32)     DEFAULT NULL COMMENT '登录名，3-32 字符',
  `password_hash` VARCHAR(255)    DEFAULT NULL COMMENT 'bcrypt cost=12，绝不存明文',
  `bio`           TEXT            COMMENT '个人简介，NULL 表示未填写',
  `avatar_url`    VARCHAR(255)    DEFAULT NULL COMMENT '默认头像见 storage.DefaultImage',
  `creat_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '注意：不是 created_at',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_account_id` (`account_id`),
  UNIQUE KEY `uk_users_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户';


-- -----------------------------------------------------------------------------
-- 2. user_tokens —— 登录令牌表
--    对应 auth/token.go（没有 GORM 模型，全部是手写 SQL）
--
--    只存 SHA-256 摘要，不存明文令牌。令牌本身 32 字节随机数，
--    base64url 编码后发给客户端，有效期 4 小时（handle.AccessTokenExpire）。
--    revoked=1 表示已登出（软撤销，不删记录）。
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `user_tokens` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`    BIGINT UNSIGNED NOT NULL,
  `token_hash` VARCHAR(255)    NOT NULL COMMENT '明文令牌的 SHA-256 十六进制串',
  `expires_at` DATETIME        NOT NULL COMMENT '过期时间 = 签发时间 + 4h',
  `revoked`    TINYINT         NOT NULL DEFAULT 0 COMMENT '0=有效 1=已登出',
  `created_at` TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_tokens_hash` (`token_hash`),
  KEY `idx_user_tokens_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='登录令牌';


-- -----------------------------------------------------------------------------
-- 3. posts —— 帖子表
--    对应 storage.Post（storage/model.go:17）
--
--    tags 是逗号分隔的纯文本（如 "校园卡,钱包"），不是关联表。
--    标签库见 tags 表，写帖子时由 Tags.NormTag 自动去重/建标签。
--    comment_count 是冗余字段，发评论时 +1，用来生成楼层号（见 storage/comment.go）。
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `posts` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`       BIGINT UNSIGNED NOT NULL COMMENT '发帖人',
  `type`          TINYINT         NOT NULL COMMENT '1=寻物 2=招领',
  `title`         VARCHAR(100)    NOT NULL,
  `description`   TEXT            NOT NULL,
  `image_url`     VARCHAR(500)    DEFAULT NULL COMMENT '封面图，取 image_urls 的第一张',
  `tags`          VARCHAR(255)    DEFAULT NULL COMMENT '逗号分隔，如 校园卡,钱包',
  `location`      VARCHAR(255)    NOT NULL,
  `contact_info`  VARCHAR(100)    NOT NULL COMMENT '联系方式，列表接口不返回，仅详情接口给',
  `status`        TINYINT         NOT NULL DEFAULT 0 COMMENT '0=进行中 1=已结帖',
  `comment_count` INT UNSIGNED    NOT NULL DEFAULT 0 COMMENT '冗余评论数，同时充当楼层计数器',
  `created_at`    TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_posts_user` (`user_id`),
  KEY `idx_posts_type` (`type`),
  KEY `idx_posts_status` (`status`),
  KEY `idx_posts_created_at` (`created_at`),
  KEY `idx_posts_status_created` (`status`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='寻物/招领帖';


-- -----------------------------------------------------------------------------
-- 4. comments —— 评论表（建楼式）
--    对应 storage.Comment（storage/model.go:39）
--
--    floor = 楼层号，来自 posts.comment_count 的原子自增，
--    删评论是软删（status=1），所以楼层会留空号（1、2、4），这是刻意设计。
--    reply_to 引用的是「楼层号」而不是评论 id，这是前端必须注意的一点。
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `comments` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `post_id`    BIGINT UNSIGNED NOT NULL,
  `user_id`    BIGINT UNSIGNED NOT NULL,
  `floor`      INT UNSIGNED    NOT NULL COMMENT '楼层号，从 1 开始',
  `reply_to`   INT UNSIGNED    DEFAULT NULL COMMENT '引用第几楼，NULL=不引用',
  `content`    TEXT            NOT NULL,
  `image_url`  VARCHAR(500)    DEFAULT NULL COMMENT '封面图，取该评论第一张',
  `status`     TINYINT         NOT NULL DEFAULT 0 COMMENT '0=正常 1=已删除',
  `created_at` TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_comments_post_floor` (`post_id`, `floor`),
  KEY `idx_comments_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='评论'


-- -----------------------------------------------------------------------------
-- 5. post_images —— 图片表（帖子和评论共用）
--    对应 storage.PostImage（storage/model.go:57）
--
--    多态关联：target_type 1=帖子 2=评论，配合 target_id 指向具体对象。
--    写入用 storage.AddImages，读取用 storage.GetImages（当前还没有接口调用它）。
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `post_images` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `target_type` TINYINT         NOT NULL COMMENT '1=帖子 2=评论',
  `target_id`   BIGINT UNSIGNED NOT NULL COMMENT '对应 posts.id 或 comments.id',
  `url`         VARCHAR(500)    NOT NULL COMMENT '/uploads/... 或外链',
  `sort_order`  INT             NOT NULL DEFAULT 0 COMMENT '展示顺序，从 0 开始',
  `created_at`  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_post_images_target` (`target_type`, `target_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='图片（帖子+评论共用）';


-- -----------------------------------------------------------------------------
-- 6. tags —— 标签库
--    对应 Tags.Tag（Tags/Tags.go:36）
--
--    ⚠ 这张表 README 里漏了，缺了它 GET /api/tags 会直接 500，
--      而前端「发布」页必须选一个标签才能提交 —— 等于整个发帖功能不可用。
--    标签名规则：非空、不含逗号、≤32 字符（Tags.MaxTagLength）。
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `tags` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name`       VARCHAR(32)     NOT NULL,
  `created_at` TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_tags_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='标签库，发帖时自动扩充';


-- -----------------------------------------------------------------------------
-- 7. 初始标签数据
--    前端 Frontend/app.js 的 TAG_EMOJI 就是这批名字，一一对应。
--    没有这批数据时，发布页标签区是空的，用户选不到标签 → 发不了帖。
--    INSERT IGNORE：已存在的标签跳过，可重复执行。
-- -----------------------------------------------------------------------------
INSERT IGNORE INTO `tags` (`name`) VALUES
  ('校园卡'), ('身份证证件'), ('钥匙'), ('耳机'), ('手机数码'),
  ('充电宝'), ('雨伞'), ('水杯'), ('书籍笔记'), ('眼镜'),
  ('钱包'), ('书包'), ('笔袋文具'), ('手表首饰'), ('衣物外套'),
  ('篮球器材'), ('球拍'), ('化妆品');

SET FOREIGN_KEY_CHECKS = 1;


-- =============================================================================
--  以下是「已有数据库」的升级脚本，新库不用执行。
--  每条都带判断说明，执行前先确认自己没有对应的表/列。
-- =============================================================================

-- ① 如果你的库现在是单数 user（README 里的老 dump），必须先改名，否则注册/登录全挂
--    （GORM 查 users、auth/token.go 也 JOIN users）
--    确认：SHOW TABLES LIKE 'user%';
-- RENAME TABLE `user` TO `users`;

-- ② 老版本的 posts 没有 comment_count（评论楼层号依赖它）
--    确认：SHOW COLUMNS FROM posts LIKE 'comment_count';
-- ALTER TABLE `posts` ADD COLUMN `comment_count` INT UNSIGNED NOT NULL DEFAULT 0 AFTER `status`;

-- ③ 只想快速验证是否建齐（应输出 6 张表：users / user_tokens / posts / comments / post_images / tags）
-- SHOW TABLES;

-- ④ 查看标签库是否已灌入
-- SELECT * FROM tags ORDER BY name;
