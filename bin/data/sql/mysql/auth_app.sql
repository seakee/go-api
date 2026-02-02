CREATE TABLE `auth_app`
(
    `id`           int                                                           NOT NULL AUTO_INCREMENT COMMENT 'id',
    `app_id`       varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci  NOT NULL COMMENT 'app ID',
    `app_name`     varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci           DEFAULT NULL COMMENT 'app name',
    `app_secret`   varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'app credential secret',
    `redirect_uri` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci          DEFAULT NULL COMMENT 'redirect callback URL after authorization',
    `description`  text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT 'description',
    `status`       tinyint(1)                                                    NOT NULL DEFAULT '0' COMMENT '0: disabled; 1: active; 2: suspended',
    `created_at`   timestamp                                                     NULL     DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   timestamp                                                     NULL     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`   timestamp                                                     NULL     DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `app_id` (`app_id`),
    KEY `app_name` (`app_name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT ='client application info table';
