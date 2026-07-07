package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/ecommerce/user/internal/domain"
	"github.com/ecommerce/user/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	repo        repository.UserRepository
	jwtSecret   string
	jwtDuration time.Duration
}

func NewUserUsecase(repo repository.UserRepository, jwtSecret string, jwtDuration time.Duration) *UserUsecase {
	return &UserUsecase{
		repo:        repo,
		jwtSecret:   jwtSecret,
		jwtDuration: jwtDuration,
	}
}

func (uc *UserUsecase) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error) {
	existing, _ := uc.repo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &domain.User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		Name:      req.Name,
		Role:      "customer",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}

func (uc *UserUsecase) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	user, err := uc.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(uc.jwtDuration).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(uc.jwtSecret))
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	user.Password = ""
	return &domain.LoginResponse{
		Token: tokenStr,
		User:  user,
	}, nil
}

func (uc *UserUsecase) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user.Password = ""
	return user, nil
}

func (uc *UserUsecase) Update(ctx context.Context, req *domain.UpdateUserRequest) (*domain.User, error) {
	user, err := uc.repo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	user.UpdatedAt = time.Now()

	if err := uc.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}
