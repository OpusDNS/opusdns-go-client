package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users",
	Long:  `Manage users and their role assignments.`,
}

var usersRoleCmd = &cobra.Command{
	Use:   "role",
	Short: "Manage a user's role assignment",
}

var usersRoleGetCmd = &cobra.Command{
	Use:   "get <user-id>",
	Short: "Get the role assigned to a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		assignment, err := getClient().Users.GetUserRole(ctx, models.UserID(args[0]))
		if err != nil {
			return fmt.Errorf("failed to get user role: %w", err)
		}

		data, err := json.MarshalIndent(assignment, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format role assignment: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

var usersRoleSetCmd = &cobra.Command{
	Use:   "set <user-id> <role>",
	Short: "Set the role for a user",
	Long:  `Set the role for a user, replacing any existing role. Pass --clear to remove the role.`,
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		clear, _ := cmd.Flags().GetBool("clear")

		var role *string
		switch {
		case clear:
			role = nil
		case len(args) == 2:
			role = models.StringPtr(args[1])
		default:
			return fmt.Errorf("a role argument is required unless --clear is set")
		}

		assignment, err := getClient().Users.SetUserRole(ctx, models.UserID(args[0]), role)
		if err != nil {
			return fmt.Errorf("failed to set user role: %w", err)
		}

		fmt.Printf("✓ User role updated successfully!\n\n")

		data, err := json.MarshalIndent(assignment, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format role assignment: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(usersCmd)

	usersCmd.AddCommand(usersRoleCmd)
	usersRoleCmd.AddCommand(usersRoleGetCmd)

	usersRoleCmd.AddCommand(usersRoleSetCmd)
	usersRoleSetCmd.Flags().Bool("clear", false, "Clear the user's role instead of setting one")
}
