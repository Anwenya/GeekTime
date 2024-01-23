CREATE TABLE `users` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `email` varchar(255) UNIQUE,
  `password` varchar(255),
  `nickname` varchar(128),
  `birthday` bigint,
  `bio` varchar(4096),
  `create_time` bigint,
  `update_time` bigint
);