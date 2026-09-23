package server

import (
	"fmt"
	"strings"
)

// ValidateBootstrapToken prevents an unattended installation from using the
// credentials published in older deployment examples.
func ValidateBootstrapToken(token string) error {
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("CHAOSSQL_ADMIN_TOKEN or --token is required")
	}
	switch strings.TrimSpace(token) {
	case "chaossql_dev_token", "chaossql_prod_secret", "chaossql_prod_secret_token_change_me", "chaossql_enterprise_token_secret":
		return fmt.Errorf("replace the public example token with a unique CHAOSSQL_ADMIN_TOKEN")
	}
	if token != strings.TrimSpace(token) {
		return fmt.Errorf("CHAOSSQL_ADMIN_TOKEN must not contain surrounding whitespace")
	}
	return nil
}

// ConfigureBootstrapOwner is shared by both server entrypoints. Upserting the
// same owner token ID rotates the credential instead of retaining an old token.
func ConfigureBootstrapOwner(store *Store, token string) error {
	if err := ValidateBootstrapToken(token); err != nil {
		return err
	}
	if err := store.EnsureOrganization("org_default", "Default Organization", "pro"); err != nil {
		return fmt.Errorf("failed to configure bootstrap organization: %w", err)
	}
	if err := store.UpsertAPITokenWithRole("tok_admin", "org_default", token, "Initial Admin Token", RoleOwner); err != nil {
		return fmt.Errorf("failed to configure bootstrap owner token: %w", err)
	}
	return nil
}
