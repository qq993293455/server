CREATE DATABASE `im`;

USE `im`;
-- im.broadcast_message definition

CREATE TABLE `broadcast_message` (
    `mid` varchar(20) NOT NULL COMMENT '唯一标识',
    `type` int NOT NULL DEFAULT 0,
    `role_id` varchar(20) NOT NULL DEFAULT '0',
    `role_name` varchar(128) NOT NULL DEFAULT '0',
    `content` varchar(1024) NOT NULL DEFAULT '0',
    `parse_type` int NOT NULL DEFAULT 0,
    `created_at` bigint NOT NULL COMMENT '创建时间' DEFAULT 0,
    `extra` text COMMENT '额外信息',
    PRIMARY KEY (`mid`),
    KEY `role_id_idx` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='世界聊天';

-- im.private_message definition

CREATE TABLE `private_message` (
    `mid` varchar(20) NOT NULL COMMENT '唯一标识',
    `type` int NOT NULL DEFAULT 0,
    `role_id` varchar(20) NOT NULL DEFAULT '0',
    `role_name` varchar(128) NOT NULL DEFAULT '0',
    `target_id` varchar(20) NOT NULL DEFAULT '0',
    `content` varchar(1024) NOT NULL DEFAULT '0',
    `parse_type` int NOT NULL DEFAULT '0',
    `created_at` bigint NOT NULL COMMENT '创建时间' DEFAULT 0,
    `extra` text COMMENT '额外信息',
    PRIMARY KEY (`mid`),
    KEY `role_id_idx` (`role_id`),
    KEY `target_id_idx` (`target_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='私聊';

-- im.role_rooms definition

CREATE TABLE `role_rooms` (
    `role_id` varchar(20) NOT NULL COMMENT '唯一标识',
    `rooms` text,
    PRIMARY KEY (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户的房间';

-- im.room_message definition

CREATE TABLE `room_message` (
    `mid` varchar(20) NOT NULL COMMENT '唯一标识',
    `type` int NOT NULL DEFAULT 0,
    `role_id` varchar(20) NOT NULL DEFAULT '0',
    `role_name` varchar(128) NOT NULL DEFAULT '0',
    `target_id` varchar(20) NOT NULL DEFAULT '0',
    `content` varchar(1024) NOT NULL DEFAULT '0',
    `parse_type` int NOT NULL DEFAULT '0',
    `created_at` bigint NOT NULL COMMENT '创建时间',
    `extra` text COMMENT '额外信息',
    PRIMARY KEY (`mid`),
    KEY `role_id_idx` (`role_id`),
    KEY `target_id_idx` (`target_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群聊';


CREATE DATABASE `rank`;

USE `rank`;
-- `rank`.`rank` definition

CREATE TABLE `rank` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `rank_id` varchar(32) NOT NULL COMMENT '排行榜id',
    `rank_type` bigint NOT NULL,
    `owner_id` varchar(32) NOT NULL,
    `value1` bigint DEFAULT '0',
    `value2` bigint DEFAULT '0',
    `value3` bigint DEFAULT '0',
    `value4` bigint DEFAULT '0',
    `extra` text,
    `shard` bigint DEFAULT '0',
    `created_at` bigint DEFAULT '0',
    `deleted_at` bigint DEFAULT '0',
    PRIMARY KEY (`id`),
    UNIQUE KEY `rank_id_owner_id_idx` (`rank_id`,`owner_id`),
    KEY `rank_type_idx` (`rank_type`),
    KEY `deleted_at_idx` (`deleted_at`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4  COMMENT='排行榜';


CREATE DATABASE `game`;

USE `game`;

CREATE TABLE `admin_user` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `username` varchar(32) NOT NULL DEFAULT '0',
    `password` char(32) NOT NULL DEFAULT '0',
    `role` int NOT NULL DEFAULT 0,
    `status` int NOT NULL DEFAULT 0,
    `created_at` bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY unique_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='gm用户表';

CREATE TABLE `admin_mail_log` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `admin_user_id` bigint NOT NULL DEFAULT 0,
    `username` varchar(32) NOT NULL DEFAULT '0',
    `type` int NOT NULL DEFAULT 0, # 1.新增邮件，2.新增全服邮件，3.删除邮件，4.删除全服邮件
    `uid` text,
    `mail` text,
    `created_at` bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='mail操作记录表';

CREATE TABLE `announcement` (
    `id` varchar(32) NOT NULL DEFAULT '0',
    `type` varchar(16) NOT NULL DEFAULT '0',
    `version` varchar(16) NOT NULL DEFAULT '0',
    `begin_time` bigint NOT NULL DEFAULT 0,
    `end_time` bigint NOT NULL DEFAULT 0,
    `store_url` varchar(256),
    `content` text,
    `show_login` bool,
    `created_at` bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `type_version_idx` (`type`,`version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公告';

CREATE TABLE `gitlab_user` (
    `id` int (11) NOT NULL DEFAULT 0,
    `username` varchar(64) NOT NULL DEFAULT '0',
    `access_level` int(11) NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='gitlab用户';

CREATE TABLE  `cd_key_gen` (
    `id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '0',
    `batch_id` int (11) NOT NULL DEFAULT 0,
    `begin_at` bigint NOT NULL DEFAULT 0,
    `end_at` bigint NOT NULL DEFAULT 0,
    `limit_cnt` int (11) NOT NULL DEFAULT 0,
    `limit_typ` int NOT NULL DEFAULT 0, # 1.多人单次,2.单人单次,3.多人单次且限定次数
    `reward` text,
    `is_active` bool,
    `use_cnt` int (11) NOT NULL DEFAULT 0,
    `created_at` bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    INDEX `batch_idx` (`batch_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='兑换码生成';

CREATE TABLE `cd_key_use` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `role_id` varchar(32) NOT NULL DEFAULT '0',
    `batch_id` int (11) NOT NULL DEFAULT 0,
    `key_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '0',
    `created_at` bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `role_batch_idx` (`role_id`,`batch_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='兑换码兑换记录';

CREATE TABLE `racing_rank` (
   `id` bigint NOT NULL AUTO_INCREMENT,
   `role_id` varchar(32) NOT NULL DEFAULT '0',
   `end_time` bigint NOT NULL COMMENT '结算时间' DEFAULT 0,
   PRIMARY KEY (`id`),
   UNIQUE KEY `unique_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='竞速赛';

CREATE TABLE `roles` (
     `role_id` varchar(20) NOT NULL COMMENT '唯一标识',
     `igg_id` varchar(32) DEFAULT '0' COMMENT 'iggid',
     `nickname` varchar(512) NOT NULL DEFAULT '0',
     `level` int NOT NULL DEFAULT 0,
     `avatar` int NOT NULL DEFAULT 0,
     `avatar_frame` int NOT NULL DEFAULT 0,
     `power` bigint NOT NULL DEFAULT 0,
     `title` int NOT NULL DEFAULT 0,
     `language` int NOT NULL DEFAULT 0,
     `create_time` bigint NOT NULL COMMENT '创建时间' DEFAULT 0,
     `login_time` bigint DEFAULT '0',
     `logout_time` bigint DEFAULT '0',
     `highest_power` bigint DEFAULT '0',
     PRIMARY KEY (`role_id`),
     KEY `level_idx` (`level`),
     KEY `power_idx` (`power`),
     KEY `highest_power_index` (`highest_power`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='角色信息';

CREATE TABLE `pay` (
   `id` bigint(20) NOT NULL AUTO_INCREMENT,
   `role_id` varchar(64) NULL DEFAULT '0',
   `sn` varchar(64) UNIQUE NOT NULL DEFAULT '0',
   `pc_id` int NOT NULL DEFAULT 0,
   `igg_id` varchar(32) NOT NULL DEFAULT '0',
   `paid_time` bigint NOT NULL DEFAULT 0,
   `created_at` bigint NOT NULL DEFAULT 0,
   PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付表';

CREATE TABLE `notice` (
  `id` varchar(20) NOT NULL DEFAULT '0',
  `title` varchar(64) NULL DEFAULT '0',
  `content` text,
  `reward_content` text,
  `is_custom` boolean DEFAULT false,
  `begin_at` bigint NOT NULL DEFAULT 0,
  `expired_at` bigint NOT NULL DEFAULT 0,
  `created_at` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='游戏内公告';

CREATE TABLE `fashion` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `role_id` varchar(64) NOT NULL DEFAULT '0',
    `user_id` varchar(64) NOT NULL DEFAULT '0',
    `server_id` bigint NOT NULL DEFAULT 0,
    `hero_id` bigint NOT NULL DEFAULT 0,
    `fashion_id` bigint NOT NULL DEFAULT 0,
    `expired_at` bigint NOT NULL DEFAULT 0,
    `created_at` bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    KEY `role_fashion_idx` (`role_id`,`fashion_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='时装';

CREATE TABLE `guild` (
     `id` varchar(32) NOT NULL DEFAULT '0',
     `name` varchar(32) NOT NULL DEFAULT '0',
     PRIMARY KEY (`id`),
     UNIQUE KEY `unique_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公会名称表';

CREATE DATABASE `log`;