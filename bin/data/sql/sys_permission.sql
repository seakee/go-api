CREATE TABLE `sys_permission`
(
    `id`          int unsigned NOT NULL AUTO_INCREMENT,
    `name`        varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
    `type`        varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
    `method`      varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
    `path`        text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci,
    `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci,
    `group`       text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci,
    `created_at`  timestamp    NULL                                            DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  timestamp    NULL                                            DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`  timestamp    NULL                                            DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;