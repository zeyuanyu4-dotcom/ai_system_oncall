package database

import (
	"fmt"
	"time"

	"ai_system_oncall/internal/config"
	"ai_system_oncall/internal/model"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.DatabaseConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Set connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto migrate
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	zap.L().Info("Database connected successfully")
	return nil
}

func autoMigrate() error {
	return DB.AutoMigrate(
		// 用户相关
		&model.User{},

		// 项目相关
		&model.Project{},
		&model.ProjectMember{},

		// 服务相关
		&model.Service{},
		&model.ServiceAPI{},
		&model.ServiceDependency{},

		// 问题相关
		&model.Issue{},
		&model.IssueComment{},
		&model.IssueStatusLog{},
		&model.IssueOperationLog{},

		// 日志相关
		&model.SimulatedLog{},

		// 知识库相关
		&model.KnowledgeDocument{},
		&model.KnowledgeDocVersion{},
		&model.KnowledgeDocAttachment{},

		// AI 分析相关
		&model.AIAnalysisTask{},

		// 报告相关
		&model.Report{},

		// 统计相关
		&model.StatDailyRecord{},
	)
}

func GetDB() *gorm.DB {
	return DB
}

func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
