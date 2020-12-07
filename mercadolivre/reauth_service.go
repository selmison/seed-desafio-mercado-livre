package mercadolivre

import (
	"context"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type ReAuthRequest struct {
	TknStr string `json:"token" validate:"required,not_blank"`
}

// Validate validates ReAuthRequest.
func (r ReAuthRequest) Validate() error {
	return Validate(r)
}

// ReAuth reauthenticates a login.
func (s *service) ReAuth(ctx context.Context, req ReAuthRequest) (*AuthResponse, error) {
	msgError := "service.re_auth"

	claims := &jwt.StandardClaims{}
	var tkn *jwt.Token
	tkn, err := jwt.ParseWithClaims(req.TknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("myJWTSecretKey"), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, errors.Wrap(fmt.Errorf("%w: %v", ErrAuthFailed, err), msgError)
		}
		return nil, ValidationErrorsResponse{
			&ValidationErrorResponse{
				FailedField: "re_auth_request.token",
				Condition:   err.Error(),
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
				FailedField: "re_auth_request.token",
				Condition:   "elapsed_time should be less than 30s",
				ActualValue: until.String(),
			},
		}
	}

	expiresAt := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = expiresAt.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tknStr, err := token.SignedString([]byte("myJWTSecretKey"))
	if err != nil {
		err := fmt.Errorf("%w: %v", ErrInternalServer, err)
		return nil, errors.Wrap(err, msgError)
	}

	return &AuthResponse{
		TknStr:    tknStr,
		ExpiresAt: expiresAt,
	}, nil
}
