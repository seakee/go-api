CREATE TABLE `sys_user`
(
    `id`           int unsigned NOT NULL AUTO_INCREMENT,
    `account`      varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
    `password`     char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci     DEFAULT NULL,
    `salt`         char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci                                                     DEFAULT NULL,
    `totp_key`     char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci     DEFAULT NULL,
    `totp_enabled` tinyint(1)                                                    DEFAULT NULL,
    `feishu_id`    varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
    `wechat_id`    varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
    `github_id`    varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
    `user_name`    varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci  DEFAULT NULL,
    `status`       tinyint                                                       DEFAULT NULL,
    `avatar`       text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci,
    `created_at`   timestamp    NULL                                             DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   timestamp    NULL                                             DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`   timestamp    NULL                                             DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `account` (`account`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;