package graceful

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"time"
)

// GracefulShutdown предоставляет механизм для graceful shutdown
type GracefulShutdown struct {
	ctx     context.Context
	cancel  context.CancelFunc
	wg      *sync.WaitGroup
	timeout time.Duration
	errors  error
}

// Task представляет интерфейс для задач, которые могут быть запущены и остановлены
type Task interface {
	Run(context.Context) error
}

// NewGracefulShutdown создает новый экземпляр GracefulShutdown
func NewGracefulShutdown(ctx context.Context, timeout time.Duration) *GracefulShutdown {
	ctx, cancel := context.WithCancel(ctx)
	return &GracefulShutdown{
		ctx:     ctx,
		cancel:  cancel,
		wg:      &sync.WaitGroup{},
		timeout: timeout,
	}
}

// AddTask добавляет задачу в GracefulShutdown
func (gs *GracefulShutdown) AddTask(task Task) {
	gs.wg.Add(1)

	go func() {
		defer gs.wg.Done()
		err := task.Run(gs.ctx)
		gs.errors = errors.Join(gs.errors, err)
	}()
}

// Wait ожидает сигнала завершения и затем ожидает завершения всех задач
func (gs *GracefulShutdown) Wait(sig ...os.Signal) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, sig...)
	<-stop
	gs.cancel()

	// Создаем канал для отслеживания завершения задач
	done := make(chan struct{})
	go func() {
		gs.wg.Wait()
		close(done)
	}()

	// Ожидаем завершения задач или истечения времени ожидания
	select {
	case <-done:
		return gs.errors
	case <-time.After(gs.timeout):
		return gs.errors
	}
}
