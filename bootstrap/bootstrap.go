package bootstrap

import (
	"context"
	"log"

	"github.com/gmcorenet/framework/kernel"
)

type Application struct {
	kernel *kernel.Kernel
	ctx    context.Context
}

func New(k *kernel.Kernel) *Application {
	return &Application{
		kernel: k,
		ctx:    context.Background(),
	}
}

func (a *Application) Run() error {
	if err := a.kernel.Bootstrap(a.ctx); err != nil {
		return err
	}

	log.Println("Application ready")
	return a.kernel.Run()
}

func Boot(k *kernel.Kernel) error {
	return k.Bootstrap(context.Background())
}

