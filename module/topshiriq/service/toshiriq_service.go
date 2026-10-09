package topshiriq_service

import (
	"context"
	topshiriq_dto "task-manager/module/topshiriq/dto"
)

type TopshiriqService interface {
	Create(ctx context.Context, req *topshiriq_dto.Create) (*topshiriq_dto.Response, error)
	Update(ctx context.Context, id int64, req *topshiriq_dto.Update) (*topshiriq_dto.Response, error)
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*topshiriq_dto.Response, error)
	List(ctx context.Context, filter *topshiriq_dto.Filter) (*topshiriq_dto.TopshiriqListResponse, error)
}

