package services

import (
	"context"
)

type LeafService interface {
	Leaf(ctx context.Context, a int64) (int64, error)
	Object(ctx context.Context, obj LeafObject) (LeafObject, error)
}

type LeafServiceImpl struct {}

func (l* LeafServiceImpl) Leaf(ctx context.Context, a int64) (int64, error) {
	return a, nil
}

func (l *LeafServiceImpl) Object(ctx context.Context, obj LeafObject) (LeafObject, error) {
	return obj, nil
}

func NewLeafServiceImpl() *LeafServiceImpl {
	return &LeafServiceImpl{}
}