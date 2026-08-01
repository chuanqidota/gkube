/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"gkube/config"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/es"
	"gkube/pkg/logger"
	"gkube/internal/router"
	"net/http"
	"os"
	"os/signal"
	"time"

	clusterService "gkube/internal/cluster/service"

	"github.com/spf13/cobra"
)

var configPath string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gkube",
	Short: "操作k8s相关资源",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// PersistentPreRunE 在所有子命令执行前运行,完成启动链:logger → config → keys → database。
	// migrate/seed 子命令也会触发,但不会初始化 ES(仅 server Run 需要)。
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logger.Init()   // 初始化日志(最先)
		config.Init(configPath)
		auth.InitKeys() // 校验密钥,缺失即 Fatal
		database.Init() // 初始化数据库,失败 Fatal
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		Run()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "配置文件路径(默认工作目录 config/config.yaml,可用 GKUBE_CONFIG 环境变量覆盖)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func Run() {
	// ES 仅 server 运行时初始化
	es.Init()

	addr := fmt.Sprintf("%s:%s", config.Conf.Server.Ip, config.Conf.Server.Port)
	server := &http.Server{
		Addr:           addr,
		Handler:        router.Engine(),
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start health checker for clusters
	healthChecker := clusterService.NewHealthChecker(30 * time.Second)
	healthChecker.Start()
	defer healthChecker.Stop()

	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Error(fmt.Sprintf("服务器启动失败:%s", err.Error()))
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logger.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("服务器关闭失败:%s", err.Error()))
	}
	logger.Info("Server exiting")
}
