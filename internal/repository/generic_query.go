package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type rowScanner interface {
	Scan(dest ...any) error
}

func queryOneByField[T any](ctx context.Context, pool *pgxpool.Pool, table string, columns string, field string,
	value any, scanFn func(row rowScanner) (T, error)) (T, error) {
	var result T

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", columns, table, field)

	err := WithTxRetry(ctx, pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query, value)
		entity, err := scanFn(row)
		if err != nil {
			return err
		}
		result = entity
		return nil
	})
	if err != nil {
		return result, err
	}

	return result, nil
}

func queryManyByField[T any](ctx context.Context, pool *pgxpool.Pool, table string, columns, field string, value any,
	orderBy string, scanFn func(row rowScanner) (T, error)) ([]T, error) {
	var result []T

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", columns, table, field)
	if orderBy != "" {
		query += " ORDER BY " + orderBy
	}

	err := WithTxRetry(ctx, pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, value)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			entity, err := scanFn(rows)
			if err != nil {
				return err
			}
			result = append(result, entity)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
