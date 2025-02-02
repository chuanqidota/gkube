package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func Init() {
	// 创建一个 lumberjack.Logger 实例，用于日志切割
	logFile := &lumberjack.Logger{
		Filename:   "./log/rbac.log", // 日志文件名
		MaxSize:    10,               // 每个日志文件最大大小，单位为 MB
		MaxBackups: 5,                // 最多保留的旧日志文件数量
		MaxAge:     7,                // 最长保留的旧日志文件天数
		LocalTime:  true,             // 以本地时间为基准
		Compress:   false,            // 是否压缩旧日志文件
	}

	// 设置日志输出格式
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:   true, // 禁用颜色输出
		TimestampFormat: "2006-01-02 15:04:05",
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			// 获取函数名和文件位置及行号
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", f.File, f.Line)
		},
	})
	// 启用调用者信息
	logger.SetReportCaller(true)

	// 创建一个 logrus.Logger 实例，用于日志记录和输出
	// 设置日志级别
	logger.SetLevel(logrus.InfoLevel)

	// 将日志输出到文件和控制台
	logger.SetOutput(io.MultiWriter(logFile, os.Stdout))
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}
