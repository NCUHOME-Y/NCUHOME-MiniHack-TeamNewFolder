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