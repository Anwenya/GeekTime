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
    INDEX         `idx_user_collection_bizs_cid` (`cid`)
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

