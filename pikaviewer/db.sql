USE `game`;

CREATE TABLE `admin_user` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `username` varchar(32) NOT NULL,
    `password` char(32) NOT NULL,
    `role` int NOT NULL,
    `status` int NOT NULL,
    `created_at` bigint NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY unique_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='gm用户表';

CREATE TABLE `admin_mail_log` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `admin_user_id` bigint NOT NULL,
    `username` varchar(32) NOT NULL,
    `type` int NOT NULL, # 1.新增邮件，2.新增全服邮件，3.删除邮件，4.删除全服邮件
    `uid` text,
    `mail` text,
    `created_at` bigint NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='mail操作记录表';

CREATE TABLE `announcement` (
    `id` varchar(32) NOT NULL,
    `type` varchar(16) NOT NULL,
    `version` varchar(16) NOT NULL,
    `begin_time` bigint NOT NULL,
    `end_time` bigint NOT NULL,
    `store_url` varchar(64),
    `content` text,
    `show_login` bool,
    `created_at` bigint NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `type_version_idx` (`type`,`version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='公告';

CREATE TABLE `gitlab_user` (
    `id` int (11) NOT NULL,
    `username` varchar(64) NOT NULL,
    `access_level` int(11) NOT NULL,
    PRIMARY KEY (`id`),
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='gitlab用户';

CREATE TABLE  `cd_key_gen` (
    `id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
    `batch_id` int (11) NOT NULL,
    `begin_at` bigint NOT NULL,
    `end_at` bigint NOT NULL,
    `limit_cnt` int (11) NOT NULL,
    `limit_typ` int NOT NULL, # 1.多人单次,2.单人单次,3.多人单次且限定次数
    `reward` text,
    `is_active` bool,
    `use_cnt` int (11) NOT NULL,
    `created_at` bigint NOT NULL,
    PRIMARY KEY (`id`),
    INDEX `batch_idx` (`batch_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='兑换码生成';

CREATE TABLE `cd_key_use` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `role_id` varchar(32) NOT NULL,
    `batch_id` int (11) NOT NULL,
    `key_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
    `created_at` bigint NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `role_batch_idx` (`role_id`,`batch_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='兑换码兑换记录';


CREATE TABLE `pay` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `role_id` varchar(64) NULL,
    `sn` varchar(64) UNIQUE NOT NULL,
    `pc_id` int NOT NULL,
    `igg_id` varchar(32) NOT NULL,
    `paid_time` bigint NOT NULL,
    `created_at` bigint NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付表';


CREATE TABLE `notice` (
    `id` varchar(20) NOT NULL,
    `title` varchar(64) NULL,
    `content` text NOT NULL,
    `reward_content` text,
    `is_custom` boolean DEFAULT false,
    `begin_at` bigint NOT NULL,
    `expired_at` bigint NOT NULL,
    `created_at` bigint NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='游戏内公告';