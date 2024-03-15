package authservice

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/F1zm0n/uni-auth/repository"
)

type Service interface {
	Register(ctx context.Context, user User) error
	Login(ctx context.Context, user User) (token string, err error)
}

type basicService struct {
	db repository.Repository
}

func NewBasicService(db repository.Repository) Service {
	return basicService{
		db: db,
	}
}

var (
	ErrWrongEmailFmt     = errors.New("wrong email format")
	ErrWrongPassFmt      = errors.New("password should be minimum 8 length long")
	ErrInsertingUser     = errors.New("error inserting user")
	ErrInvalidCreds      = errors.New("invalid credentials")
	ErrGeneratingToken   = errors.New("error generating jwt token")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

type User struct {
	ID       uuid.UUID
	Email    string
	Password string
}

func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (s basicService) Register(ctx context.Context, user User) error {
	repoUser, err := validateUserCreds(user)
	if err != nil {
		return err
	}

	err = s.db.InsertUser(ctx, repoUser)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) ||
			strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return ErrUserAlreadyExists
		}

		return ErrInsertingUser
	}
	return nil
}

func (s basicService) Login(ctx context.Context, user User) (string, error) {
	resUser, err := s.db.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return "", ErrInvalidCreds
	}
	if err := bcrypt.CompareHashAndPassword(resUser.Password, []byte(user.Password)); err != nil {
		return "", ErrInvalidCreds
	}
	tok, err := NewToken(user)
	if err != nil {
		return "", ErrGeneratingToken
	}

	return tok, nil
}

func NewToken(user User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(viper.GetDuration("auth.tokenttl")).Unix()

	tokenString, err := token.SignedString([]byte(viper.GetString("auth.secret")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func validateUserCreds(user User) (repository.User, error) {
	if ok := validateEmail(user.Email); !ok {
		return repository.User{}, ErrWrongEmailFmt
	}
	if len(user.Password) < 8 {
		return repository.User{}, ErrWrongPassFmt
	}
	passHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return repository.User{}, err
	}
	return repository.User{
		ID:       uuid.New(),
		Email:    user.Email,
		Password: passHash,
	}, nil
}

func New(logger log.Logger, repo repository.Repository) Service {
	var svc Service
	{
		svc = NewBasicService(repo)
		svc = LoggingMiddleware(logger)(svc)
		svc = InstrumentingMiddleware()(svc)
	}
	return svc
}
