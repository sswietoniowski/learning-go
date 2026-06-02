package app

type ModulesContract interface{}

type Service struct{}

func NewService() *Service {
	return &Service{}
}
