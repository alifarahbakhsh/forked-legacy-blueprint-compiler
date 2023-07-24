package services

import (
	"context"
)

type HelloEvaluatorsService interface {
	KickTheTires(ctx context.Context) (string, error)
}

type HelloEvaluatorsImpl struct{}

func (h *HelloEvaluatorsImpl) KickTheTires(ctx context.Context) (string, error) {
	return "The tires are kicking! Ready for Artifact Evaluation! :)", nil
}

func NewHelloEvaluatorsImpl() *HelloEvaluatorsImpl {
	return &HelloEvaluatorsImpl{}
}
