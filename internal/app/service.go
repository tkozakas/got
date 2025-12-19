package app

import "context"

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetCryptoPrice(ctx context.Context, ticker string) (float64, error) {
	return 100.0, nil
}
