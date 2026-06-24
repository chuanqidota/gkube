package config

import (
	"fmt"
	"gkube/pkg/logger"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Ip   string `json:"ip"`
		Port string `json:"port"`
	} `json:"server"`
	Database struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Name     string `json:"name"`
	} `json:"database"`
	ElasticSearch struct {
		Enable   bool   `json:"enable" mapstructure:"enable" comment:"是否启用es"`
		Url      string `json:"url" mapstructure:"url" comment:"es地址"`
		Username string `json:"username" mapstructure:"username" comment:"es用户名"`
		Password string `json:"password" mapstructure:"password" comment:"es密码"`
	} `json:"elasticSearch"`
	Audit struct {
		RecordAuditIndex string `json:"record_audit" mapstructure:"record_audit" comment:"操作审计-es索引"`
	} `json:"audit"`
	S3 struct {
		EndPoint        string `json:"endpoint" comment:"地址"`
		AccessKeyID     string `json:"accessKeyID" comment:"密钥ID"`
		SecretAccessKey string `json:"secretAccessKey" comment:"密钥KEY"`
		UseSSL          bool   `json:"useSSL" comment:"是否使用SSL"`
		Bucket          string `json:"bucket" comment:"桶名字"`
	} `json:"s3"`
}

var Conf = new(Config)

func Init() {
	// 获取 config 包所在目录，config.yaml 与其同级
	pc, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(pc)
	filePath, _ := fn.FileLine(0)
	configDir := filepath.Dir(filePath)
	absolutePath := filepath.Join(configDir, "config.yaml")
	viper.SetConfigFile(absolutePath)

	if err := viper.ReadInConfig(); err != nil {
		logger.Error(fmt.Sprintf("读取配置文件失败:%s", err.Error()))
	}
	// 解析配置文件
	if err := viper.Unmarshal(&Conf); err != nil {
		logger.Error(fmt.Sprintf("解析配置文件失败:%s", err.Error()))
	}
	logger.Info(fmt.Sprintf("解析配置文件：%v", *Conf))
}
