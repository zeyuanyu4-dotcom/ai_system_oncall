-- 智能OnCall问题处理与研发协同平台数据库初始化脚本
-- 创建数据库
CREATE DATABASE IF NOT EXISTS oncall_platform DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE oncall_platform;

-- 用户表
CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `username` VARCHAR(64) NOT NULL,
    `email` VARCHAR(128) NOT NULL,
    `password_hash` VARCHAR(255) NOT NULL,
    `phone` VARCHAR(32) DEFAULT NULL,
    `role` VARCHAR(32) NOT NULL DEFAULT 'normal_user',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1:正常, 0:禁用',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_email` (`email`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 项目表
CREATE TABLE IF NOT EXISTS `projects` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(128) NOT NULL,
    `code` VARCHAR(64) NOT NULL,
    `description` TEXT,
    `owner_id` BIGINT UNSIGNED DEFAULT NULL,
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1:正常, 0:停用',
    `created_by` BIGINT UNSIGNED DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_code` (`code`),
    KEY `idx_owner_id` (`owner_id`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目表';

-- 项目成员表
CREATE TABLE IF NOT EXISTS `project_members` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `project_id` BIGINT UNSIGNED NOT NULL,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `project_role` VARCHAR(32) NOT NULL DEFAULT 'member',
    `joined_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_project_user` (`project_id`, `user_id`),
    KEY `idx_project_id` (`project_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目成员表';

-- 服务表
CREATE TABLE IF NOT EXISTS `services` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `project_id` BIGINT UNSIGNED NOT NULL,
    `name` VARCHAR(128) NOT NULL,
    `code` VARCHAR(64) NOT NULL,
    `description` TEXT,
    `service_type` VARCHAR(32) NOT NULL DEFAULT 'backend',
    `owner_id` BIGINT UNSIGNED DEFAULT NULL,
    `language` VARCHAR(32) DEFAULT NULL,
    `repo_url` VARCHAR(255) DEFAULT NULL,
    `deploy_env` VARCHAR(64) DEFAULT NULL,
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1:正常, 0:下线',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_project_code` (`project_id`, `code`),
    KEY `idx_project_id` (`project_id`),
    KEY `idx_owner_id` (`owner_id`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务表';

-- 服务接口表
CREATE TABLE IF NOT EXISTS `service_apis` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `service_id` BIGINT UNSIGNED NOT NULL,
    `method` VARCHAR(16) NOT NULL,
    `path` VARCHAR(255) NOT NULL,
    `name` VARCHAR(128) DEFAULT NULL,
    `description` TEXT,
    `status` TINYINT NOT NULL DEFAULT 1,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_service_method_path` (`service_id`, `method`, `path`),
    KEY `idx_service_id` (`service_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务接口表';

-- 服务依赖表
CREATE TABLE IF NOT EXISTS `service_dependencies` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `service_id` BIGINT UNSIGNED NOT NULL,
    `depends_on_service_id` BIGINT UNSIGNED NOT NULL,
    `dependency_type` VARCHAR(32) NOT NULL DEFAULT 'http',
    `description` TEXT,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_service_id` (`service_id`),
    KEY `idx_depends_on_service_id` (`depends_on_service_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务依赖表';

-- 问题单表
CREATE TABLE IF NOT EXISTS `issues` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `issue_no` VARCHAR(64) NOT NULL,
    `title` VARCHAR(255) NOT NULL,
    `description` TEXT,
    `project_id` BIGINT UNSIGNED NOT NULL,
    `service_id` BIGINT UNSIGNED DEFAULT NULL,
    `issue_type` VARCHAR(32) NOT NULL DEFAULT 'other',
    `priority` VARCHAR(16) NOT NULL DEFAULT 'P2',
    `environment` VARCHAR(32) DEFAULT NULL,
    `status` VARCHAR(32) NOT NULL DEFAULT 'pending_analysis',
    `impact_scope` VARCHAR(255) DEFAULT NULL,
    `error_message` TEXT,
    `log_excerpt` TEXT,
    `creator_id` BIGINT UNSIGNED NOT NULL,
    `assignee_id` BIGINT UNSIGNED DEFAULT NULL,
    `ai_summary` TEXT,
    `ai_analysis` TEXT,
    `root_cause` TEXT,
    `solution` TEXT,
    `resolved_at` DATETIME DEFAULT NULL,
    `closed_at` DATETIME DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_issue_no` (`issue_no`),
    KEY `idx_project_id` (`project_id`),
    KEY `idx_service_id` (`service_id`),
    KEY `idx_status` (`status`),
    KEY `idx_creator_id` (`creator_id`),
    KEY `idx_assignee_id` (`assignee_id`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='问题单表';

-- 问题评论表
CREATE TABLE IF NOT EXISTS `issue_comments` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `issue_id` BIGINT UNSIGNED NOT NULL,
    `user_id` BIGINT UNSIGNED DEFAULT NULL,
    `comment_type` VARCHAR(32) NOT NULL DEFAULT 'comment',
    `content` TEXT NOT NULL,
    `visibility` VARCHAR(32) NOT NULL DEFAULT 'public',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_issue_id` (`issue_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='问题评论表';

-- 状态流转日志表
CREATE TABLE IF NOT EXISTS `issue_status_logs` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `issue_id` BIGINT UNSIGNED NOT NULL,
    `from_status` VARCHAR(32) DEFAULT NULL,
    `to_status` VARCHAR(32) NOT NULL,
    `operator_id` BIGINT UNSIGNED NOT NULL,
    `reason` VARCHAR(255) DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_issue_id` (`issue_id`),
    KEY `idx_operator_id` (`operator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='状态流转日志表';

-- 操作日志表
CREATE TABLE IF NOT EXISTS `issue_operation_logs` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `issue_id` BIGINT UNSIGNED NOT NULL,
    `operator_id` BIGINT UNSIGNED NOT NULL,
    `operation_type` VARCHAR(64) NOT NULL,
    `operation_content` TEXT,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_issue_id` (`issue_id`),
    KEY `idx_operator_id` (`operator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';

-- 插入默认系统管理员
INSERT INTO `users` (`username`, `email`, `password_hash`, `role`, `status`) VALUES
('admin', 'admin@oncall.com', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z5EHsM8lE9lBOsl7iAt6Z5EH', 'system_admin', 1);
