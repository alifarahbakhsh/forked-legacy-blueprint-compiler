// Blueprint: auto-generated by Blueprint Core plugin
package proc3

import (
	"spec/services"
	"context"
)

type WebServiceImpl struct {
	service *services.WebServiceImpl
}
func NewWebServiceImpl(handler *services.WebServiceImpl) *WebServiceImpl {
	return &WebServiceImpl{service:handler}
}

func (this *WebServiceImpl) Leaf(ctx context.Context, a int64) (int64, error) {
	return this.service.Leaf(ctx, a)
}

func (this *WebServiceImpl) Hello(ctx context.Context, world string) (string, error) {
	return this.service.Hello(ctx, world)
}

