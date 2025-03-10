// Blueprint: auto-generated by Blueprint Core plugin
package proc1

import (
	"spec/services"
	"context"
)

type LeafServiceImpl struct {
	service *services.LeafServiceImpl
}
func NewLeafServiceImpl(handler *services.LeafServiceImpl) *LeafServiceImpl {
	return &LeafServiceImpl{service:handler}
}

func (this *LeafServiceImpl) Leaf(ctx context.Context, a int64) (int64, error) {
	return this.service.Leaf(ctx, a)
}

func (this *LeafServiceImpl) Object(ctx context.Context, obj services.LeafObject) (services.LeafObject, error) {
	return this.service.Object(ctx, obj)
}

