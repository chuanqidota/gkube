package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	authmodel "gkube/internal/auth/model"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/logger"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed default admin user",
	Long: `Initialize the database with default seed data:
- Default admin user with a randomly generated password if no users exist`,
	Run: func(cmd *cobra.Command, args []string) {
		seedAdmin()
		logger.Info("Seed data initialized successfully")
	},
}

func init() {
	rootCmd.AddCommand(seedCmd)
}

// seedAdmin creates a default admin user with a random password if no users exist.
func seedAdmin() {
	var count int64
	database.DB.Model(&authmodel.User{}).Count(&count)
	if count > 0 {
		logger.Info("Users already exist, skipping admin creation")
		return
	}

	// 随机生成 16 字节口令并打印一次,不再使用固定弱口令
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		logger.Fatal(fmt.Sprintf("生成随机口令失败:%s", err.Error()))
	}
	password := hex.EncodeToString(b)

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to hash password: %v", err))
	}

	adminUser := authmodel.User{
		Username:     "admin",
		PasswordHash: hashedPassword,
		Email:        "admin@gkube.local",
		DisplayName:  "System Administrator",
		Status:       1,
	}
	if err := database.DB.Create(&adminUser).Error; err != nil {
		logger.Fatal(fmt.Sprintf("Failed to create admin user: %v", err))
	}

	// 仅打印一次随机口令,请妥善保存
	logger.Info(fmt.Sprintf("Admin user created successfully (username: admin, password: %s)", password))
}
