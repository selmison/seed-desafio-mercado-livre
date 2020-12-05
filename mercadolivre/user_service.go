package mercadolivre

import (
	"context"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type UserRequest struct {
	Name     string `validate:"required,not_blank,email,should_be_unique"`
	Password string `validate:"required,not_blank,min=6"`
}

type UserResponse struct {
	Name      string
	CreatedAt time.Time
}

// Validate validates UserRequest.
func (u UserRequest) Validate() error {
	return Validate(u)
}

// User represents a single user.
// ID should be globally unique.
type User struct {
	ID        string
	Name      string
	Password  string
	CreatedAt time.Time `db:"created_at"`
}

// UserPost creates user.
func (s *service) UserPost(ctx context.Context, user UserRequest) (string, error) {
	stmt, err := s.db.Prepare("INSERT INTO users (id, name, password, created_at) VALUES ($1, $2, $3, $4)")
	msgError := "service.user_post"
	if err != nil {
		return "", errors.Wrap(err, msgError)
	}
	now := time.Now()
	layout := "2006-01-02 15:04:05"
	id := uuid.New().String()
	var hash string
	hash, err = hashPassword(user.Password)
	if err != nil {
		return "", errors.Wrap(err, msgError)
	}
	_, err = stmt.Exec(
		id,
		user.Name,
		hash,
		now.Format(layout))
	if err != nil {
		return "", errors.Wrap(err, msgError)
	}
	return id, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
