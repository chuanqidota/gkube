package config

import (
	"fmt"
	"github.com/spf13/viper"
	"gkube/pkg/logger"
	"path/filepath"
	"runtime"
)

type Config struct {
	Server struct {
		Ip   string `json:"ip"`
		Port string `json:"port"`
	}
	Database struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Name     string `json:"name"`
	}
	Redis struct {
		Addr     string `json:"addr"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	}
	ElasticSearch struct {
		Url      string `json:"url" comment:"es地址"`
		Username string `json:"username" comment:"es用户名"`
		Password string `json:"password" comment:"es密码"`
	}
	Audit struct {
		RecordAuditIndex string `json:"record_audit" comment:"操作审计-es索引"`
	}
	S3 struct {
		EndPoint        string `json:"endpoint" comment:"地址"`
		AccessKeyID     string `json:"accessKeyID" comment:"密钥ID"`
		SecretAccessKey string `json:"secretAccessKey" comment:"密钥KEY"`
		UseSSL          bool   `json:"useSSL" comment:"是否使用SSL"`
		Bucket          string `json:"bucket" comment:"桶名字"`
	}
}

var Conf = new(Config)

func Init() {
	viper.SetConfigFile("./config/config.yaml")

	pc, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(pc)
	filePath, _ := fn.FileLine(0)
	dirPath := filepath.Dir(filepath.Dir(filePath))
	absolutePath := filepath.Join(dirPath, "config.yaml")
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
