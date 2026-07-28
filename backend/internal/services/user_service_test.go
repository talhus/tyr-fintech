package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/internal/services"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepository struct {
	createFunc     func(ctx context.Context, user *models.User) error
	getByIDFunc    func(ctx context.Context, id string) (*models.User, error)
	getByEmailFunc func(ctx context.Context, email string) (*models.User, error)
}

// MOCK REPO METHODS
func (m *mockUserRepository) Create(ctx context.Context, user *models.User) error {
	return m.createFunc(ctx, user)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	return m.getByIDFunc(ctx, id)
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.getByEmailFunc(ctx, email)
}

// TESTS
func TestUserService_Register(t *testing.T) {
	tests := []struct {
		name          string
		inputName     string
		inputEmail    string
		inputPassword string
		mockCreateErr error
		wantErr       bool
	}{
		{
			name:          "Success",
			inputName:     "John Doe",
			inputEmail:    "john@mail.com",
			inputPassword: "securepassword123",
			mockCreateErr: nil,
			wantErr:       false,
		},
		{
			name:          "FailedCreate",
			inputName:     "Jane Doe",
			inputEmail:    "jane@mail.com",
			inputPassword: "securepassword123",
			mockCreateErr: errors.New("db connection failure"),
			wantErr:       true,
		},
		{
			name:          "UserAlreadyExists",
			inputName:     "John Doe",
			inputEmail:    "john@mail.com",
			inputPassword: "securepassword123",
			mockCreateErr: nil,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepository{
				createFunc: func(ctx context.Context, user *models.User) error {
					return tt.mockCreateErr
				},
				getByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
					if tt.name == "UserAlreadyExists" {
						return &models.User{ID: "1", Email: email}, nil
					}
					return nil, errors.New("user not found")
				},
			}
			service := services.NewUserService(mockRepo)
			err := service.Register(context.Background(), &dto.RegisterUserRequest{
				Name:     tt.inputName,
				Email:    tt.inputEmail,
				Password: tt.inputPassword,
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func TestUserService_Login(t *testing.T) {
	hashedPass, _ := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)
	tests := []struct {
		name          string
		inputEmail    string
		inputPassword string
		mockUser      *models.User
		mockErr       error
		wantErr       bool
	}{
		{
			name:          "success",
			inputEmail:    "talha@gmail.com",
			inputPassword: "12345678",
			mockUser: &models.User{
				ID:           "1",
				Name:         "Talha",
				Email:        "talha@gmail.com",
				PasswordHash: string(hashedPass),
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:          "user not found",
			inputEmail:    "nouser@gmail.com",
			inputPassword: "12345678",
			mockErr:       errors.New("user not found"),
			wantErr:       true,
		},
		{
			name:          "password not match",
			inputEmail:    "talha@gmail.com",
			inputPassword: "wrongpassword",
			mockUser: &models.User{
				ID:           "1",
				Name:         "Talha",
				Email:        "talha@gmail.com",
				PasswordHash: string(hashedPass),
			},
			mockErr: nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepository{
				getByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
					return tt.mockUser, tt.mockErr
				},
			}
			service := services.NewUserService(mockRepo)
			_, err := service.Login(context.Background(), &dto.LoginUserRequest{
				Email:    tt.inputEmail,
				Password: tt.inputPassword,
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
