package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductChecker struct {
	db *pgxpool.Pool
}

func NewProductChecker(db *pgxpool.Pool) *ProductChecker {
	return &ProductChecker{
		db: db,
	}
}

func (p *ProductChecker) Exists(ctx context.Context, productID string) (bool, error) {
	// Оператор EXISTS возвращает логическое значение (boolean)
	selectProductQuery := `SELECT EXISTS(SELECT 1 FROM products WHERE product_id = @product_id);`
	args := pgx.NamedArgs{"product_id": productID}
	var exists bool

	err := p.db.QueryRow(ctx, selectProductQuery, args).Scan(&exists)
	// Ошибки ErrNoRows здесь быть не может по определению оператора EXISTS
	if err != nil {
		return exists, fmt.Errorf("query product existence: %w", err)
	}
	return exists, nil
}

type CustomerChecker struct {
	db *pgxpool.Pool
}

func NewCustomerChecker(db *pgxpool.Pool) *CustomerChecker {
	return &CustomerChecker{
		db: db,
	}
}

func (p *CustomerChecker) Exists(ctx context.Context, customerID string) (bool, error) {
	// Оператор EXISTS возвращает логическое значение (boolean)
	selectProductQuery := `SELECT EXISTS(SELECT 1 FROM customers WHERE customer_id = @customer_id);`
	args := pgx.NamedArgs{"customer_id": customerID}
	var exists bool

	err := p.db.QueryRow(ctx, selectProductQuery, args).Scan(&exists)
	// Ошибки ErrNoRows здесь быть не может по определению оператора EXISTS
	if err != nil {
		return exists, fmt.Errorf("query customer existence: %w", err)
	}
	return exists, nil
}
