CREATE TABLE `auth_app`
(
    `id`           int(11) NOT NULL AUTO_INCREMENT COMMENT 'id',
    `app_id`       varchar(30)  NOT NULL DEFAULT '' COMMENT '应用ID',
    `app_name`     varchar(50)           DEFAULT NULL COMMENT '应用名称',
    `app_secret`   varchar(255) NOT NULL DEFAULT '' COMMENT '应用的凭证密钥',
    `redirect_uri` varchar(500)          DEFAULT NULL COMMENT '授权后重定向的回调链接地址',
    `description`  varchar(1000)         DEFAULT NULL COMMENT '描述信息',
    `status`       tinyint(1) NOT NULL DEFAULT '0' COMMENT '0表示未开通；1表示正常使用；2表示已被禁用',
    `created_at`   timestamp NULL DEFAULT NULL,
    `updated_at`   timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`   timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='接入的客户端信息表';