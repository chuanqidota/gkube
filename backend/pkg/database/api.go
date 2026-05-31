package database

import (
	"fmt"
	"gkube/config"
	"gkube/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Init
//
//	@Description: 初始化数据库
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
		logger.Error(fmt.Sprintf("连接数据库出错：%s", err.Error()))
		return
	}
	logger.Info(fmt.Sprintf("连接数据库成功：%s", dsn))
}
