package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProductChecker struct {
	db *pgxpool.Pool
}

func NewPostgresProductChecker(db *pgxpool.Pool) *PostgresProductChecker {
	return &PostgresProductChecker{
		db: db,
	}
}

func (p *PostgresProductChecker) Exists(ctx context.Context, productID string) (bool, error) {
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

type PostgresCustomerChecker struct {
	db *pgxpool.Pool
}

func NewPostgresCustomerChecker(db *pgxpool.Pool) *PostgresCustomerChecker {
	return &PostgresCustomerChecker{
		db: db,
	}
}

func (p *PostgresCustomerChecker) Exists(ctx context.Context, customerID string) (bool, error) {
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
