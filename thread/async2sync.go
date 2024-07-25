package thread

import (
	"time"
)

type Sync struct {
	closed bool
	event  chan any
}

func NewSync() *Sync {
	return &Sync{false, make(chan any, 1)}
}

func (s *Sync) Close() {
	if !s.closed {
		s.closed = true
		close(s.event)
	}
}

func (s *Sync) Reset() {
	if s.closed {
		return
	}

	for len(s.event) > 0 {
		<-s.event
	}
}

func (s *Sync) Wait() (base.Result, any) {
	return s.WaitWithTimeout(-1)
}

func (s *Sync) WaitWithTimeout(timeout time.Duration) (base.Result, any) {
	if s.closed {
		return base.ACTION_ILLEGAL, nil
	}

	if timeout < 0 {
		timeout = 0xffffffff * time.Second
	}

	t := time.After(timeout)
	for {
		select {
		case <-t:
			return base.ACTION_TIMEOUT, nil
		case data := <-s.event:
			return base.SUCCESS, data
		}
	}
}

func (s *Sync) Notify() base.Result {
	return s.NotifyWithData(nil)
}

func (s *Sync) NotifyWithData(data any) base.Result {
	if s.closed {
		return base.ACTION_ILLEGAL
	}
	if len(s.event) > 0 {
		return base.LOGICAL_ERROR
	}

	s.event <- data
	return base.SUCCESS
}

func SyncByHandler(handler AsyncHandler, impl func(), timeout time.Duration) base.Result {
	if handler == nil {
		return base.ACTION_ILLEGAL
	}

	if timeout < 0 {
		timeout = 0xffffffff * time.Second
	}

	sync := NewSync()
	defer sync.Close()

	handler.Post(func() {
		impl()
		sync.Notify()
	})

	res, _ := sync.WaitWithTimeout(timeout)
	return res
}

func SyncByHandlerD(handler AsyncHandler, impl func() any,
	timeout time.Duration) (base.Result, any) {
	if handler == nil {
		return base.ACTION_ILLEGAL, nil
	}

	if timeout < 0 {
		timeout = 0xffffffff * time.Second
	}

	sync := NewSync()
	defer sync.Close()

	handler.Post(func() {
		data := impl()
		sync.NotifyWithData(data)
	})

	return sync.WaitWithTimeout(timeout)
}

func SyncCloseByHandler(handler AsyncHandler, impl func(), timeout time.Duration) base.Result {
	if timeout < 0 {
		timeout = 0xffffffff * time.Second
	}

	if handler != nil {
		if handler.IsCurrentInWorker() {
			impl()
		} else {
			sync := NewSync()
			defer sync.Close()

			handler.Post(func() {
				impl()
				sync.Notify()
			})

			res, _ := sync.WaitWithTimeout(timeout)
			if !res.IsOk() {
				return res
			}
		}
	} else {
		impl()
	}

	return base.SUCCESS
}

func SyncCall(impl func(callback base.Callback), timeout time.Duration) base.Result {
	if timeout < 0 {
		timeout = 0xffffffff * time.Second
	}

	sync := NewSync()
	defer sync.Close()

	impl(func(result base.Result) {
		sync.NotifyWithData(result)
	})

	res, internalRes := sync.WaitWithTimeout(timeout)
	if !res.IsOk() {
		return res
	}

	return internalRes.(base.Result)
}
