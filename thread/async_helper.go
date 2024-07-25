package thread

import (
	"time"

	"github.com/livekit/protocol/logger"
)

type AsyncCaller interface {
	IsInited() bool
	GetHandler() AsyncHandler
}

func AsyncCall(caller AsyncCaller, callback base.Callback, f func(), checking ...func() bool) {
	if !checkAndCallback(caller.IsInited(),
		"init first", callback) {
		return
	}
	for _, c := range checking {
		if !checkAndCallback(c(), "init first2", callback) {
			return
		}
	}

	caller.GetHandler().Post(func() {
		if !checkAndCallback(caller.IsInited(),
			"init first3", callback) {
			return
		}
		for _, c := range checking {
			if !checkAndCallback(c(), "init first4", callback) {
				return
			}
		}

		f()
	})
}

func AsyncCallDelay(caller AsyncCaller, callback base.Callback, delay time.Duration, f func(), checking ...func() bool) {
	if !checkAndCallback(caller.IsInited(),
		"init first", callback) {
		return
	}
	for _, c := range checking {
		if !checkAndCallback(c(), "init first2", callback) {
			return
		}
	}

	caller.GetHandler().PostDelayed(func() {
		if !checkAndCallback(caller.IsInited(),
			"init first3", callback) {
			return
		}
		for _, c := range checking {
			if !checkAndCallback(c(), "init first4", callback) {
				return
			}
		}

		f()
	}, delay)
}

func GoCall(caller AsyncCaller, callback base.Callback, f func(), checking ...func() bool) {
	if !checkAndCallback(caller.IsInited(),
		"init first", callback) {
		return
	}
	for _, c := range checking {
		if !checkAndCallback(c(), "init first2", callback) {
			return
		}
	}

	go func() {
		if !checkAndCallback(caller.IsInited(),
			"init first3", callback) {
			return
		}
		for _, c := range checking {
			if !checkAndCallback(c(), "init first4", callback) {
				return
			}
		}

		f()
	}()
}

func GoCall2(callback base.Callback, f func(), checking ...func() bool) {
	for _, c := range checking {
		if !checkAndCallback(c(), "condition false", callback) {
			return
		}
	}

	go func() {
		for _, c := range checking {
			if !checkAndCallback(c(), "condition false2", callback) {
				return
			}
		}

		f()
	}()
}

func checkAndCallback(reqCond bool, message string, callback base.Callback) bool {
	if reqCond {
		return true
	} else {
		logger.Warnw(message, nil)
		callback.On(base.ACTION_ILLEGAL.AppendMsg(message))
		return false
	}
}
