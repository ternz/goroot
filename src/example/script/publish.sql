create database publish character set utf8;
use publish;

create table `history_opt` (
	`id` bigint unsigned auto_increment,
	`user_id` varchar(255) not null,
	`content` varchar(2000) not null,
	`create_time` bigint unsigned not null,
	primary key (`id`)
);
	