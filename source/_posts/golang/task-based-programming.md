
---
title: golang http 使用总结
categories:
- Golang
---




在复杂的业务逻辑中，一个逻辑或者说一个接口通常包含许多子逻辑，为了使得代码比较清晰，以前常使用责任链的设计模式来实现。最近发现有些场景下使用基于task的编程方式可以使得代码很清晰。这里简单记录下这种编程方式的特点以及实现方式。

责任链设计：https://refactoringguru.cn/design-patterns/chain-of-responsibility/go/example
  
主要特点：
1. 将一个复杂的任务分解成多个任务，任务之间支持串行和并行
2. Task执行的独立
3. Task执行的参数和中间数据保存Session中，使得数据能够灵活共享，需要保证协程安全
    
<img src=https://raw.githubusercontent.com/wxquare/wxquare.github.io/hexo/source/images/1ae6b34b-78d9-4e27-b7d8-cb3c89f13d7b.png width=400/> 

实现：
1. task.go定义 itask interface
2. 定义serialtask.go
3. 定义paralleltask.go
4. main主要业务逻辑定义session、任务编排、执行


```Go
package main

import (
	"context"
)

type iTask interface {
	Do(c context.Context) error
}
```


```GO
package main

import "context"

type SerialTask struct {
	tasks []iTask
}

func (s *SerialTask) Add(task iTask) {
	s.tasks = append(s.tasks, task)

}

func (S *SerialTask) Do(c context.Context) error {
	for _, t := range S.tasks {
		if err := t.Do(c); err != nil {
			return err
		}
	}
	return nil
}
```


```Go
package main

import (
	"context"
	"errors"
	"sync"
)

type ParallelTask struct {
	tasks []iTask
}

func (s *ParallelTask) Add(task iTask) {
	s.tasks = append(s.tasks, task)

}

func (S ParallelTask) Do(c context.Context) error {
	Errs := make(chan error, len(S.tasks))
	wg := sync.WaitGroup{}
	for _, t := range S.tasks {
		wg.Add(1)
		go func(i iTask) {
			defer wg.Done()
			if err := i.Do(c); err != nil {
				Errs <- err
			}
		}(t)
	}
	wg.Wait()
	if len(Errs) != 0 {
		return errors.New("parallel task error")
	}
	return nil
}
```

```Go
package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Session struct {
	ctx    context.Context
	cancel context.CancelFunc

	Param   string
	Lock    *sync.Mutex
	Errs    chan error
	Timeout time.Duration
}

func (s Session) Deadline() (deadline time.Time, ok bool) {
	return s.ctx.Deadline()
}

func (s Session) Done() <-chan struct{} {
	return s.ctx.Done()
}

func (s Session) Err() error {
	return s.ctx.Err()
}

func (s Session) Value(key interface{}) interface{} {
	return s.ctx.Value(key)
}

type Task1 struct{}

func (t Task1) Do(ctx context.Context) error {
	session, ok := ctx.(*Session)
	if !ok {
		return errors.New("38 type assertion abort")
	}
	for {
		select {
		case <-session.ctx.Done():
			return session.ctx.Err()
		default:
			time.Sleep(time.Duration(rand.Int31n(5)) * time.Second)
			session.Lock.Lock()
			defer session.Lock.Unlock()
			session.Param = "task1"
			fmt.Printf("run task1 param%s\n", session.Param)
			return nil
		}
	}
}

type Task2 struct{}

func (t Task2) Do(ctx context.Context) error {
	session, ok := ctx.(*Session)
	if !ok {
		return errors.New("type assertion abort")
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(time.Duration(rand.Int31n(5)) * time.Second)
			session.Lock.Lock()
			defer session.Lock.Unlock()
			session.Param = "task2"
			fmt.Printf("run task2 param%s\n", session.Param)
			return nil
		}
	}
}

type Task3 struct{}

func (t Task3) Do(ctx context.Context) error {
	session, ok := ctx.(*Session)
	if !ok {
		return errors.New("type assertion abort")
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(time.Duration(rand.Int31n(5)) * time.Second)
			session.Lock.Lock()
			defer session.Lock.Unlock()
			session.Param = "task3"
			fmt.Printf("run task3 param%s\n", session.Param)
			return nil
		}
	}
}

type Task4 struct{}

func (t Task4) Do(ctx context.Context) error {
	session, ok := ctx.(*Session)
	if !ok {
		return errors.New("type assertion abort")
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(time.Duration(rand.Int31n(5)) * time.Second)
			session.Lock.Lock()
			defer session.Lock.Unlock()
			session.Param = "task4"
			fmt.Printf("run task4 param%s\n", session.Param)
			return nil
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	m := SerialTask{}
	m.Add(Task1{})
	p := ParallelTask{}
	p.Add(Task2{})
	p.Add(Task3{})
	m.Add(p)
	m.Add(Task4{})

	session := Session{
		Param:   "initial",
		Lock:    &sync.Mutex{},
		Timeout: 7 * time.Second,
	}
	session.ctx, session.cancel = context.WithTimeout(context.Background(), session.Timeout)
	defer session.cancel()

	if err := m.Do(&session); err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	fmt.Printf("%+v\n", session.Param)
}

```


