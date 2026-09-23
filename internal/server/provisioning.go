package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidPlan = errors.New("invalid plan")

const maxOrganizationNameLength = 120

// OrganizationSummary adds usage counters for operator listings.
type OrganizationSummary struct {
	Organization
	Repositories int
	Tokens       int
}

func validPlanID(plan string) bool {
	switch plan {
	case PlanDeveloper.ID, PlanTeam.ID, PlanPro.ID, PlanEnterprise.ID:
		return true
	}
	return false
}

// NewAPIToken returns a raw bearer credential. Only its hash is persisted.
func NewAPIToken() (string, error) {
	var b [24]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "csql_" + hex.EncodeToString(b[:]), nil
}

// ProvisionOrganization creates a tenant and its first owner token in one
// transaction, so a tenant never exists without a credential to manage it.
func (s *Store) ProvisionOrganization(name, plan string) (Organization, string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > maxOrganizationNameLength {
		return Organization{}, "", fmt.Errorf("organization name must be 1-%d characters", maxOrganizationNameLength)
	}
	if !validPlanID(plan) {
		return Organization{}, "", fmt.Errorf("%w: %q", ErrInvalidPlan, plan)
	}
	orgID, err := randomIdentifier("org_")
	if err != nil {
		return Organization{}, "", err
	}
	tokenID, err := randomIdentifier("tok_")
	if err != nil {
		return Organization{}, "", err
	}
	token, err := NewAPIToken()
	if err != nil {
		return Organization{}, "", err
	}
	org := Organization{ID: orgID, Name: name, Plan: plan, CreatedAt: time.Now().UTC()}
	err = s.withTransaction(context.Background(), func(bound *Store) error {
		if _, err := bound.tx.Exec(`INSERT INTO organizations (id, name, plan, created_at) VALUES (?, ?, ?, ?)`,
			org.ID, org.Name, org.Plan, org.CreatedAt); err != nil {
			return fmt.Errorf("create organization: %w", err)
		}
		if err := bound.CreateAPITokenWithRole(tokenID, org.ID, token, "Initial Owner Token", RoleOwner); err != nil {
			return fmt.Errorf("create owner token: %w", err)
		}
		return nil
	})
	if err != nil {
		return Organization{}, "", err
	}
	return org, token, nil
}

// IssueToken never creates organizations: a mistyped ID must fail instead of
// silently provisioning a new tenant.
func (s *Store) IssueToken(orgID, name string, role Role) (string, error) {
	if !role.valid() {
		return "", fmt.Errorf("%w: %q", ErrInvalidRole, role)
	}
	tokenID, err := randomIdentifier("tok_")
	if err != nil {
		return "", err
	}
	token, err := NewAPIToken()
	if err != nil {
		return "", err
	}
	err = s.withTransaction(context.Background(), func(bound *Store) error {
		var exists int
		if err := bound.tx.QueryRow(`SELECT COUNT(*) FROM organizations WHERE id = ?`, orgID).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			return fmt.Errorf("organization %q: %w", orgID, ErrNotFound)
		}
		return bound.CreateAPITokenWithRole(tokenID, orgID, token, name, role)
	})
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *Store) SetOrganizationPlan(orgID, plan string) error {
	if !validPlanID(plan) {
		return fmt.Errorf("%w: %q", ErrInvalidPlan, plan)
	}
	result, err := s.queryer().Exec(`UPDATE organizations SET plan = ? WHERE id = ?`, plan, orgID)
	if err != nil {
		return err
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if updated == 0 {
		return fmt.Errorf("organization %q: %w", orgID, ErrNotFound)
	}
	return nil
}

func (s *Store) ListOrganizations() ([]OrganizationSummary, error) {
	rows, err := s.queryer().Query(`SELECT o.id, o.name, o.plan, o.created_at,
		(SELECT COUNT(*) FROM repositories r WHERE r.org_id = o.id),
		(SELECT COUNT(*) FROM api_tokens t WHERE t.org_id = o.id)
		FROM organizations o ORDER BY o.created_at, o.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orgs []OrganizationSummary
	for rows.Next() {
		var org OrganizationSummary
		if err := rows.Scan(&org.ID, &org.Name, &org.Plan, &org.CreatedAt, &org.Repositories, &org.Tokens); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, rows.Err()
}
