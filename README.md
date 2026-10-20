# NCUHOME-MiniHack-TeamNewFolder

//User信息存储
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