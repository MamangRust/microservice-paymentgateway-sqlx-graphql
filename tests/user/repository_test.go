package user_test

import (
	"context"
	"fmt"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-test"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	ts     *tests.TestSuite
	dbPool *sqlx.DB
	repo   repository.UserCommandRepository
	userID int
}

func (s *UserRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user"))

	pool, err := sqlx.ConnectContext(s.ts.Ctx, "pgx", s.ts.DBURL)
	s.Require().NoError(err)
	s.dbPool = pool
	s.repo = repository.NewUserCommandRepository(pool)
}

func (s *UserRepositoryTestSuite) TearDownSuite() {
	s.dbPool.Close()
	s.ts.Teardown()
}

func (s *UserRepositoryTestSuite) Test1_CreateUser() {
	req := &requests.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     fmt.Sprintf("john.doe.%d@example.com", time.Now().UnixNano()),
		Password:  "password123",
	}

	user, err := s.repo.CreateUser(context.Background(), req)
	s.NoError(err)
	s.NotNil(user)
	s.Equal(req.FirstName, user.Firstname)
	s.Equal(req.Email, user.Email)
	s.userID = int(user.UserID)
}

func (s *UserRepositoryTestSuite) Test2_FindById() {
	s.Require().NotZero(s.userID)
	ctx := context.Background()

	pool, _ := sqlx.ConnectContext(s.ts.Ctx, "pgx", s.ts.DBURL)
	defer pool.Close()
	queryRepo := repository.NewUserQueryRepository(pool)

	found, err := queryRepo.FindById(ctx, s.userID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(s.userID, int(found.UserID))
}

func TestUserRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(UserRepositoryTestSuite))
}
