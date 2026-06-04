-- Create reports table for Phase 6: Report Generation System

CREATE TABLE IF NOT EXISTS `reports` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `report_type` varchar(32) NOT NULL COMMENT 'daily, weekly, incident',
  `report_date` varchar(16) DEFAULT NULL COMMENT '日期: 2026-06-04',
  `report_week` varchar(16) DEFAULT NULL COMMENT '周: 2026-W23',
  `title` varchar(255) NOT NULL,
  `summary` text,
  `content` longtext,
  `issue_count` int DEFAULT 0,
  `creator_id` bigint unsigned NOT NULL,
  `is_auto` tinyint(1) DEFAULT 0 COMMENT '是否自动生成',
  `status` varchar(32) NOT NULL DEFAULT 'generated' COMMENT 'generating, generated, failed',
  `error_msg` text,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_report_type` (`report_type`),
  KEY `idx_report_date` (`report_date`),
  KEY `idx_report_week` (`report_week`),
  KEY `idx_creator_id` (`creator_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='报告表';