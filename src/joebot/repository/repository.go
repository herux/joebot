package repository

import "context"

type Repository interface {
	Create(ctx context.Context, model interface{}) (int, error)
}
