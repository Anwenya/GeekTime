CREATE TABLE `users` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `email` varchar(255) UNIQUE,
  `password` varchar(255),
  `nickname` varchar(128),
  `birthday` bigint,
  `bio` varchar(4096),
  `phone` varchar(16) UNIQUE,
  `wechat_open_id` varchar(32) UNIQUE,
  `wechat_union_id` varchar(32),
  `create_time` bigint,
  `update_time` bigint
);