package flow

import "github.com/pkg/errors"

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Usecase() error {
	return errors.New("not implemented")
}
