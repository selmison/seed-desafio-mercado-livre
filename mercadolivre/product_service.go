package mercadolivre

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type ProductRequest struct {
	Name       string    `validate:"required,not_blank"`
	Price      *float32  `validate:"required,gt=0"`
	Amount     *int16    `validate:"required,gte=0"`
	Features   []Feature `validate:"required,min=2"`
	Desc       string    `validate:"required,max=100"`
	CategoryID string    `json:"category_id" validate:"required,not_blank,should_exist"`
	CreatedAt  time.Time
}

// Validate validates ProductRequest.
func (u ProductRequest) Validate() error {
	return Validate(u)
}

type ProductResponse struct {
	ID   string
	Name string
}

// Product represents a single Product.
// ID should be globally unique.
type Product struct {
	ID         string
	Name       string
	Price      float32
	Amount     int16
	Features   []Feature
	Desc       string
	CategoryID string
	Category   Category
	CreatedAt  time.Time `db:"created_at"`
}

// Featrue represents a single Product's Feature.
type Feature struct {
	Type    string
	Name    string
	Details string
}

// ProductPost creates Product.
func (s *service) ProductPost(ctx context.Context, product ProductRequest) (productID string, err error) {
	msgError := "service.product_post"
	var tx *sql.Tx
	tx, err = s.db.Begin()
	if err != nil {
		return "", errors.Wrap(err, msgError)
	}
	defer func() {
		if p := recover(); p != nil {
			productID = ""
			err = rollback(tx, err)
			panic(p)
		} else if err != nil {
			productID = ""
			err = rollback(tx, err)
		} else {
			if err := tx.Commit(); err != nil {
				productID = ""
				err = errors.Wrap(err, msgError)
			}
		}
	}()

	pStmt, err := tx.Prepare("INSERT INTO products (id, name, price, amount, description, category_id, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)")
	if err != nil {
		return "", errors.Wrap(err, msgError)
	}
	productID = uuid.New().String()
	now := time.Now()
	layout := "2006-01-02 15:04:05"
	_, err = pStmt.Exec(
		productID,
		product.Name,
		product.Price,
		product.Amount,
		product.Desc,
		product.CategoryID,
		now.Format(layout))
	if err != nil {
		return "", errors.Wrap(err, msgError)
	}

	tStmt, err := tx.Prepare("INSERT INTO types_of_features (id, product_id, type) VALUES ($1, $2, $3)")
	if err != nil {
		return "", errors.Wrap(err, msgError)
	}
	fStmt, err := tx.Prepare("INSERT INTO features (id, type_id, name, details) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return "", errors.Wrap(err, msgError)
	}
	for _, feature := range product.Features {
		typeID := uuid.New().String()
		_, err = tStmt.Exec(
			typeID,
			productID,
			feature.Type)
		if err != nil {
			return "", errors.Wrap(err, msgError)
		}

		_, err = fStmt.Exec(
			uuid.New().String(),
			typeID,
			feature.Name,
			feature.Details)
		if err != nil {
			return "", errors.Wrap(err, msgError)
		}
	}

	return
}

func rollback(tx *sql.Tx, err error) error {
	if e := tx.Rollback(); e != nil && e != sql.ErrTxDone {
		if err != nil {
			return fmt.Errorf("%w: %v", err, e)
		}
		return e
	}
	return err
}
