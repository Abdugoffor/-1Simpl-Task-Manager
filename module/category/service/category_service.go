package category_service

import (
	"context"

	"task-manager/helper"
	category_dto "task-manager/module/category/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryService interface {
	Create(ctx context.Context, req *category_dto.Create) (*category_dto.Response, error)
	Update(ctx context.Context, id int64, req *category_dto.Update) (*category_dto.Response, error)
	Delete(ctx context.Context, id int64) error
	Show(ctx context.Context, id int64) (*category_dto.Response, error)
	List(ctx context.Context, filter *category_dto.Filter) (*category_dto.CategoryListResponse, error)
}

type categoryService struct {
	db *pgxpool.Pool
}

func NewCategoryService(db *pgxpool.Pool) CategoryService {
	return &categoryService{
		db: db,
	}
}

func (service *categoryService) Create(ctx context.Context, req *category_dto.Create) (*category_dto.Response, error) {
	var res category_dto.Response

	slug := req.Slug
	{
		if slug == "" {
			slug = helper.Slug(req.Name)
		}
	}

	err := service.db.QueryRow(ctx, `
		INSERT INTO categorys(name, slug, description, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, slug, COALESCE(description, ''), is_active, created_at, updated_at
	`, req.Name, slug, req.Description, req.IsActive).Scan(
		&res.ID, &res.Name, &res.Slug, &res.Description, &res.IsActive, &res.CreatedAt, &res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *categoryService) Update(ctx context.Context, id int64, req *category_dto.Update) (*category_dto.Response, error) {
	var res category_dto.Response

	err := service.db.QueryRow(ctx, `
		UPDATE categorys
		SET name = $1,  description = $2, is_active = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING id, name, slug, COALESCE(description, ''), is_active, created_at, updated_at
	`, req.Name, req.Description, req.IsActive, id).Scan(
		&res.ID, &res.Name, &res.Slug, &res.Description, &res.IsActive, &res.CreatedAt, &res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *categoryService) Delete(ctx context.Context, id int64) error {
	cmd, err := service.db.Exec(ctx, "DELETE FROM categorys WHERE id = $1", id)
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
func (service *categoryService) Show(ctx context.Context, id int64) (*category_dto.Response, error) {
	var res category_dto.Response

	err := service.db.QueryRow(ctx, `
		SELECT id, name, slug, COALESCE(description, ''), is_active, created_at, updated_at
		FROM categorys
		WHERE id = $1
	`, id).Scan(
		&res.ID, &res.Name, &res.Slug, &res.Description, &res.IsActive, &res.CreatedAt, &res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *categoryService) List(ctx context.Context, filter *category_dto.Filter) (*category_dto.CategoryListResponse, error) {
	if filter == nil {
		filter = &category_dto.Filter{Limit: 20}
	}

	if filter.Limit < 1 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	const nextQuery = `
		SELECT id, name, slug, COALESCE(description, ''), is_active, created_at, updated_at
		FROM categorys
		WHERE ($1::bigint = 0 OR id = $1)
			AND name ILIKE '%' || $2 || '%'
			AND slug ILIKE '%' || $3 || '%'
			AND ($4::text = '' OR description ILIKE '%' || $4 || '%')
			AND ($5::boolean IS NULL OR is_active = $5)
			AND ($6::bigint IS NULL OR id < $6)
		ORDER BY id DESC
		LIMIT $7
	`
	const prevQuery = `
		SELECT id, name, slug, COALESCE(description, ''), is_active, created_at, updated_at
		FROM categorys
		WHERE ($1::bigint = 0 OR id = $1)
			AND name ILIKE '%' || $2 || '%'
			AND slug ILIKE '%' || $3 || '%'
			AND ($4::text = '' OR description ILIKE '%' || $4 || '%')
			AND ($5::boolean IS NULL OR is_active = $5)
			AND id > $6
		ORDER BY id ASC
		LIMIT $7
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
		filter.Slug,
		filter.Description,
		filter.IsActive,
		filter.Cursor,
		filter.Limit+1,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := make([]category_dto.Response, 0, filter.Limit)

	for rows.Next() {
		var item category_dto.Response

		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Slug,
			&item.Description,
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

	response := &category_dto.CategoryListResponse{Data: items}
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
