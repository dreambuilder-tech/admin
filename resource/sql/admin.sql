-- ============================================
-- Admin 数据库表结构
-- ============================================

USE `admin`;

-- 角色配置表
DROP TABLE IF EXISTS `roles`;
CREATE TABLE `roles` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '角色名称',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `idx_name` (`name`) USING BTREE
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='角色配置表';

-- 插入初始角色数据
INSERT INTO `roles` (`name`) VALUES
('super_admin'),
('admin'),
('operation'),
('support'),
('finance');

-- 用户表
DROP TABLE IF EXISTS `admins`;
CREATE TABLE `admins` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `account` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '账号',
    `password` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '密码',
    `role_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '角色ID',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 0=禁用，1=启用',
    `last_login_at` DATETIME NULL COMMENT '最近登录时间',
    `last_login_ip` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '最近登录IP',
    `mfa_secret` VARBINARY(255) NULL COMMENT '谷歌验证器密钥',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `idx_account` (`account`) USING BTREE,
    KEY `idx_role_id` (`role_id`) USING BTREE
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='管理员用户表';

-- 用户权限表
DROP TABLE IF EXISTS `admin_perms`;
CREATE TABLE `admin_perms` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `uid` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `perms` JSON NOT NULL COMMENT '权限列表JSON数组',
    `created_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    `updated_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间戳',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `idx_uid` (`uid`) USING BTREE
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='用户权限表';

