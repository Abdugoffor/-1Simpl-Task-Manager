package config

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SeedCategories(ctx context.Context, db *pgxpool.Pool, count int) (int64, error) {
	result, err := db.Exec(ctx, `
		INSERT INTO categorys(name, slug, description, is_active)
		SELECT
			'Category ' || number,
			'category-' || gen_random_uuid(),
			'Description for category ' || number,
			TRUE
		FROM generate_series(1, $1) AS number
	`, count)
	{
		if err != nil {
			return 0, err
		}
	}

	return result.RowsAffected(), nil
}
