package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	authmodel "gkube/internal/auth/model"
	"gkube/pkg/auth"
	"gkube/pkg/database"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed default admin user",
	Long: `Initialize the database with default seed data:
- Default admin user (admin/admin123) if no users exist`,
	Run: func(cmd *cobra.Command, args []string) {
		seedAdmin()
		fmt.Println("Seed data initialized successfully")
	},
}

func init() {
	rootCmd.AddCommand(seedCmd)
}

// seedAdmin creates a default admin user if no users exist.
func seedAdmin() {
	var count int64
	database.DB.Model(&authmodel.User{}).Count(&count)
	if count > 0 {
		fmt.Println("Users already exist, skipping admin creation")
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword("admin123")
	if err != nil {
		fmt.Printf("Failed to hash password: %v\n", err)
		return
	}

	// Create admin user
	adminUser := authmodel.User{
		Username:     "admin",
		PasswordHash: hashedPassword,
		Email:        "admin@gkube.local",
		DisplayName:  "System Administrator",
		Status:       1,
	}
	if err := database.DB.Create(&adminUser).Error; err != nil {
		fmt.Printf("Failed to create admin user: %v\n", err)
		return
	}

	fmt.Println("Admin user created successfully (username: admin, password: admin123)")
}
