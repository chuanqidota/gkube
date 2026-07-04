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
	Short: "Seed default permissions, roles, and admin user",
	Long: `Initialize the database with default seed data:
- Default permissions for all resources (cluster, user, role, pod, deployment)
- Default roles: super_admin (all permissions), viewer (read-only)
- Default admin user (admin/admin123) if no users exist`,
	Run: func(cmd *cobra.Command, args []string) {
		seedPermissions()
		seedRoles()
		seedAdmin()
		fmt.Println("Seed data initialized successfully")
	},
}

func init() {
	rootCmd.AddCommand(seedCmd)
}

// seedPermissions creates default permissions for all resources.
func seedPermissions() {
	resources := []string{"cluster", "user", "role", "pod", "deployment"}
	actions := []string{"create", "read", "update", "delete"}

	// Create CRUD permissions for each resource
	for _, resource := range resources {
		for _, action := range actions {
			perm := authmodel.Permission{
				Resource:    resource,
				Action:      action,
				Description: fmt.Sprintf("Permission to %s %s", action, resource),
			}
			if err := database.DB.Where(authmodel.Permission{Resource: resource, Action: action}).
				FirstOrCreate(&perm).Error; err != nil {
				fmt.Printf("Failed to create permission %s:%s: %v\n", resource, action, err)
			}
		}
	}

	// Create wildcard permission
	wildcard := authmodel.Permission{
		Resource:    "*",
		Action:      "*",
		Description: "Wildcard permission for all resources and actions",
	}
	if err := database.DB.Where(authmodel.Permission{Resource: "*", Action: "*"}).
		FirstOrCreate(&wildcard).Error; err != nil {
		fmt.Printf("Failed to create wildcard permission: %v\n", err)
	}
}

// seedRoles creates default roles with associated permissions.
func seedRoles() {
	// Get all permissions
	var allPermissions []authmodel.Permission
	if err := database.DB.Find(&allPermissions).Error; err != nil {
		fmt.Printf("Failed to query permissions: %v\n", err)
		return
	}

	// Create super_admin role with all permissions
	superAdminRole := authmodel.Role{
		Name:        "super_admin",
		Description: "Super administrator with full access to all resources",
		Permissions: allPermissions,
	}
	if err := database.DB.Where(authmodel.Role{Name: "super_admin"}).
		FirstOrCreate(&superAdminRole).Error; err != nil {
		fmt.Printf("Failed to create super_admin role: %v\n", err)
	}

	// Create viewer role with read permissions only
	var readPermissions []authmodel.Permission
	for _, perm := range allPermissions {
		if perm.Action == "read" {
			readPermissions = append(readPermissions, perm)
		}
	}

	viewerRole := authmodel.Role{
		Name:        "viewer",
		Description: "Viewer role with read-only access to resources",
		Permissions: readPermissions,
	}
	if err := database.DB.Where(authmodel.Role{Name: "viewer"}).
		FirstOrCreate(&viewerRole).Error; err != nil {
		fmt.Printf("Failed to create viewer role: %v\n", err)
	}
}

// seedAdmin creates a default admin user if no users exist.
func seedAdmin() {
	var count int64
	database.DB.Model(&authmodel.User{}).Count(&count)
	if count > 0 {
		fmt.Println("Users already exist, skipping admin creation")
		return
	}

	// Find super_admin role
	var superAdminRole authmodel.Role
	if err := database.DB.Where(authmodel.Role{Name: "super_admin"}).First(&superAdminRole).Error; err != nil {
		fmt.Printf("Failed to find super_admin role: %v\n", err)
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
		Roles:        []authmodel.Role{superAdminRole},
	}
	if err := database.DB.Create(&adminUser).Error; err != nil {
		fmt.Printf("Failed to create admin user: %v\n", err)
		return
	}

	fmt.Println("Admin user created successfully (username: admin, password: admin123)")
}
