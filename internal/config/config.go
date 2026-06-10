package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Asynq    AsynqConfig    `mapstructure:"asynq"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
	AI       AIConfig       `mapstructure:"ai"`
}

type ServerConfig struct {
	Port     int    `mapstructure:"port"`
	GRPCPort int    `mapstructure:"grpc_port"` // gRPC 端口（0=不启用）
	Mode     string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AsynqConfig struct {
	RedisAddr   string `mapstructure:"redis_addr"`
	Concurrency int    `mapstructure:"concurrency"`
	RetryLimit  int    `mapstructure:"retry_limit"`
	Timeout     int    `mapstructure:"timeout"` // 秒
	Enabled     bool   `mapstructure:"enabled"` // 功能开关
}

type RabbitMQConfig struct {
	URL           string `mapstructure:"url"`
	Exchange      string `mapstructure:"exchange"`
	CommandQueue  string `mapstructure:"command_queue"`
	ResultQueue   string `mapstructure:"result_queue"`
	ProgressQueue string `mapstructure:"progress_queue"`
	Enabled       bool   `mapstructure:"enabled"`        // MQ 功能开关
	GrayPercent   int    `mapstructure:"gray_percent"`   // 灰度百分比 (0-100)
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireTime int    `mapstructure:"expire_time"` // hours
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`    // MB
	MaxBackups int    `mapstructure:"max_backups"` // number of backups
	MaxAge     int    `mapstructure:"max_age"`     // days
	Compress   bool   `mapstructure:"compress"`
}

type AIConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	BaseURL       string `mapstructure:"base_url"`        // Python Agent HTTP（兜底 / 健康检查）
	GRPCAddr      string `mapstructure:"grpc_addr"`       // Python Agent gRPC 地址（worker 直连）
	Timeout       int    `mapstructure:"timeout"`         // HTTP 超时（秒）
	GRPCTimeout   int    `mapstructure:"grpc_timeout"`    // gRPC 默认超时（秒）
}

var GlobalConfig *Config

func Init(configPath string) error {
	// 如果是相对路径，转换为绝对路径
	if !filepath.IsAbs(configPath) {
		absPath, err := filepath.Abs(configPath)
		if err == nil {
			configPath = absPath
		}
	}

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 尝试从工作目录向上查找
		cwd, _ := os.Getwd()
		tryPaths := []string{
			filepath.Join(cwd, configPath),
			filepath.Join(cwd, "configs", "config.yaml"),
			filepath.Join(cwd, "..", "configs", "config.yaml"),
		}
		
		for _, tryPath := range tryPaths {
			if _, err := os.Stat(tryPath); err == nil {
				configPath = tryPath
				break
			}
		}
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	GlobalConfig = &Config{}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

func GetConfig() *Config {
	return GlobalConfig
}
