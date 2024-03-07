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

CREATE TABLE `async_sms_tasks`
(
    `id`          bigint AUTO_INCREMENT,
    `config`      longtext,
    `retry_cnt`   bigint,
    `retry_max`   bigint,
    `status`      tinyint unsigned,
    `create_time` bigint,
    `update_time` bigint,
    PRIMARY KEY (`id`),
    INDEX `idx_async_sms_tasks_update_time` (`update_time`)
);

CREATE TABLE `user_like_bizs`
(
    `id`          bigint AUTO_INCREMENT,
    `uid`         bigint,
    `biz_id`      bigint,
    `biz`         varchar(128),
    `status`      bigint,
    `update_time` bigint,
    `create_time` bigint,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `uid_biz_type_id` (`uid`, `biz_id`, `biz`)
);

CREATE TABLE `user_collection_bizs`
(
    `id`          bigint AUTO_INCREMENT,
    `uid`         bigint,
    `biz_id`      bigint,
    `biz`         varchar(128),
    `cid`         bigint,
    `update_time` bigint,
    `create_time` bigint,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `uid_biz_type_id` (`uid`, `biz_id`, `biz`),
    INDEX `idx_user_collection_bizs_cid` (`cid`)
);

CREATE TABLE `interactives`
(
    `id`          bigint AUTO_INCREMENT,
    `biz_id`      bigint,
    `biz`         varchar(128),
    `read_cnt`    bigint,
    `like_cnt`    bigint,
    `collect_cnt` bigint,
    `update_time` bigint,
    `create_time` bigint,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `biz_type_id` (`biz_id`, `biz`)
);

CREATE TABLE `read_histories`
(
    `id`          bigint AUTO_INCREMENT,
    `uid`         bigint,
    `biz_id`      bigint,
    `biz`         varchar(128),
    `read_time`   bigint,
    `create_time` bigint,
    `update_time` bigint,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `uid_biz_type_id` (`uid`, `biz_id`, `biz`)
);