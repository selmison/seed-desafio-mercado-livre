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

type AuthRequest struct {
	UserName string `json:"user_name" validate:"required,not_blank,email"`
	Password string `validate:"required,not_blank,min=6"`
}

type AuthResponse struct {
	TknStr    string `json:"token"`
	ExpiresAt time.Time
}

// Validate validates AuthRequest.
func (a AuthRequest) Validate() error {
	return Validate(a)
}

// Auth authenticates a user.
func (s *service) Auth(ctx context.Context, req AuthRequest) (*AuthResponse, error) {
	stmt, err := s.db.Preparex(`SELECT * FROM users WHERE name=$1`)
	msgError := "service.auth"
	if err != nil {
		err := fmt.Errorf("%w: %v", ErrInternalServer, err)
		return nil, errors.Wrap(err, msgError)
	}

	user := User{}
	err = stmt.QueryRowx(req.UserName).StructScan(&user)
	errAuthFailed := fmt.Errorf("%w: %s's credentials are not correct", ErrAuthFailed, req.UserName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(errAuthFailed, msgError)
		}
		err := fmt.Errorf("%w: %v", ErrInternalServer, err)
		return nil, errors.Wrap(err, msgError)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.Wrap(errAuthFailed, msgError)
	}

	var response *AuthResponse
	response, err = createToken(user.ID, "myJWTSecretKey")
	if err != nil {
		err := fmt.Errorf("%w: %v", ErrInternalServer, err)
		return nil, errors.Wrap(err, msgError)
	}
	return response, nil
}

func createToken(userID, jwtSecretKey string) (*AuthResponse, error) {
	var err error
	expiresAt := time.Now().Add(time.Minute * 5)
	claims := jwt.StandardClaims{
		ExpiresAt: expiresAt.Unix(),
		Id:        userID,
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tknStr, err := at.SignedString([]byte(jwtSecretKey))
	if err != nil {
		return nil, err
	}
	return &AuthResponse{
		TknStr:    tknStr,
		ExpiresAt: expiresAt,
	}, nil
}
