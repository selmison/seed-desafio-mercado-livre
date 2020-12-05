package mercadolivre

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	UserName string `json:"user_name" validate:"required,not_blank,email"`
	Password string `validate:"required,not_blank,min=6"`
}

type LoginResponse struct {
	Token     string
	ExpiresAt time.Time
}

// Validate validates LoginRequest.
func (l LoginRequest) Validate() error {
	return Validate(l)
}

// Login authenticates a login.
func (s *service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	stmt, err := s.db.Preparex(`SELECT * FROM users WHERE name=$1`)
	msgError := "service.login"
	if err != nil {
		err := fmt.Errorf("%w: %v", ErrInternalServer, err)
		return nil, errors.Wrap(err, msgError)
	}

	user := User{}
	err = stmt.QueryRowx(req.UserName).StructScan(&user)
	errLoginFailed := fmt.Errorf("%w: %s's credentials are not correct", ErrLoginFailed, req.UserName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(errLoginFailed, msgError)
		}
		err := fmt.Errorf("%w: %v", ErrInternalServer, err)
		return nil, errors.Wrap(err, msgError)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.Wrap(errLoginFailed, msgError)
	}

	var response *LoginResponse
	response, err = createToken(user.ID, "myJWTSecretKey")
	if err != nil {
		err := fmt.Errorf("%w: %v", ErrInternalServer, err)
		return nil, errors.Wrap(err, msgError)
	}
	return response, nil
}

func createToken(userID, jwtSecretKey string) (*LoginResponse, error) {
	var err error
	expiresAt := time.Now().Add(time.Minute * 15)
	claims := jwt.StandardClaims{
		ExpiresAt: expiresAt.Unix(),
		Id:        userID,
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := at.SignedString([]byte(jwtSecretKey))
	if err != nil {
		return nil, err
	}
	return &LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}
