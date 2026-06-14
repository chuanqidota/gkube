/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"gkube/config"
	"gkube/pkg/database"
	"gkube/pkg/es"
	"gkube/pkg/logger"
	"gkube/router"
	"net/http"
	"os"
	"os/signal"
	"time"

	clusterService "gkube/app/cluster/service"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gkube",
	Short: "操作k8s相关资源",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.webcontainer-go.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	config.Init()   // 初始化配置文件
	logger.Init()   // 初始化日志
	database.Init() // 初始化数据库
	es.Init()       // 初始化es

}

func Run() {
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
			fmt.Println(err.Error())
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("Server exiting")

}
