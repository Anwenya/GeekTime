CREATE TABLE `rewards`
(
    `id`          bigint AUTO_INCREMENT,
    `biz`         varchar(191),
    `biz_id`      bigint,
    `biz_name`    longtext,
    `target_uid`  bigint,
    `status`      tinyint unsigned,
    `uid`         bigint,
    `amount`      bigint,
    `create_time` bigint,
    `update_time` bigint,
    PRIMARY KEY (`id`),
    INDEX         `biz_biz_id` (`biz`,`biz_id`),
    INDEX         `idx_rewards_target_uid` (`target_uid`)
)
