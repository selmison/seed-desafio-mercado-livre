package mercadolivre

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	jwtKit "github.com/go-kit/kit/auth/jwt"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type AuthRequest struct {
	UserName string `json:"user_name" validate:"required,not_blank,email"`
	Password string `validate:"required,not_blank,min=6"`
}

// Validate validates AuthRequest.
func (a AuthRequest) Validate() error {
	return Validate(a)
}

type AuthResponse struct {
	TknStr    string `json:"token"`
	ExpiresAt time.Time
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

// ReAuth reauthenticates a user.
func (s *service) ReAuth(ctx context.Context) (*AuthResponse, error) {
	msgError := "service.re_auth"

	tknStr, ok := ctx.Value(jwtKit.JWTTokenContextKey).(string)
	if !ok {
		return nil, ValidationErrorsResponse{
			&ValidationErrorResponse{
				Condition: ErrMissingToken.Error(),
			},
		}
	}

	claims := &jwt.StandardClaims{}
	var tkn *jwt.Token
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("myJWTSecretKey"), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, errors.Wrap(fmt.Errorf("%w: %v", ErrAuthFailed, err), msgError)
		}
		return nil, ValidationErrorsResponse{
			&ValidationErrorResponse{
				Condition: err.Error(),
			},
		}
	}
	if !tkn.Valid {
		return nil, errors.Wrap(fmt.Errorf("%w: %v", ErrAuthFailed, "token should be valid"), msgError)
	}

	until := time.Until(time.Unix(claims.ExpiresAt, 0))
	if until > 30*time.Second {
		return nil, ValidationErrorsResponse{
			&ValidationErrorResponse{
				Condition:   "elapsed_time should be less than 30s",
				ActualValue: until.String(),
			},
		}
	}

	expiresAt := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = expiresAt.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tknStr, err = token.SignedString([]byte("myJWTSecretKey"))
	if err != nil {
		err := fmt.Errorf("%w: %v", ErrInternalServer, err)
		return nil, errors.Wrap(err, msgError)
	}

	return &AuthResponse{
		TknStr:    tknStr,
		ExpiresAt: expiresAt,
	}, nil
}
