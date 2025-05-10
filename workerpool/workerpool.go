package workerpool

import (
	"context"
	"github.com/tony-zhuo/cex/pkg/logger"
	"sync"
)

// IWorker 接口定義 worker 必須實現的方法
type IWorker interface {
	Name() string
	Health() bool
	Process() error
}

// WorkerPool 結構
type WorkerPool struct {
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	logger  *logger.Logger
	workers []IWorker // 註冊的自定義 worker 實現
	done    chan struct{}
}

// NewWorkerPool 創建新的 worker pool
func NewWorkerPool() *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		ctx:     ctx,
		cancel:  cancel,
		wg:      sync.WaitGroup{},
		logger:  logger.GetInstance().WithGroup("worker pool"),
		workers: make([]IWorker, 0),
	}
}

// Start 啟動 worker pool
func (wp *WorkerPool) Start() {
	for _, worker := range wp.workers {
		wp.wg.Add(1)
		go func() {
			defer func() {
				if err := recover(); err != nil {
					wp.logger.Panic("panic: %v", err)
				}
				wp.wg.Done()
			}()
			if err := worker.Process(); err != nil {
				wp.logger.Error("worker process err: %v", err)
				return
			}
		}()
	}
	wp.wg.Wait()
	wp.done <- struct{}{}
}

func (wp *WorkerPool) Register(worker IWorker) {
	if worker == nil {
		return
	}
	wp.workers = append(wp.workers, worker)
}

// Close 關閉 worker pool
func (wp *WorkerPool) Close() {
	wp.logger.Info("worker pool closing")
	wp.cancel()
	<-wp.done
	wp.logger.Info("worker pool closed")
}
