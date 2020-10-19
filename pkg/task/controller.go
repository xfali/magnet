// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package task

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type Handle interface {
	// 获得任务名称
	Name() string

	// 增加任务步骤，当Done/TryDone调用次数等于所有Add delta总和时视为任务完成
	Add(delta int)

	// 标识任务完成，当所有任务都完成则会清除Controller的任务标记
	// 注意：当Done调用次数大于Add的delta总数时会触发panic
	Done()

	// 与Done作用一致
	// 不同的是TryDone调用次数大于Add的delta总数时不会panic，会返回error
	TryDone() error

	// 等待任务完成
	// 当所有任务步骤都完成或者到达参数中的指定超时时间会返回，否则等待
	// 参数为负数为一直等待，直到任务完成
	Wait(time.Duration)
}

type Controller interface {
	// 判断任务是否正在执行，是返回true，否则返回false
	IsRunning(taskName string) bool

	// 增加一个任务
	// Param：delta 任务步骤数
	// Return： Handler 任务handler，用于控制任务完成情况
	// Return：error 成功返回true，否则返回false
	AddTask(taskName string) (Handle, error)

	// 查询任务Handle
	// Return： Handler 任务handler，用于控制任务完成情况
	FindTask(taskName string) Handle
}

type cleanFunc func(name string)

type entity struct {
	name        string
	waitTimeout time.Duration
	delta       int32

	once      sync.Once
	cleanFunc cleanFunc
	c         chan struct{}
}

type defaultController struct {
	taskMap sync.Map
}

func NewController() *defaultController {
	return &defaultController{}
}

func (c *defaultController) IsRunning(taskName string) bool {
	_, ok := c.taskMap.Load(taskName)
	return ok
}

func NewHandle(name string, f cleanFunc) *entity {
	return &entity{
		name:      name,
		cleanFunc: f,
		c:         make(chan struct{}),
	}
}

func (e *entity) Name() string {
	return e.name
}

func (e *entity) Add(delta int) {
	if delta < 0 {
		panic("delta cannot be negative")
	}
	atomic.AddInt32(&e.delta, int32(delta))
}

func (e *entity) TryDone() error {
	v := atomic.AddInt32(&e.delta, -1)
	if v == 0 {
		e.once.Do(func() {
			close(e.c)
			e.cleanFunc(e.name)
		})
		return nil
	} else if v < 0 {
		return errors.New("Call Done is more than delta. ")
	}
	return nil
}

func (e *entity) Done() {
	err := e.TryDone()
	if err != nil {
		panic(err)
	}
}

func (e *entity) Wait(timeout time.Duration) {
	if timeout > 0 {
		t := time.NewTimer(timeout)
		select {
		case <-e.c:
			return
		case <-t.C:
			return
		}
	} else if timeout < 0 {
		select {
		case <-e.c:
			return
		}
	}
}

func (c *defaultController) AddTask(taskName string) (Handle, error) {
	e := NewHandle(taskName, func(name string) {
		c.taskMap.Delete(name)
	})
	v, ok := c.taskMap.LoadOrStore(taskName, e)
	if ok {
		return v.(*entity), errors.New("Task: " + taskName + " is Exists")
	}
	return e, nil
}

func (c *defaultController) FindTask(taskName string) Handle {
	v, ok := c.taskMap.Load(taskName)
	if !ok {
		return nil
	}
	return v.(*entity)
}
