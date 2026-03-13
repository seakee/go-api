CREATE TABLE `sys_user_identity`
(
    `id`               int unsigned NOT NULL AUTO_INCREMENT,
    `user_id`          int unsigned NOT NULL,
    `provider`         varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci  NOT NULL,
    `provider_tenant`  varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
    `provider_subject` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
    `display_name`     varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
    `avatar_url`       text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci,
    `raw_profile_json` json                                                        DEFAULT NULL,
    `bound_at`         timestamp                                                   DEFAULT CURRENT_TIMESTAMP,
    `last_login_at`    timestamp                                                   DEFAULT NULL,
    `created_at`       timestamp                                                   DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       timestamp                                                   DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`       timestamp                                                   DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_sys_user_identity_provider_subject` (`provider`, `provider_tenant`, `provider_subject`),
    KEY `idx_sys_user_identity_user_id` (`user_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;
