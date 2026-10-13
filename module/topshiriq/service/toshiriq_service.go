package topshiriq_service

import (
	"context"
	"errors"
	topshiriq_dto "task-manager/module/topshiriq/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TopshiriqService interface {
	Create(ctx context.Context, req *topshiriq_dto.Create) (*topshiriq_dto.Response, error)
	Update(ctx context.Context, id int64, req *topshiriq_dto.Update) (*topshiriq_dto.Response, error)
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*topshiriq_dto.Response, error)
	List(ctx context.Context, filter *topshiriq_dto.Filter) (*topshiriq_dto.TopshiriqListResponse, error)
}

type topshiriqService struct {
	db *pgxpool.Pool
}

func NewTopshiriqService(db *pgxpool.Pool) TopshiriqService {
	return &topshiriqService{
		db: db,
	}
}

func (service *topshiriqService) Create(ctx context.Context, req *topshiriq_dto.Create) (*topshiriq_dto.Response, error) {
	var res topshiriq_dto.Response

	err := service.db.QueryRow(ctx, `
		INSERT INTO topshiriqs (name, description, category_id, hudud_id, file, didline)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING *
	`, req.Name, req.Description, req.CategoryID, req.HududID, req.File, req.Didline).Scan(
		&res.ID, &res.Name, &res.Description, &res.CategoryID, &res.HududID, &res.File, &res.Didline, &res.Status, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *topshiriqService) Update(ctx context.Context, id int64, req *topshiriq_dto.Update) (*topshiriq_dto.Response, error) {
	var res topshiriq_dto.Response

	err := service.db.QueryRow(ctx, `
		UPDATE topshiriqs
		SET name = $1,
			description = $2,
			category_id = $3,
			hudud_id = $4,
			file = $5,
			didline = $6
		WHERE id = $7
		RETURNING *
	`,
		req.Name, req.Description, req.CategoryID, req.HududID, req.File, req.Didline, id).Scan(
		&res.ID, &res.Name, &res.Description, &res.CategoryID, &res.HududID, &res.File, &res.Didline, &res.Status, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *topshiriqService) Delete(ctx context.Context, id int64) error {
	cmd, err := service.db.Exec(ctx, `
		DELETE FROM topshiriqs
		WHERE id = $1
	`, id)

	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (service *topshiriqService) GetByID(ctx context.Context, id int64) (*topshiriq_dto.Response, error) {
	var res topshiriq_dto.Response

	err := service.db.QueryRow(ctx, `
		SELECT id, name, description, category_id,
			hudud_id, file, didline, status, created_at, updated_at
		FROM topshiriqs
		WHERE id = $1
	`, id).
		Scan(&res.ID, &res.Name, &res.Description, &res.CategoryID,
			&res.HududID, &res.File, &res.Didline, &res.Status,
			&res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *topshiriqService) List(ctx context.Context, filter *topshiriq_dto.Filter) (*topshiriq_dto.TopshiriqListResponse, error) {

	if filter == nil {
		filter = &topshiriq_dto.Filter{Limit: 20}
	}

	if filter.Limit < 1 {
		filter.Limit = 20
	}

	if filter.Limit > 100 {
		filter.Limit = 100
	}

	if filter.Direction == "" {
		filter.Direction = "next"
	}

	if filter.Direction != "next" && filter.Direction != "prev" {
		return nil, errors.New("direction must be next or prev")
	}

	if filter.Direction == "prev" && filter.Cursor == nil {
		return nil, errors.New("cursor is required for prev direction")
	}

	const nextQuery = `
		SELECT id, name, description, category_id, hudud_id,
			COALESCE(file, ''), didline::TEXT, COALESCE(status, ''),
			COALESCE(created_at::TEXT, ''), COALESCE(updated_at::TEXT, '')
		FROM topshiriqs
		WHERE ($1::BIGINT = 0 OR id = $1)
			AND name ILIKE '%' || $2 || '%'
			AND ($3::TEXT = '' OR description ILIKE '%' || $3 || '%')
			AND ($4::BIGINT = 0 OR category_id = $4)
			AND ($5::BIGINT = 0 OR hudud_id = $5)
			AND ($6::TEXT = '' OR didline = NULLIF($6::TEXT, '')::TIMESTAMP)
			AND ($7::BIGINT IS NULL OR id < $7)
		ORDER BY id DESC
		LIMIT $8
	`
	const prevQuery = `
		SELECT id, name, description, category_id, hudud_id,
			COALESCE(file, ''), didline::TEXT, COALESCE(status, ''),
			COALESCE(created_at::TEXT, ''), COALESCE(updated_at::TEXT, '')
		FROM topshiriqs
		WHERE ($1::BIGINT = 0 OR id = $1)
			AND name ILIKE '%' || $2 || '%'
			AND ($3::TEXT = '' OR description ILIKE '%' || $3 || '%')
			AND ($4::BIGINT = 0 OR category_id = $4)
			AND ($5::BIGINT = 0 OR hudud_id = $5)
			AND ($6::TEXT = '' OR didline = NULLIF($6::TEXT, '')::TIMESTAMP)
			AND id > $7
		ORDER BY id ASC
		LIMIT $8
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
		filter.Description,
		filter.CategoryID,
		filter.HududID,
		filter.Didline,
		filter.Cursor,
		filter.Limit+1,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := make([]topshiriq_dto.Response, 0, filter.Limit)

	for rows.Next() {
		var item topshiriq_dto.Response

		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.CategoryID,
			&item.HududID,
			&item.File,
			&item.Didline,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		)

		if err != nil {
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

	response := &topshiriq_dto.TopshiriqListResponse{Data: items}
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
