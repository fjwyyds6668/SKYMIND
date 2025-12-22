package appcore

import (
	"context"

	"skymind/api/smart_query"
	"syscall"
)

// App struct - 应用核心结构体
type App struct {
	Ctx           context.Context
	Quitting      bool
	InTray        bool
	Mutex         syscall.Handle
	SmartQueryAPI smart_query.ApiGroup
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}
