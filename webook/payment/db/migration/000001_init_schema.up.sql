CREATE TABLE `payments`
(
    `id`           bigint AUTO_INCREMENT,
    `amount`       bigint,
    `currency`     longtext,
    `description`  longtext,
    `biz_trade_no` varchar(256) UNIQUE,
    `txn_id`       varchar(128) UNIQUE,
    `status`       tinyint unsigned,
    `update_time`  bigint,
    `create_time`  bigint,
    PRIMARY KEY (`id`)
)
