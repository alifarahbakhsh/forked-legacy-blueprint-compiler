package services

import (
	"context"
)

type WebService interface {
	Leaf(ctx context.Context, a int64) (int64, error)
	Hello(ctx context.Context, world string) (string, error)
}

type WebServiceImpl struct {
	leafService LeafService
}

func NewWebServiceImpl(leafService LeafService) *WebServiceImpl {
	return &WebServiceImpl{leafService}
}

func (w *WebServiceImpl) Leaf(ctx context.Context, a int64) (int64, error) {
	return w.leafService.Leaf(ctx, a)
}

func (w *WebServiceImpl) Hello(ctx context.Context, world string) (string, error) {
	return "Hello" + world, nil
}