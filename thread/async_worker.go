package thread

import (
	"avd/meeting/base"
	"avd/meeting/base/logger"
	"avd/meeting/base/utils"
	"bytes"
	"runtime"
	"strconv"
	"time"
)

type AsyncHandler interface {
	// IsCurrentInWorker 当前执行(协程)是否为工作队列本身
	//	(该方法性能不佳,不能频繁调用)
	IsCurrentInWorker() bool

	// Post 向工作队列投递任务
	Post(f func()) base.Result

	// PostDelayed 延期向工作队列投递任务
	PostDelayed(f func(), delay time.Duration) base.Result
}

type AsyncWorker struct {
	name    string
	selfGID uint64
	running bool
	tasks   *taskQueue
	maxTask int
}

func NewAsyncWorker(ownerName string) *AsyncWorker {
	return &AsyncWorker{"worker@" + ownerName, 0, false, nil, -1}
}

func (w *AsyncWorker) SetTaskMaxLimit(max int) *AsyncWorker {
	if w.tasks != nil {
		logger.Warnw("AsyncWorker has been running, ignore this call(SetTaskMaxLimit)", nil)
	} else {
		w.maxTask = max
	}
	return w
}

// RunInNewThreadUntilReady 创建一个新协程运行当前工作队列
//
//	(阻塞当前执行,直到工作队列准备就绪)
func (w *AsyncWorker) RunInNewThreadUntilReady() base.Result {
	return w.RunInNewThreadUntilReady2(false, false)
}

func (w *AsyncWorker) RunInNewThreadUntilReady2(listenPanic bool, abort bool) base.Result {
	if w.running {
		return base.LOGICAL_ERROR.AppendMsg("had been running worker")
	}

	sync := NewSync()
	defer sync.Close()

	go func() {
		if listenPanic {
			defer utils.ListenPanic(abort)
		}

		w.prepareLooper()
		sync.Notify()
		w.doLooper()
	}()

	res, _ := sync.Wait()
	return res
}

// RunLoop 在当前协程中运行工作队列
//
//	(将阻塞当前执行,直到调用`stop`)
func (w *AsyncWorker) RunLoop() base.Result {
	if w.running {
		return base.LOGICAL_ERROR.AppendMsg("had been running worker")
	}

	w.prepareLooper()
	w.doLooper()
	return base.SUCCESS
}

// NotifyQuit 通知工作队列退出
func (w *AsyncWorker) NotifyQuit() base.Result {
	w.running = false
	return base.SUCCESS
}

// StopUntilQuit 通知工作队列退出
//
//	(将阻塞当前执行,直到队列中`立即型`任务都执行完并退出)
func (w *AsyncWorker) StopUntilQuit(timeout time.Duration) base.Result {
	if !w.running {
		return base.SUCCESS
	}

	if w.IsCurrentInWorker() {
		w.running = false
		return base.SUCCESS
	}

	sync := NewSync()
	defer sync.Close()

	w.tasks.queue <- func() {
		w.running = false
		sync.Notify()
	}

	res, _ := sync.WaitWithTimeout(timeout)
	return res
}

// IsCurrentInWorker 当前执行(协程)是否为工作队列本身
//
//	(该方法性能不佳,不能频繁调用)
func (w *AsyncWorker) IsCurrentInWorker() bool {
	if !w.running {
		return false
	}

	return getGID() == w.selfGID
}

// Post 向工作队列投递任务
func (w *AsyncWorker) Post(f func()) base.Result {
	if w.tasks == nil || w.tasks.closed {
		logger.Warnw("Post("+w.name+") ACTION_ILLEGAL", nil)
		return base.ACTION_ILLEGAL
	}
	if len(w.tasks.queue) >= w.tasks.maxCap {
		logger.Warnw("Post("+w.name+") full", nil)
		return base.ACTION_CANCELED
	}

	w.tasks.queue <- f
	return base.SUCCESS
}

// PostDelayed 延期向工作队列投递任务
func (w *AsyncWorker) PostDelayed(f func(), delay time.Duration) base.Result {
	if w.tasks == nil || w.tasks.closed {
		logger.Warnw("PostDelayed("+w.name+") ACTION_ILLEGAL", nil)
		return base.ACTION_ILLEGAL
	}

	time.AfterFunc(delay, func() {
		w.Post(f)
	})

	return base.SUCCESS
}

//////////////////////////////////// private functions

const taskQueCap int = 100

type taskQueue struct {
	closed bool
	maxCap int
	queue  chan func()
}

func (w *AsyncWorker) waitState(running bool, waiter chan byte) base.Result {
	return w.waitState2(running, waiter, 0xffffffff*time.Millisecond)
}

func (w *AsyncWorker) waitState2(running bool, waiter chan byte, timeout time.Duration) base.Result {
	t := time.After(timeout)
	for {
		select {
		case <-t:
			return base.ACTION_TIMEOUT
		case <-waiter:
			if running == w.running {
				return base.SUCCESS
			} else {
				return base.ACTION_CANCELED
			}
		}
	}
}

func (w *AsyncWorker) prepareLooper() {
	queCap := taskQueCap
	if w.maxTask > 0 {
		queCap = w.maxTask
	}

	w.selfGID = getGID()
	w.tasks = &taskQueue{false, queCap,
		make(chan func(), queCap)}
	w.running = true
}

func (w *AsyncWorker) doLooper() {
	logger.Infow("worker(" + w.name + ") started")

	for w.running {
		select {
		case fun := <-w.tasks.queue:
			fun()
		}
	}

	w.running = false
	w.selfGID = 0
	w.tasks.closed = true
	close(w.tasks.queue)
	w.tasks.queue = nil
	w.tasks = nil

	logger.Infow("worker(" + w.name + ") stopped")
}

// 下列代码源于 Brad Fitzpatrick 的 http/2 库。它被整合进了 Go 1.6 中，
//
//	仅仅被用于调试而非常规开发，它的性能不佳
func getGID() uint64 {
	buf := make([]byte, 64)
	buf = buf[:runtime.Stack(buf, false)]
	buf = bytes.TrimPrefix(buf, []byte("goroutine "))
	buf = buf[:bytes.IndexByte(buf, ' ')]
	n, _ := strconv.ParseUint(string(buf), 10, 64)
	return n
}
