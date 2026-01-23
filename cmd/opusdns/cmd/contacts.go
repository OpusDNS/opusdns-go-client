package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/spf13/cobra"
)

var contactsCmd = &cobra.Command{
	Use:   "contacts",
	Short: "Manage contacts",
	Long:  `List, create, update, delete, and verify contacts for domain registrations.`,
}

var contactsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all contacts",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		search, _ := cmd.Flags().GetString("search")
		email, _ := cmd.Flags().GetString("email")
		country, _ := cmd.Flags().GetString("country")
		verifiedFlag, _ := cmd.Flags().GetBool("verified")
		verifiedChanged := cmd.Flags().Changed("verified")

		opts := &models.ListContactsOptions{}
		if search != "" {
			opts.Search = search
		}
		if email != "" {
			opts.Email = email
		}
		if country != "" {
			opts.Country = country
		}
		if verifiedChanged {
			opts.Verified = &verifiedFlag
		}

		contacts, err := getClient().Contacts.ListContacts(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to list contacts: %w", err)
		}

		if len(contacts) == 0 {
			fmt.Println("No contacts found.")
			return nil
		}

		fmt.Printf("Found %d contact(s):\n\n", len(contacts))
		for _, contact := range contacts {
			verified := "✗"
			if contact.Verified {
				verified = "✓"
			}
			org := ""
			if contact.Org != nil && *contact.Org != "" {
				org = fmt.Sprintf(" (%s)", *contact.Org)
			}
			fmt.Printf("  • %s: %s%s <%s> [verified: %s]\n",
				contact.ContactID,
				contact.FullName(),
				org,
				contact.Email,
				verified,
			)
		}

		return nil
	},
}

var contactsGetCmd = &cobra.Command{
	Use:   "get <contact-id>",
	Short: "Get details of a contact",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		contactID := models.ContactID(args[0])

		contact, err := getClient().Contacts.GetContact(ctx, contactID)
		if err != nil {
			return fmt.Errorf("failed to get contact: %w", err)
		}

		data, err := json.MarshalIndent(contact, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format contact: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

var contactsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new contact",
	Long: `Create a new contact for domain registrations.

Required flags: --first-name, --last-name, --email, --phone, --street, --city, --postal-code, --country

Examples:
  opusdns contacts create --first-name John --last-name Doe --email john@example.com \
    --phone "+1.2125551234" --street "123 Main St" --city "New York" \
    --postal-code "10001" --country US`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		firstName, _ := cmd.Flags().GetString("first-name")
		lastName, _ := cmd.Flags().GetString("last-name")
		email, _ := cmd.Flags().GetString("email")
		phone, _ := cmd.Flags().GetString("phone")
		street, _ := cmd.Flags().GetString("street")
		city, _ := cmd.Flags().GetString("city")
		postalCode, _ := cmd.Flags().GetString("postal-code")
		country, _ := cmd.Flags().GetString("country")
		disclose, _ := cmd.Flags().GetBool("disclose")

		req := &models.ContactCreateRequest{
			FirstName:  firstName,
			LastName:   lastName,
			Email:      email,
			Phone:      phone,
			Street:     street,
			City:       city,
			PostalCode: postalCode,
			Country:    country,
			Disclose:   disclose,
		}

		// Optional fields
		if cmd.Flags().Changed("org") {
			org, _ := cmd.Flags().GetString("org")
			req.Org = &org
		}
		if cmd.Flags().Changed("title") {
			title, _ := cmd.Flags().GetString("title")
			req.Title = &title
		}
		if cmd.Flags().Changed("fax") {
			fax, _ := cmd.Flags().GetString("fax")
			req.Fax = &fax
		}
		if cmd.Flags().Changed("state") {
			state, _ := cmd.Flags().GetString("state")
			req.State = &state
		}

		contact, err := getClient().Contacts.CreateContact(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to create contact: %w", err)
		}

		fmt.Printf("✓ Contact '%s' created successfully!\n\n", contact.ContactID)

		data, err := json.MarshalIndent(contact, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format contact: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

var contactsUpdateCmd = &cobra.Command{
	Use:   "update <contact-id>",
	Short: "Update an existing contact",
	Long: `Update an existing contact's information.

Examples:
  opusdns contacts update ct_abc123 --email newemail@example.com
  opusdns contacts update ct_abc123 --phone "+1.2125559999" --city "Los Angeles"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		contactID := models.ContactID(args[0])

		req := &models.ContactUpdateRequest{}
		hasChanges := false

		if cmd.Flags().Changed("first-name") {
			firstName, _ := cmd.Flags().GetString("first-name")
			req.FirstName = &firstName
			hasChanges = true
		}
		if cmd.Flags().Changed("last-name") {
			lastName, _ := cmd.Flags().GetString("last-name")
			req.LastName = &lastName
			hasChanges = true
		}
		if cmd.Flags().Changed("org") {
			org, _ := cmd.Flags().GetString("org")
			req.Org = &org
			hasChanges = true
		}
		if cmd.Flags().Changed("title") {
			title, _ := cmd.Flags().GetString("title")
			req.Title = &title
			hasChanges = true
		}
		if cmd.Flags().Changed("email") {
			email, _ := cmd.Flags().GetString("email")
			req.Email = &email
			hasChanges = true
		}
		if cmd.Flags().Changed("phone") {
			phone, _ := cmd.Flags().GetString("phone")
			req.Phone = &phone
			hasChanges = true
		}
		if cmd.Flags().Changed("fax") {
			fax, _ := cmd.Flags().GetString("fax")
			req.Fax = &fax
			hasChanges = true
		}
		if cmd.Flags().Changed("street") {
			street, _ := cmd.Flags().GetString("street")
			req.Street = &street
			hasChanges = true
		}
		if cmd.Flags().Changed("city") {
			city, _ := cmd.Flags().GetString("city")
			req.City = &city
			hasChanges = true
		}
		if cmd.Flags().Changed("state") {
			state, _ := cmd.Flags().GetString("state")
			req.State = &state
			hasChanges = true
		}
		if cmd.Flags().Changed("postal-code") {
			postalCode, _ := cmd.Flags().GetString("postal-code")
			req.PostalCode = &postalCode
			hasChanges = true
		}
		if cmd.Flags().Changed("country") {
			country, _ := cmd.Flags().GetString("country")
			req.Country = &country
			hasChanges = true
		}
		if cmd.Flags().Changed("disclose") {
			disclose, _ := cmd.Flags().GetBool("disclose")
			req.Disclose = &disclose
			hasChanges = true
		}

		if !hasChanges {
			return fmt.Errorf("no changes specified, use flags like --email, --phone, etc")
		}

		contact, err := getClient().Contacts.UpdateContact(ctx, contactID, req)
		if err != nil {
			return fmt.Errorf("failed to update contact: %w", err)
		}

		fmt.Printf("✓ Contact '%s' updated successfully!\n\n", contact.ContactID)

		data, err := json.MarshalIndent(contact, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format contact: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

var contactsDeleteCmd = &cobra.Command{
	Use:   "delete <contact-id>",
	Short: "Delete a contact",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		contactID := models.ContactID(args[0])

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Are you sure you want to delete contact '%s'? This action cannot be undone.\n", contactID)
			fmt.Print("Type 'yes' to confirm: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		err := getClient().Contacts.DeleteContact(ctx, contactID)
		if err != nil {
			return fmt.Errorf("failed to delete contact: %w", err)
		}

		fmt.Printf("✓ Contact '%s' deleted successfully!\n", contactID)
		return nil
	},
}

var contactsVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Manage contact verification",
	Long:  `Request verification, check verification status, or verify a contact with a token.`,
}

var contactsVerifyRequestCmd = &cobra.Command{
	Use:   "request <contact-id>",
	Short: "Request email verification for a contact",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		contactID := models.ContactID(args[0])

		verification, err := getClient().Contacts.RequestVerification(ctx, contactID)
		if err != nil {
			return fmt.Errorf("failed to request verification: %w", err)
		}

		fmt.Printf("✓ Verification requested for contact '%s'\n\n", contactID)
		fmt.Printf("Status: %s\n", verification.Status)
		if verification.ExpiresOn != nil {
			fmt.Printf("Expires: %s\n", verification.ExpiresOn.Format("2006-01-02 15:04:05"))
		}

		return nil
	},
}

var contactsVerifyStatusCmd = &cobra.Command{
	Use:   "status <contact-id>",
	Short: "Get verification status for a contact",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		contactID := models.ContactID(args[0])

		verification, err := getClient().Contacts.GetVerificationStatus(ctx, contactID)
		if err != nil {
			return fmt.Errorf("failed to get verification status: %w", err)
		}

		data, err := json.MarshalIndent(verification, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format verification: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

var contactsVerifyTokenCmd = &cobra.Command{
	Use:   "token <verification-token>",
	Short: "Verify a contact using a verification token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		token := args[0]

		err := getClient().Contacts.VerifyContact(ctx, &models.ContactVerificationRequest{
			Token: token,
		})
		if err != nil {
			return fmt.Errorf("failed to verify contact: %w", err)
		}

		fmt.Println("✓ Contact verified successfully!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(contactsCmd)

	// List subcommand
	contactsCmd.AddCommand(contactsListCmd)
	contactsListCmd.Flags().String("search", "", "Search contacts by name or email")
	contactsListCmd.Flags().String("email", "", "Filter by email address")
	contactsListCmd.Flags().String("country", "", "Filter by country code (e.g., US, DE)")
	contactsListCmd.Flags().Bool("verified", false, "Filter by verification status")

	// Get subcommand
	contactsCmd.AddCommand(contactsGetCmd)

	// Create subcommand
	contactsCmd.AddCommand(contactsCreateCmd)
	contactsCreateCmd.Flags().String("first-name", "", "Contact's first name (required)")
	contactsCreateCmd.Flags().String("last-name", "", "Contact's last name (required)")
	contactsCreateCmd.Flags().String("org", "", "Organization name")
	contactsCreateCmd.Flags().String("title", "", "Title (e.g., Mr., Dr.)")
	contactsCreateCmd.Flags().String("email", "", "Email address (required)")
	contactsCreateCmd.Flags().String("phone", "", "Phone number in E.164 format (required)")
	contactsCreateCmd.Flags().String("fax", "", "Fax number")
	contactsCreateCmd.Flags().String("street", "", "Street address (required)")
	contactsCreateCmd.Flags().String("city", "", "City (required)")
	contactsCreateCmd.Flags().String("state", "", "State or province")
	contactsCreateCmd.Flags().String("postal-code", "", "Postal/ZIP code (required)")
	contactsCreateCmd.Flags().String("country", "", "Two-letter country code (required)")
	contactsCreateCmd.Flags().Bool("disclose", false, "Publicly disclose contact information")
	_ = contactsCreateCmd.MarkFlagRequired("first-name")
	_ = contactsCreateCmd.MarkFlagRequired("last-name")
	_ = contactsCreateCmd.MarkFlagRequired("email")
	_ = contactsCreateCmd.MarkFlagRequired("phone")
	_ = contactsCreateCmd.MarkFlagRequired("street")
	_ = contactsCreateCmd.MarkFlagRequired("city")
	_ = contactsCreateCmd.MarkFlagRequired("postal-code")
	_ = contactsCreateCmd.MarkFlagRequired("country")

	// Update subcommand
	contactsCmd.AddCommand(contactsUpdateCmd)
	contactsUpdateCmd.Flags().String("first-name", "", "Contact's first name")
	contactsUpdateCmd.Flags().String("last-name", "", "Contact's last name")
	contactsUpdateCmd.Flags().String("org", "", "Organization name")
	contactsUpdateCmd.Flags().String("title", "", "Title (e.g., Mr., Dr.)")
	contactsUpdateCmd.Flags().String("email", "", "Email address")
	contactsUpdateCmd.Flags().String("phone", "", "Phone number in E.164 format")
	contactsUpdateCmd.Flags().String("fax", "", "Fax number")
	contactsUpdateCmd.Flags().String("street", "", "Street address")
	contactsUpdateCmd.Flags().String("city", "", "City")
	contactsUpdateCmd.Flags().String("state", "", "State or province")
	contactsUpdateCmd.Flags().String("postal-code", "", "Postal/ZIP code")
	contactsUpdateCmd.Flags().String("country", "", "Two-letter country code")
	contactsUpdateCmd.Flags().Bool("disclose", false, "Publicly disclose contact information")

	// Delete subcommand
	contactsCmd.AddCommand(contactsDeleteCmd)
	contactsDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	// Verify subcommands
	contactsCmd.AddCommand(contactsVerifyCmd)
	contactsVerifyCmd.AddCommand(contactsVerifyRequestCmd)
	contactsVerifyCmd.AddCommand(contactsVerifyStatusCmd)
	contactsVerifyCmd.AddCommand(contactsVerifyTokenCmd)
}
