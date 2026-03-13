CREATE TABLE `sys_user_passkey`
(
    `id`                    int unsigned NOT NULL AUTO_INCREMENT,
    `user_id`               int unsigned NOT NULL,
    `credential_id`         varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
    `credential_public_key` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
    `sign_count`            int unsigned DEFAULT 0,
    `aaguid`                varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
    `transports_json`       json DEFAULT NULL,
    `user_handle`           varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
    `display_name`          varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
    `last_used_at`          timestamp DEFAULT NULL,
    `created_at`            timestamp DEFAULT CURRENT_TIMESTAMP,
    `updated_at`            timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`            timestamp DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_sys_user_passkey_credential_id` (`credential_id`),
    KEY `idx_sys_user_passkey_user_id` (`user_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;
