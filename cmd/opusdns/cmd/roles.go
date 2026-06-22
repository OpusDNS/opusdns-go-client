package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/spf13/cobra"
)

var rolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "Manage organization roles",
	Long:  `List, get, create, update, and delete organization roles, and list grantable permissions.`,
}

var rolesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all roles (built-in and custom)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		roles, err := getClient().Organizations.ListRoles(ctx)
		if err != nil {
			return fmt.Errorf("failed to list roles: %w", err)
		}

		if len(roles) == 0 {
			fmt.Println("No roles found.")
			return nil
		}

		fmt.Printf("Found %d role(s):\n\n", len(roles))
		for _, role := range roles {
			kind := "custom"
			if role.BuiltIn {
				kind = "built-in"
			}
			fmt.Printf("  • %s (%s) — %s\n", role.Label, kind, role.Name)
		}

		return nil
	},
}

var rolesGetCmd = &cobra.Command{
	Use:   "get <label>",
	Short: "Get details of a role",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		role, err := getClient().Organizations.GetRole(ctx, args[0])
		if err != nil {
			return fmt.Errorf("failed to get role: %w", err)
		}

		data, err := json.MarshalIndent(role, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format role: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

var rolesCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a custom role",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		permissions, _ := cmd.Flags().GetStringArray("permission")
		description, _ := cmd.Flags().GetString("description")

		req := &models.CustomRoleCreateRequest{
			Name:        args[0],
			Permissions: permissions,
		}
		if description != "" {
			req.Description = models.StringPtr(description)
		}

		role, err := getClient().Organizations.CreateRole(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to create role: %w", err)
		}

		fmt.Printf("✓ Role '%s' created successfully!\n\n", role.Label)

		data, err := json.MarshalIndent(role, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format role: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

var rolesUpdateCmd = &cobra.Command{
	Use:   "update <label>",
	Short: "Update a custom role",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		req := &models.CustomRoleUpdateRequest{}
		if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			req.Name = models.StringPtr(name)
		}
		if cmd.Flags().Changed("description") {
			description, _ := cmd.Flags().GetString("description")
			req.Description = models.StringPtr(description)
		}
		if cmd.Flags().Changed("permission") {
			permissions, _ := cmd.Flags().GetStringArray("permission")
			req.Permissions = &permissions
		}

		role, err := getClient().Organizations.UpdateRole(ctx, args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update role: %w", err)
		}

		fmt.Printf("✓ Role '%s' updated successfully!\n\n", role.Label)

		data, err := json.MarshalIndent(role, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format role: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

var rolesDeleteCmd = &cobra.Command{
	Use:   "delete <label>",
	Short: "Delete a custom role",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		label := args[0]

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Are you sure you want to delete role '%s'? This action cannot be undone.\n", label)
			fmt.Print("Type 'yes' to confirm: ")
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		if err := getClient().Organizations.DeleteRole(ctx, label); err != nil {
			return fmt.Errorf("failed to delete role: %w", err)
		}

		fmt.Printf("✓ Role '%s' deleted successfully!\n", label)
		return nil
	},
}

var rolesPermissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "List the permissions a custom role may grant",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		catalog, err := getClient().Organizations.ListRolePermissions(ctx)
		if err != nil {
			return fmt.Errorf("failed to list role permissions: %w", err)
		}

		fmt.Printf("Found %d grantable permission(s):\n\n", len(catalog.Permissions))
		for _, permission := range catalog.Permissions {
			fmt.Printf("  • %s\n", permission)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(rolesCmd)

	rolesCmd.AddCommand(rolesListCmd)
	rolesCmd.AddCommand(rolesGetCmd)

	rolesCmd.AddCommand(rolesCreateCmd)
	rolesCreateCmd.Flags().StringArray("permission", nil, "Permission to grant in 'resource:scope' form (repeatable)")
	rolesCreateCmd.Flags().String("description", "", "Description of the role")

	rolesCmd.AddCommand(rolesUpdateCmd)
	rolesUpdateCmd.Flags().String("name", "", "New display name")
	rolesUpdateCmd.Flags().String("description", "", "New description")
	rolesUpdateCmd.Flags().StringArray("permission", nil, "Full replacement set of 'resource:scope' permissions (repeatable)")

	rolesCmd.AddCommand(rolesDeleteCmd)
	rolesDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	rolesCmd.AddCommand(rolesPermissionsCmd)
}
