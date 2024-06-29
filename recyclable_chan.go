package base

import (
	"errors"
	"go.uber.org/atomic"
	"io"
)

var (
	ErrorChannelIsFull = errors.New("channel is full")
)

type RecyclableChan struct {
	closed  bool
	desc    string
	extData any

	filledChan chan []byte

	pendingBuffer []byte
	pendingData   []byte

	isReadingAsPkt       bool
	shouldDropWhileError bool

	readReady   bool
	onReadReady func()

	flag1 atomic.Int32
	flag2 atomic.Int32
}

func NewRecyclableChan(desc string, maxPktCnt uint32, extData any) *RecyclableChan {
	return &RecyclableChan{
		false, desc, extData, make(chan []byte, maxPktCnt),
		nil, nil, false, false, false, nil,
		atomic.Int32{}, atomic.Int32{},
	}
}

func (c *RecyclableChan) Desc() string {
	return c.desc
}

func (c *RecyclableChan) ExtData() any {
	return c.extData
}

func (c *RecyclableChan) SetExtData(extData any) {
	c.extData = extData
}

func (c *RecyclableChan) SetReadingAsPkt(asPkt bool) *RecyclableChan {
	c.isReadingAsPkt = asPkt
	return c
}

func (c *RecyclableChan) SetShouldDropWhileError(should bool) *RecyclableChan {
	c.shouldDropWhileError = should
	return c
}

func (c *RecyclableChan) SetOnReadReady(callback func()) *RecyclableChan {
	c.onReadReady = callback
	return c
}

func (c *RecyclableChan) ReadStream(b []byte) (n int, err error) {
	if err = c.preRead(); err != nil {
		return
	}

	return c.readStream2Buffer(b)
}

func (c *RecyclableChan) ReadPacket() (b []byte, err error) {
	if err = c.preRead(); err != nil {
		return
	}

	return c.readPkt()
}

func (c *RecyclableChan) IsClosed() bool {
	return c.closed
}

func (c *RecyclableChan) IsReadReady() bool {
	return c.readReady
}

func (c *RecyclableChan) Flag1() int32 {
	return c.flag1.Load()
}

func (c *RecyclableChan) SetFlag1(flag1 int32) {
	c.flag1.Store(flag1)
}

func (c *RecyclableChan) Flag2() int32 {
	return c.flag2.Load()
}

func (c *RecyclableChan) SetFlag2(flag2 int32) {
	c.flag2.Store(flag2)
}

//////////////////////////// implementation of io.Writer

func (c *RecyclableChan) Write(data []byte) (n int, err error) {
	if c.closed {
		return 0, io.EOF
	}
	if len(c.filledChan) == cap(c.filledChan) {
		return 0, ErrorChannelIsFull
	}

	n = len(data)
	buffer := make([]byte, 0, n)
	buffer = append(buffer, data...)

	c.filledChan <- buffer
	return
}

//////////////////////////// implementation of io.Reader

func (c *RecyclableChan) Read(b []byte) (n int, err error) {
	if err = c.preRead(); err != nil {
		return
	}

	if c.isReadingAsPkt {
		return c.readPkt2Buffer(b)
	} else {
		return c.readStream2Buffer(b)
	}
}

//////////////////////////// implementation of io.Closer

func (c *RecyclableChan) Close() error {
	if c != nil && !c.closed {
		c.closed = true

		//buffer := make([]byte, 0, 1)
		//c.filledChan <- buffer

		close(c.filledChan)
	}

	return nil
}

//////////////////////////// private

func (c *RecyclableChan) preRead() (err error) {
	if c.closed {
		return io.EOF
	}

	if !c.readReady {
		c.readReady = true
		if c.onReadReady != nil {
			c.onReadReady()
		}
	}

	if c.closed {
		return io.EOF
	} else {
		return nil
	}
}

func (c *RecyclableChan) readPkt() (b []byte, err error) {
	if c.closed {
		return nil, io.EOF
	}

	buffer := <-c.filledChan

	if c.closed {
		return nil, io.EOF
	}

	b = append(b, buffer...)
	return
}

func (c *RecyclableChan) readPkt2Buffer(b []byte) (n int, err error) {
	if c.closed {
		return 0, io.EOF
	}

	var buffer []byte
	if c.pendingBuffer == nil {
		buffer = <-c.filledChan
	} else {
		buffer = c.pendingBuffer
		c.pendingBuffer = nil
	}

	if c.closed {
		return 0, io.EOF
	}

	if len(b) < len(buffer) {
		if !c.shouldDropWhileError {
			c.pendingBuffer = buffer
		}
		return 0, io.ErrShortBuffer
	}

	n = copy(b, buffer)
	return
}

func (c *RecyclableChan) readStream2Buffer(b []byte) (n int, err error) {
	if c.closed {
		return 0, io.EOF
	}

	var buffer []byte
	if c.pendingData != nil {
		buffer = c.pendingData
	} else {
		buffer = <-c.filledChan

		if c.closed {
			return 0, io.EOF
		}
	}

	n = copy(b, buffer)

	if n == len(buffer) { // buffer or pendingBuffer has been used up
		c.pendingData = nil
		c.pendingBuffer = nil
	} else { // there is still some pending bytes in buffer
		if c.pendingBuffer == nil {
			c.pendingBuffer = buffer
		}
		c.pendingData = buffer[n:]
	}

	return
}
