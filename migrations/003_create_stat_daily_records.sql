-- Create stat_daily_records table for Dashboard Statistics

CREATE TABLE IF NOT EXISTS `stat_daily_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `stat_date` varchar(16) NOT NULL COMMENT '统计日期 YYYY-MM-DD',
  `project_id` bigint unsigned DEFAULT NULL COMMENT '项目ID，NULL表示全局统计',
  `total_issues` bigint DEFAULT 0 COMMENT '问题总数',
  `new_issues` bigint DEFAULT 0 COMMENT '新增问题',
  `resolved_issues` bigint DEFAULT 0 COMMENT '已解决问题',
  `unresolved_issues` bigint DEFAULT 0 COMMENT '未闭环问题',
  `p0_p1_issues` bigint DEFAULT 0 COMMENT '高优问题P0+P1',
  `avg_resolve_minutes` bigint DEFAULT 0 COMMENT '平均处理时长(分钟)',
  `repeated_issues` bigint DEFAULT 0 COMMENT '重复问题数',
  `ai_analysis_count` bigint DEFAULT 0 COMMENT 'AI分析次数',
  `ai_adopted_count` bigint DEFAULT 0 COMMENT 'AI建议采纳数',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_stat_date` (`stat_date`),
  KEY `idx_project_id` (`project_id`),
  UNIQUE KEY `uk_date_project` (`stat_date`, `project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='每日统计快照表';