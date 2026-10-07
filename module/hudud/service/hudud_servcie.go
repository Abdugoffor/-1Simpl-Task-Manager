package hudud_service

import (
	"context"

	hudud_dto "task-manager/module/hudud/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HududService interface {
	Create(ctx context.Context, req *hudud_dto.Create) (*hudud_dto.Response, error)
	Update(ctx context.Context, id int64, req *hudud_dto.Update) (*hudud_dto.Response, error)
	Delete(ctx context.Context, id int64) error
	Show(ctx context.Context, id int64) (*hudud_dto.Response, error)
	List(ctx context.Context, filter *hudud_dto.Filter) (*hudud_dto.HududListResponse, error)
}

type hududService struct {
	db *pgxpool.Pool
}

func NewHududService(db *pgxpool.Pool) HududService {
	return &hududService{
		db: db,
	}
}

func (service *hududService) Create(ctx context.Context, req *hudud_dto.Create) (*hudud_dto.Response, error) {
	var res hudud_dto.Response

	err := service.db.QueryRow(ctx, `
		INSERT INTO hududs(name, is_active)
		VALUES ($1, $2)
		RETURNING id, name, is_active, created_at, updated_at
	`, req.Name, req.IsActive).Scan(
		&res.ID, &res.Name, &res.IsActive, &res.CreatedAt, &res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *hududService) Update(ctx context.Context, id int64, req *hudud_dto.Update) (*hudud_dto.Response, error) {
	var res hudud_dto.Response

	err := service.db.QueryRow(ctx, `
		UPDATE hududs
		SET name = $1, is_active = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, is_active, created_at, updated_at
	`, req.Name, req.IsActive, id).Scan(
		&res.ID, &res.Name, &res.IsActive, &res.CreatedAt, &res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *hududService) Delete(ctx context.Context, id int64) error {
	cmd, err := service.db.Exec(ctx, "DELETE FROM hududs WHERE id = $1", id)
	{
		if err != nil {
			return err
		}
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (service *hududService) Show(ctx context.Context, id int64) (*hudud_dto.Response, error) {
	var res hudud_dto.Response

	err := service.db.QueryRow(ctx, `
		SELECT id, name, is_active, created_at, updated_at
		FROM hududs
		WHERE id = $1
	`, id).Scan(
		&res.ID, &res.Name, &res.IsActive, &res.CreatedAt, &res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *hududService) List(ctx context.Context, filter *hudud_dto.Filter) (*hudud_dto.HududListResponse, error) {
	if filter == nil {
		filter = &hudud_dto.Filter{Limit: 20}
	}

	if filter.Limit < 1 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	const nextQuery = `
		SELECT id, name, is_active, created_at, updated_at
		FROM hududs
		WHERE ($1::bigint = 0 OR id = $1)
			AND name ILIKE '%' || $2 || '%'
			AND ($3::boolean IS NULL OR is_active = $3)
			AND ($4::bigint IS NULL OR id < $4)
		ORDER BY id DESC
		LIMIT $5
	`
	const prevQuery = `
		SELECT id, name, is_active, created_at, updated_at
		FROM hududs
		WHERE ($1::bigint = 0 OR id = $1)
			AND name ILIKE '%' || $2 || '%'
			AND ($3::boolean IS NULL OR is_active = $3)
			AND id > $4
		ORDER BY id ASC
		LIMIT $5
	`

	query := nextQuery
	{
		if filter.Direction == "prev" {
			query = prevQuery
		}
	}

	rows, err := service.db.Query(ctx, query,
		filter.ID,
		filter.Name,
		filter.IsActive,
		filter.Cursor,
		filter.Limit+1,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := make([]hudud_dto.Response, 0, filter.Limit)

	for rows.Next() {
		var item hudud_dto.Response

		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	hasExtra := len(items) > filter.Limit
	{
		if hasExtra {
			items = items[:filter.Limit]
		}
	}

	if filter.Direction == "prev" {

		for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
			items[left], items[right] = items[right], items[left]
		}
	}

	response := &hudud_dto.HududListResponse{Data: items}
	{
		if filter.Direction == "prev" {

			response.HasPrev = hasExtra

			response.HasNext = filter.Cursor != nil

		} else {

			response.HasPrev = filter.Cursor != nil

			response.HasNext = hasExtra
		}
	}

	if len(items) > 0 {

		if response.HasPrev {

			prevCursor := items[0].ID

			response.PrevCursor = &prevCursor
		}

		if response.HasNext {

			nextCursor := items[len(items)-1].ID

			response.NextCursor = &nextCursor
		}
	}

	return response, nil
}
