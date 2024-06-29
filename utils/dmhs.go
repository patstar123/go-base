package utils

import (
	"avd/meeting/base/logger"
	"context"
	"crypto/rand"
	"github.com/jxskiss/base62"
	"github.com/redis/go-redis/v9"
	"io"
	"time"
)

// DMHS 双机热备 Dual-Machine Hot Standby
type DMHS struct {
	randomSecret string
	ctx          context.Context
	rdb          redis.UniversalClient
	instanceName string
	timeout      time.Duration
}

func NewDMHS(ctx context.Context, rdb redis.UniversalClient,
	instanceName string, timeout time.Duration) *DMHS {
	if timeout < 500*time.Millisecond {
		timeout = 500 * time.Millisecond
	}

	return &DMHS{
		randomSecret: randomSecret(),
		ctx:          ctx,
		rdb:          rdb,
		instanceName: instanceName,
		timeout:      timeout,
	}
}

// CheckAndWaitStandby 检查并等待就绪
func (d *DMHS) CheckAndWaitStandby() error {
	logger.Infow("DMHS: waiting...")

	for {
		err, instance, lastTime := d.getOnlineInstance()
		if err != nil {
			return err
		}

		if instance == "" || time.Now().Sub(lastTime) >= d.timeout {
			err, succ := d.updateOnlineTime()
			if err != nil {
				return err
			}

			if succ {
				logger.Warnw("DMHS: activated", nil)
				go d.keepAliveLoop()
				return nil
			}
		}

		time.Sleep(d.timeout / 3)
	}
}

// //////////////////////////////// Redis字段
//
//	. 互斥锁: key=<name>-dmhs-lock value=<randomSecret>
//	. 运行在实例信息: key=<name>-dmhs-running fields: {
//		"instance": "JwnbRbbzmOBJtQ0WwA8EjIeK2s8rc1pAL0AeDfKQGomB"
//		"updateTime": 1715067355257,
//	}

func (d *DMHS) keepAliveLoop() {
	for {
		time.Sleep(d.timeout / 2)

		err, instance, lastTime := d.getOnlineInstance()
		if err != nil {
			panic("keepAliveLoop.getOnlineInstance failed: " + err.Error())
		}

		if instance != d.randomSecret && time.Now().Sub(lastTime) < d.timeout {
			panic("keepAliveLoop logical error")
		}

		err, succ := d.updateOnlineTime()
		if err != nil {
			panic("keepAliveLoop.updateOnlineTime failed: " + err.Error())
		}

		if !succ {
			logger.Warnw("logical error: updateOnlineTime failed", nil)
		}
	}
}

func (d *DMHS) updateOnlineTime() (error, bool) {
	// 尝试获取分布式锁
	lockKey := d.instanceName + "-dmhs-lock"
	locked, err := d.rdb.SetNX(d.ctx, lockKey, d.randomSecret, 5*time.Second).Result()
	if err != nil {
		logger.Warnw("updateOnlineTime, lock failed", err)
		return err, false
	}
	if !locked {
		return nil, false
	}

	// 获取锁成功，执行写操作
	key := d.instanceName + "-dmhs-running"
	ret := d.rdb.HMSet(d.ctx, key, "instance", d.randomSecret, "updateTime", time.Now().UnixMilli())
	if ret.Err() != nil {
		logger.Warnw("updateOnlineTime, update failed", ret.Err())
		d.rdb.Del(d.ctx, lockKey)
		return ret.Err(), false
	}

	// 写操作完成后释放锁
	_, err = d.rdb.Del(d.ctx, lockKey).Result()
	if err != nil {
		logger.Warnw("updateOnlineTime, unlock failed", err)
		return err, true
	}

	return nil, true
}

func (d *DMHS) getOnlineInstance() (error, string, time.Time) {
	key := d.instanceName + "-dmhs-running"
	ret := d.rdb.HMGet(d.ctx, key, "instance", "updateTime")
	if ret.Err() != nil {
		logger.Warnw("getOnlineInstance failed", ret.Err())
		return ret.Err(), "", time.Time{}
	}

	values := ret.Val()
	if len(values) == 0 || values[0] == nil {
		return nil, "", time.Time{}
	}

	return nil, values[0].(string), GetUnmarshalTime(values[1])
}

func randomSecret() string {
	// 256 bit secret
	buf := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, buf)
	// cannot error
	if err != nil {
		panic("could not read random")
	}
	return base62.EncodeToString(buf)
}
