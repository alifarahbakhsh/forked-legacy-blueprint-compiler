package services

import (
	"context"
)

type NonLeafService interface {
	Leaf(ctx context.Context, a int64) (int64, error)
}

type NonLeafServiceImpl struct {
	leafService LeafService
}

func (nl* NonLeafServiceImpl) Leaf(ctx context.Context, a int64) (int64, error) {
	return nl.leafService.Leaf(ctx, a)
}

func NewNonLeafServiceImpl(leafService LeafService) *NonLeafServiceImpl {
	return &NonLeafServiceImpl{leafService}
}