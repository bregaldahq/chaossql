// Package serveradmin holds the operator commands shared by the `chaossql server`
// and `chaossql-server` entrypoints. They require direct database access and
// print newly issued credentials exactly once.
package serveradmin

import (
	"database/sql"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/bregaldahq/chaossql/internal/server"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

func withStore(dbPath string, operation func(*server.Store) error) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()
	store := server.NewStore(db)
	if err := store.AutoMigrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	return operation(store)
}

// NewCreateTokenCommand issues a token for an existing organization.
func NewCreateTokenCommand(dbPath *string) *cobra.Command {
	var orgID, name, role string
	cmd := &cobra.Command{
		Use:   "create-token",
		Short: "Generate a new CI API token for an existing organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withStore(*dbPath, func(store *server.Store) error {
				token, err := store.IssueToken(orgID, name, server.Role(role))
				if err != nil {
					return fmt.Errorf("failed to issue token: %w", err)
				}
				out := cmd.OutOrStdout()
				fmt.Fprintln(out, "=== ChaosSQL API Token Created ===")
				fmt.Fprintf(out, "Organization: %s\n", orgID)
				fmt.Fprintf(out, "Token Name:   %s\n", name)
				fmt.Fprintf(out, "Role:         %s\n", role)
				fmt.Fprintf(out, "API Token:    %s\n", token)
				fmt.Fprintln(out, "===================================")
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&orgID, "org", "org_default", "Existing organization ID")
	cmd.Flags().StringVar(&name, "name", "CI Token", "Token descriptive name")
	cmd.Flags().StringVar(&role, "role", string(server.RoleMember), "Organization role: owner, admin, or member")
	return cmd
}

// NewOrgCommand manages tenants on a shared installation.
func NewOrgCommand(dbPath *string) *cobra.Command {
	orgCmd := &cobra.Command{
		Use:   "org",
		Short: "Provision and manage organizations (tenants)",
	}

	var name, plan string
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create an organization and print its initial owner token once",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return withStore(*dbPath, func(store *server.Store) error {
				org, token, err := store.ProvisionOrganization(name, plan)
				if err != nil {
					return fmt.Errorf("failed to create organization: %w", err)
				}
				out := cmd.OutOrStdout()
				fmt.Fprintln(out, "=== ChaosSQL Organization Created ===")
				fmt.Fprintf(out, "Organization: %s\n", org.ID)
				fmt.Fprintf(out, "Name:         %s\n", org.Name)
				fmt.Fprintf(out, "Plan:         %s\n", org.Plan)
				fmt.Fprintf(out, "Owner Token:  %s\n", token)
				fmt.Fprintln(out, "======================================")
				fmt.Fprintln(out, "Deliver the owner token privately; it is not shown again.")
				return nil
			})
		},
	}
	createCmd.Flags().StringVar(&name, "name", "", "Organization display name (required)")
	createCmd.Flags().StringVar(&plan, "plan", server.PlanDeveloper.ID, "Plan: developer, team, pro, or enterprise")
	_ = createCmd.MarkFlagRequired("name")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List organizations with plan and usage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return withStore(*dbPath, func(store *server.Store) error {
				orgs, err := store.ListOrganizations()
				if err != nil {
					return fmt.Errorf("failed to list organizations: %w", err)
				}
				return writeOrganizations(cmd.OutOrStdout(), orgs)
			})
		},
	}

	setPlanCmd := &cobra.Command{
		Use:   "set-plan <org-id> <plan>",
		Short: "Change an organization's plan",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withStore(*dbPath, func(store *server.Store) error {
				if err := store.SetOrganizationPlan(args[0], args[1]); err != nil {
					return fmt.Errorf("failed to set plan: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Organization %s is now on plan %s\n", args[0], args[1])
				return nil
			})
		},
	}

	orgCmd.AddCommand(createCmd, listCmd, setPlanCmd)
	return orgCmd
}

func writeOrganizations(out io.Writer, orgs []server.OrganizationSummary) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tPLAN\tREPOSITORIES\tTOKENS\tCREATED")
	for _, org := range orgs {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\n", org.ID, org.Name, org.Plan, org.Repositories, org.Tokens, org.CreatedAt.UTC().Format(time.RFC3339))
	}
	return w.Flush()
}
