package app

type ModulesContract interface{}

type Service struct {
	customerRepository CustomerRepository
	modules            ModulesContract
}

func NewService(
	customerRepository CustomerRepository,
	modules ModulesContract,
) *Service {
	if customerRepository == nil {
		panic("customerRepository cannot be nil")
	}
	if modules == nil {
		panic("modules cannot be nil")
	}

	return &Service{
		customerRepository: customerRepository,
		modules:            modules,
	}
}
