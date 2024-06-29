package utils

import (
	"time"
)

const MAXIMUM_CACHED_FRAMES = 60

type FrameRateStatist struct {
	frames          [MAXIMUM_CACHED_FRAMES]time.Time
	iter            int
	totalFrameCount uint64
}

func NewFrameRateStatist() *FrameRateStatist {
	return &FrameRateStatist{}
}

func (f *FrameRateStatist) IncomingFrame() {
	f.frames[f.iter] = time.Now()
	f.iter = (f.iter + 1) % MAXIMUM_CACHED_FRAMES
	f.totalFrameCount++
}

func (f *FrameRateStatist) AverageFrameRate() float64 {
	var frameRate float64
	oldest_ts := f.frames[f.iter]
	if !oldest_ts.IsZero() {
		latest_iter := (f.iter + MAXIMUM_CACHED_FRAMES - 1) % MAXIMUM_CACHED_FRAMES
		latest_ts := f.frames[latest_iter]
		if !latest_ts.IsZero() && time.Now().Sub(latest_ts).Milliseconds() < 1000 {
			duration := latest_ts.Sub(oldest_ts)
			frameRate = float64(MAXIMUM_CACHED_FRAMES-2) / float64(duration.Seconds())
		}
	}
	return frameRate
}

func (f *FrameRateStatist) TotalFrameCount() uint64 {
	return f.totalFrameCount
}

func (f *FrameRateStatist) Reset() {
	f.frames = [MAXIMUM_CACHED_FRAMES]time.Time{}
	f.iter = 0
	f.totalFrameCount = 0
}
