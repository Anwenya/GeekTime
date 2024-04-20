CREATE TABLE `follow_statics`
(
    `id`          bigint AUTO_INCREMENT,
    `uid`         bigint UNIQUE,
    `followers`   bigint,
    `followees`   bigint,
    `update_time` bigint,
    `create_time` bigint,
    PRIMARY KEY (`id`)
)

CREATE TABLE `follow_relations`
(
    `id`          bigint AUTO_INCREMENT,
    `follower`    bigint,
    `followee`    bigint,
    `status`      tinyint unsigned,
    `create_time` bigint,
    `update_time` bigint,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `follower_followee` (`follower`,`followee`)
)
