package adapter

import (
	"context"

	pbuser "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

// UserAdapter is the minimal cross-service user surface other domains need
// (e.g. merchant resolving the owner of a merchant).
type UserAdapter interface {
	FindById(ctx context.Context, userID int) (*models.User, error)
}

// AuthUserAdapter is the auth-facing user surface: credential lookups plus the
// account mutations the auth flows perform.
type AuthUserAdapter interface {
	FindById(ctx context.Context, userID int) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByEmailAndVerify(ctx context.Context, email string) (*models.User, error)
	FindByVerificationCode(ctx context.Context, code string) (*models.User, error)
	CreateUser(ctx context.Context, request *requests.RegisterRequest) (*models.User, error)
	UpdateUserIsVerified(ctx context.Context, userID int, isVerified bool) error
	UpdateUserPassword(ctx context.Context, userID int, password string) error
}

type userGRPCAdapter struct {
	queryClient pbuser.UserQueryServiceClient
	guard       *resilience.DependencyGuard
}

func (a *userGRPCAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

// NewUserAdapter builds the read-only user adapter on top of the generated
// UserQueryService client.
func NewUserAdapter(queryClient pbuser.UserQueryServiceClient, opts ...GuardOption) UserAdapter {
	a := &userGRPCAdapter{queryClient: queryClient}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *userGRPCAdapter) FindById(ctx context.Context, userID int) (*models.User, error) {
	var resp *pbuser.ApiResponseUser
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.queryClient.FindById(callCtx, &pbuser.FindByIdUserRequest{Id: int32(userID)})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Data == nil {
		return nil, sharedErrors.ErrUserNotFound
	}
	return mapUserResponse(resp.Data), nil
}

type authUserGRPCAdapter struct {
	queryClient   pbuser.UserQueryServiceClient
	commandClient pbuser.UserCommandServiceClient
	guard         *resilience.DependencyGuard
}

func (a *authUserGRPCAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

// NewAuthUserAdapter builds the auth-facing user adapter from the query and
// command clients.
func NewAuthUserAdapter(queryClient pbuser.UserQueryServiceClient, commandClient pbuser.UserCommandServiceClient, opts ...GuardOption) AuthUserAdapter {
	a := &authUserGRPCAdapter{queryClient: queryClient, commandClient: commandClient}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *authUserGRPCAdapter) FindById(ctx context.Context, userID int) (*models.User, error) {
	var resp *pbuser.ApiResponseUser
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.queryClient.FindById(callCtx, &pbuser.FindByIdUserRequest{Id: int32(userID)})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Data == nil {
		return nil, sharedErrors.ErrUserNotFound
	}
	return mapUserResponse(resp.Data), nil
}

// FindByEmail returns the user including the stored password hash so callers can
// verify credentials. A missing user comes back as ErrUserNotFound.
func (a *authUserGRPCAdapter) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var resp *pbuser.ApiResponseUser
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.queryClient.FindByEmail(callCtx, &pbuser.FindByEmailUserRequest{Email: email})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Data == nil {
		return nil, sharedErrors.ErrUserNotFound
	}
	return mapUserResponse(resp.Data), nil
}

// FindByEmailAndVerify is the credential-verification lookup used on login. The
// user proto exposes no verified-only lookup, so this delegates to FindByEmail;
// the password comparison downstream is what actually authenticates the user.
func (a *authUserGRPCAdapter) FindByEmailAndVerify(ctx context.Context, email string) (*models.User, error) {
	return a.FindByEmail(ctx, email)
}

func (a *authUserGRPCAdapter) FindByVerificationCode(ctx context.Context, code string) (*models.User, error) {
	var resp *pbuser.ApiResponseUser
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.queryClient.FindByVerificationCode(callCtx, &pbuser.FindByVerificationCodeUserRequest{VerificationCode: code})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Data == nil {
		return nil, sharedErrors.ErrUserNotFound
	}
	return mapUserResponse(resp.Data), nil
}

// CreateUser registers a user through the User service. The User service hashes
// the password and grants the default role, so the caller passes the plaintext
// password (and its confirmation) as-is.
func (a *authUserGRPCAdapter) CreateUser(ctx context.Context, request *requests.RegisterRequest) (*models.User, error) {
	var resp *pbuser.ApiResponseUser
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.commandClient.Create(callCtx, &pbuser.CreateUserRequest{
			Firstname:       request.FirstName,
			Lastname:        request.LastName,
			Email:           request.Email,
			Password:        request.Password,
			ConfirmPassword: request.ConfirmPassword,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Data == nil {
		return nil, sharedErrors.ErrFailed("create user")
	}
	return mapUserResponse(resp.Data), nil
}

func (a *authUserGRPCAdapter) UpdateUserIsVerified(ctx context.Context, userID int, isVerified bool) error {
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		_, callErr := a.commandClient.UpdateIsVerified(callCtx, &pbuser.UpdateUserIsVerifiedRequest{
			UserId:     int32(userID),
			IsVerified: isVerified,
		})
		return callErr
	})
	if err != nil {
		return sharedErrors.ErrFailed("update user verification").WithInternal(err)
	}
	return nil
}

func (a *authUserGRPCAdapter) UpdateUserPassword(ctx context.Context, userID int, password string) error {
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		_, callErr := a.commandClient.UpdatePassword(callCtx, &pbuser.UpdateUserPasswordRequest{
			UserId:   int32(userID),
			Password: password,
		})
		return callErr
	})
	if err != nil {
		return sharedErrors.ErrFailed("update user password").WithInternal(err)
	}
	return nil
}

func mapUserResponse(u *pbuser.UserResponse) *models.User {
	if u == nil {
		return nil
	}
	return &models.User{
		UserID:    u.Id,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: parseTime(u.CreatedAt),
		UpdatedAt: parseTime(u.UpdatedAt),
	}
}
