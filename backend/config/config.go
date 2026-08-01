package config

import (
	"fmt"
	"gkube/pkg/logger"
	"os"
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
	Security Security `json:"security"`
}

// Security 安全相关配置：密钥、CORS 白名单、管理员白名单。
// 所有字段必须通过环境变量或配置文件注入，禁止硬编码。
type Security struct {
	// JWTSecret JWT 签名密钥，必填。
	JWTSecret string `json:"jwt_secret" mapstructure:"jwt_secret" comment:"JWT签名密钥"`
	// AESKey AES-256-GCM 加密密钥(32 字节)，用于加密集群 kubeconfig，必填。
	AESKey string `json:"aes_key" mapstructure:"aes_key" comment:"AES-256-GCM加密密钥(32字节)"`
	// CORSOrigins 允许的跨域来源白名单，空时允许所有来源，生产环境应显式配置白名单。
	CORSOrigins []string `json:"cors_origins" mapstructure:"cors_origins" comment:"CORS允许的来源白名单"`
	// AdminUsers 管理员用户名白名单，登录响应会标记 isAdmin，并作为 RequireAdmin 依据。
	AdminUsers []string `json:"admin_users" mapstructure:"admin_users" comment:"管理员用户名白名单"`
}

var Conf = new(Config)

// Init 加载配置文件。优先级:path 参数 > GKUBE_CONFIG 环境变量 > 工作目录 config/config.yaml > 包目录 config.yaml。
func Init(path string) {
	if path == "" {
		path = os.Getenv("GKUBE_CONFIG")
	}
	if path == "" {
		// 默认工作目录相对路径
		if _, err := os.Stat("config/config.yaml"); err == nil {
			path = "config/config.yaml"
		} else {
			// 兜底:config 包所在目录(config.yaml 与其同级)
			pc, _, _, _ := runtime.Caller(0)
			fn := runtime.FuncForPC(pc)
			filePath, _ := fn.FileLine(0)
			configDir := filepath.Dir(filePath)
			path = filepath.Join(configDir, "config.yaml")
		}
	}
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logger.Fatal(fmt.Sprintf("读取配置文件失败:%s", err.Error()))
	}
	// 解析配置文件
	if err := viper.Unmarshal(&Conf); err != nil {
		logger.Fatal(fmt.Sprintf("解析配置文件失败:%s", err.Error()))
	}
	logger.Info(fmt.Sprintf("解析配置文件成功:%s", path))
}
