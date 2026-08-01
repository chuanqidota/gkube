package database

import (
	"fmt"
	"gkube/config"
	"gkube/pkg/logger"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// sanitizeDSN 返回不带密码的 DSN,用于日志打印。
func sanitizeDSN() string {
	return fmt.Sprintf("%s:***@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Conf.Database.User,
		config.Conf.Database.Host,
		config.Conf.Database.Port,
		config.Conf.Database.Name,
	)
}

// Init 初始化数据库连接,失败直接 Fatal 退出,不留 nil DB。
func Init() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Conf.Database.User,
		config.Conf.Database.Password,
		config.Conf.Database.Host,
		config.Conf.Database.Port,
		config.Conf.Database.Name,
	)
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		logger.Fatal(fmt.Sprintf("连接数据库出错：%s", err.Error()))
	}

	// 配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		logger.Fatal(fmt.Sprintf("获取底层 sql.DB 失败：%s", err.Error()))
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logger.Info(fmt.Sprintf("连接数据库成功：%s", sanitizeDSN()))
}
