package middlewares

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	mycontext "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/context"
)

type fakeRoleChecker struct {
	calls      int
	lastUserID int
	lastRoles  []string
	err        error
}

func (f *fakeRoleChecker) CheckRole(_ context.Context, userID int, requiredRoles ...string) error {
	f.calls++
	f.lastUserID = userID
	f.lastRoles = requiredRoles
	return f.err
}

// passthroughResolver records whether the wrapped field resolver ran.
func passthroughResolver(called *bool) graphql.Resolver {
	return func(ctx context.Context) (any, error) {
		*called = true
		return "resolved", nil
	}
}

func TestHasRole_SkipsWhenNoUserInContext(t *testing.T) {
	checker := &fakeRoleChecker{}
	called := false

	res, err := HasRole(checker)(context.Background(), nil, passthroughResolver(&called), []string{"Admin"})
	if err != nil {
		t.Fatalf("expected public field to pass through, got error %v", err)
	}
	if !called || res != "resolved" {
		t.Fatal("expected the field resolver to run")
	}
	if checker.calls != 0 {
		t.Fatalf("expected checker not to run, ran %d times", checker.calls)
	}
}

func TestHasRole_AllowsWhenCheckerPermits(t *testing.T) {
	checker := &fakeRoleChecker{}
	called := false
	ctx := mycontext.WithUserID(context.Background(), 42)

	res, err := HasRole(checker)(ctx, nil, passthroughResolver(&called), []string{"Admin", "ROLE_ADMIN"})
	if err != nil {
		t.Fatalf("expected the field to be authorised, got error %v", err)
	}
	if !called || res != "resolved" {
		t.Fatal("expected the field resolver to run")
	}
	if checker.lastUserID != 42 {
		t.Fatalf("expected user id 42 to reach the checker, got %d", checker.lastUserID)
	}
	if len(checker.lastRoles) != 2 || checker.lastRoles[0] != "Admin" || checker.lastRoles[1] != "ROLE_ADMIN" {
		t.Fatalf("expected the schema roles to be forwarded, got %v", checker.lastRoles)
	}
}

func TestHasRole_DeniesWhenCheckerRejects(t *testing.T) {
	checker := &fakeRoleChecker{err: errors.New("role not permitted")}
	ctx := mycontext.WithUserID(context.Background(), 7)

	_, err := HasRole(checker)(ctx, nil, func(context.Context) (any, error) {
		t.Fatal("resolver should not run for an unauthorised user")
		return nil, nil
	}, []string{"Admin"})

	if err == nil {
		t.Fatal("expected an error for an unauthorised user")
	}
	if !strings.HasPrefix(err.Error(), "forbidden:") {
		t.Fatalf("expected a 'forbidden:' error, got %q", err.Error())
	}
}

func TestHasRole_WithoutRolesSkipsChecker(t *testing.T) {
	checker := &fakeRoleChecker{}
	called := false
	ctx := mycontext.WithUserID(context.Background(), 1)

	_, err := HasRole(checker)(ctx, nil, passthroughResolver(&called), nil)
	if err != nil {
		t.Fatalf("expected the field to pass through, got error %v", err)
	}
	if !called {
		t.Fatal("expected the field resolver to run")
	}
	if checker.calls != 0 {
		t.Fatalf("expected checker not to run, ran %d times", checker.calls)
	}
}

func TestHasRole_ErrorsWhenCheckerMissing(t *testing.T) {
	ctx := mycontext.WithUserID(context.Background(), 1)

	_, err := HasRole(nil)(ctx, nil, func(context.Context) (any, error) {
		t.Fatal("resolver should not run without an RBAC checker")
		return nil, nil
	}, []string{"Admin"})

	if !errors.Is(err, ErrRoleCheckerNotConfigured) {
		t.Fatalf("expected ErrRoleCheckerNotConfigured, got %v", err)
	}
}
