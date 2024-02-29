CREATE TABLE `users`
(
    `id`              bigint PRIMARY KEY AUTO_INCREMENT,
    `email`           varchar(255) UNIQUE,
    `password`        varchar(255),
    `nickname`        varchar(128),
    `birthday`        bigint,
    `bio`             longtext,
    `phone`           varchar(16) UNIQUE,
    `wechat_open_id`  varchar(32) UNIQUE,
    `wechat_union_id` varchar(32),
    `create_time`     bigint,
    `update_time`     bigint
);


CREATE TABLE `articles`
(
    `id`          bigint AUTO_INCREMENT,
    `title`       longtext,
    `content`     longtext,
    `author_id`   bigint,
    `status`      tinyint unsigned,
    `create_time` bigint,
    `update_time` bigint,
    PRIMARY KEY (`id`),
    INDEX `idx_articles_author_id` (`author_id`)
);


CREATE TABLE `published_articles`
(
    `id`          bigint AUTO_INCREMENT,
    `title`       longtext,
    `content`     longtext,
    `author_id`   bigint,
    `status`      tinyint unsigned,
    `create_time` bigint,
    `update_time` bigint,
    PRIMARY KEY (`id`),
    INDEX `idx_published_articles_author_id` (`author_id`)
);