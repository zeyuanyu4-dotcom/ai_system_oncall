-- AI Analysis Tasks 表
-- 创建时间: 2026-06-04

CREATE TABLE IF NOT EXISTS `ai_analysis_tasks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `issue_id` bigint unsigned NOT NULL,
  `status` varchar(32) NOT NULL DEFAULT 'pending' COMMENT 'pending/running/completed/failed/cancelled',
  `progress` varchar(32) DEFAULT NULL COMMENT '当前步骤/总步骤，如 3/8',
  `current_step` varchar(255) DEFAULT NULL COMMENT '当前执行步骤描述',
  `tool_calls` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'JSON，工具调用记录',
  `result` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'JSON，最终分析结果',
  `error_message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '错误信息',
  `started_at` datetime(3) DEFAULT NULL,
  `completed_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_issue_id` (`issue_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI分析任务表';
