# NCUHOME-MiniHack-TeamNewFolder

//user建表信息
+---------------+-----------------+------+-----+-------------------+-------------------+
| Field         | Type            | Null | Key | Default           | Extra             |
+---------------+-----------------+------+-----+-------------------+-------------------+
| id            | bigint unsigned | NO   | PRI | NULL              | auto_increment    |
| account_id    | varchar(36)     | YES  | UNI | NULL              |                   |
| username      | varchar(32)     | YES  | UNI | NULL              |                   |
| avatar_url    | varchar(255)    | YES  |     | NULL              |                   |
| password_hash | varchar(255)    | YES  |     | NULL              |                   |
| bio           | text            | YES  |     | NULL              |                   |
| creat_at      | datetime        | NO   |     | CURRENT_TIMESTAMP | DEFAULT_GENERATED |
+---------------+-----------------+------+-----+-------------------+-------------------+

//user_tokens建表信息
+------------+-----------------+------+-----+-------------------+-------------------+
| Field      | Type            | Null | Key | Default           | Extra             |
+------------+-----------------+------+-----+-------------------+-------------------+
| id         | bigint unsigned | NO   | PRI | NULL              | auto_increment    |
| user_id    | bigint unsigned | NO   | MUL | NULL              |                   |
| token_hash | varchar(255)    | NO   | UNI | NULL              |                   |
| expires_at | datetime        | NO   |     | NULL              |                   |
| revoked    | tinyint         | NO   |     | 0                 |                   |
| created_at | timestamp       | NO   |     | CURRENT_TIMESTAMP | DEFAULT_GENERATED |
+------------+-----------------+------+-----+-------------------+-------------------+

//posts建表信息
+--------------+-----------------+------+-----+-------------------+-----------------------------------------------+
| Field        | Type            | Null | Key | Default           | Extra                                         |
+--------------+-----------------+------+-----+-------------------+-----------------------------------------------+
| id           | bigint unsigned | NO   | PRI | NULL              | auto_increment                                |
| user_id      | bigint unsigned | NO   | MUL | NULL              |                                               |
| type         | tinyint         | NO   | MUL | NULL              |                                               |
| title        | varchar(100)    | NO   |     | NULL              |                                               |
| description  | text            | NO   |     | NULL              |                                               |
| image_url    | varchar(500)    | YES  |     | NULL              |                                               |
| tags         | varchar(255)    | YES  |     | NULL              |                                               |
| location     | varchar(255)    | NO   |     | NULL              |                                               |
| contact_info | varchar(100)    | NO   |     | NULL              |                                               |
| status       | tinyint         | NO   |     | 0                 |                                               |
| created_at   | timestamp       | NO   | MUL | CURRENT_TIMESTAMP | DEFAULT_GENERATED                             |
| updated_at   | timestamp       | NO   |     | CURRENT_TIMESTAMP | DEFAULT_GENERATED on update CURRENT_TIMESTAMP |
+--------------+-----------------+------+-----+-------------------+-----------------------------------------------+

//comment建表数据
-- 评论表：每条评论属于一个帖子，floor 是楼层号
CREATE TABLE comments (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  post_id    BIGINT UNSIGNED NOT NULL,
  user_id    BIGINT UNSIGNED NOT NULL,
  floor      INT    UNSIGNED NOT NULL COMMENT '楼层号，从 1 开始',
  reply_to   INT    UNSIGNED NULL     COMMENT '引用第几楼，NULL 表示不引用',
  content    TEXT         NOT NULL,
  image_url  VARCHAR(500) NULL        COMMENT '封面图，取该评论的第一张',
  status     TINYINT      NOT NULL DEFAULT 0 COMMENT '0正常 1已删除',
  created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_post_floor (post_id, floor),
  INDEX idx_user (user_id)
);

-- 图片表：帖子和评论共用，靠 target_type 区分
CREATE TABLE post_images (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  target_type TINYINT       NOT NULL COMMENT '1=帖子 2=评论',
  target_id   BIGINT UNSIGNED NOT NULL,
  url         VARCHAR(500)  NOT NULL,
  sort_order  INT           NOT NULL DEFAULT 0,
  created_at  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_target (target_type, target_id)
);

-- 给帖子加一个评论数（冗余字段，列表页直接显示，不用每次 COUNT）
ALTER TABLE posts ADD COLUMN comment_count INT UNSIGNED NOT NULL DEFAULT 0;