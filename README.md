# NCUHOME-MiniHack-TeamNewFolder

//user建表信息
+---------------+-----------------+------+-----+-------------------+-------------------+
| Field         | Type            | Null | Key | Default           | Extra             |
+---------------+-----------------+------+-----+-------------------+-------------------+
| id            | bigint unsigned | NO   | PRI | NULL              | auto_increment    |
| account_id    | varchar(36)     | YES  | UNI | NULL              |                   |
| username      | varchar(32)     | YES  | UNI | NULL              |                   |
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
备注：user_tokens.user_id是user.id