CREATE TABLE `comments`
(
    `id`          bigint AUTO_INCREMENT,
    `uid`         bigint,
    `biz`         varchar(191),
    `biz_id`      bigint,
    `content`     longtext,
    `root_id`     bigint,
    `pid`         bigint,
    `create_time` bigint,
    `update_time` bigint,
    PRIMARY KEY (`id`),
    INDEX         `biz_type_id` (`biz`,`biz_id`),
    INDEX         `idx_comments_root_id` (`root_id`),
    INDEX         `idx_comments_p_id` (`pid`),
    CONSTRAINT `fk_comments_parent_comment` FOREIGN KEY (`pid`) REFERENCES `comments` (`id`) ON DELETE CASCADE
)
