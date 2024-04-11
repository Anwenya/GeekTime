CREATE TABLE `accounts`
(
    `id`          bigint AUTO_INCREMENT,
    `uid`         bigint,
    `account`     bigint,
    `type`        tinyint unsigned,
    `balance`     bigint,
    `currency`    longtext,
    `update_time` bigint,
    `create_time` bigint,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `account_type` (`account`,`type`)
);

CREATE TABLE `account_activities`
(
    `id`           bigint AUTO_INCREMENT,
    `uid`          bigint,
    `biz`          varchar(32),
    `biz_id`       bigint,
    `account`      bigint,
    `account_type` tinyint unsigned,
    `amount`       bigint,
    `currency`     longtext,
    `update_time`  bigint,
    `create_time`  bigint,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `biz_type_id` (`biz`,`biz_id`,`account`,`account_type`),
    INDEX          `account_type` (`account`,`account_type`)
);