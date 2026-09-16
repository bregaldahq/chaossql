package server

import "context"

type principalContextKey struct{}

func contextWithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

func principalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	return principal, ok && principal.OrgID != "" && principal.Role.valid()
}

func organizationFromContext(ctx context.Context) string {
	principal, ok := principalFromContext(ctx)
	if !ok {
		return ""
	}
	return principal.OrgID
}
