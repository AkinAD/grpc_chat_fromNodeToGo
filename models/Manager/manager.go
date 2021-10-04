package manager

import (
	"context"
)

type Manager struct {
	Ctx context.Context
	Logger Logger
}

func NewManager() Manager {
	logger := NewLogger()

	logger.Info.Print("Initializing Manager...")
	return Manager{
		Ctx: context.Background(),
		Logger: *logger,
	}
}