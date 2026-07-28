package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/pkg/apperrors"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(r UserRepository) *UserService {
	return &UserService{repo: r}
}

// REGISTER
func (s *UserService) Register(ctx context.Context, req *dto.RegisterUserRequest) error {
	_, err := s.repo.GetByEmail(ctx, req.Email)
	if err == nil {
		return apperrors.ErrUserAlreadyExists
	}
	//hash pass
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = s.repo.Create(ctx, &models.User{ID: uuid.New().String(), Name: req.Name, Email: req.Email, PasswordHash: string(hashedPass)})

	return err

}

// LOGIN
func (s *UserService) Login(ctx context.Context, req *dto.LoginUserRequest) (*dto.LoginResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	return &dto.LoginResponse{
		User: dto.UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil

}
