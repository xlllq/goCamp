package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type requestState int32

const (
	CLOSED    requestState = 0
	OPEN      requestState = 1
	HALF_OPEN requestState = 2
)

type statistic struct {
	total   int
	failure int
	round   int64
}

type reqMeta struct {
	status     requestState
	window     []statistic
	sleepSince time.Time
}

type Hystrix struct {
	meta map[string]*reqMeta

	maxWindow int
	failRatio float32
	sleepTime time.Duration

	rwMutex sync.RWMutex
}

var Hys Hystrix = Hystrix{
	meta:      make(map[string]*reqMeta),
	maxWindow: 10,
	failRatio: 0.5,
	sleepTime: time.Second * 1,
	rwMutex:   sync.RWMutex{},
}

func (hys *Hystrix) getFailureRatio(str string) float32 {
	hys.rwMutex.RLock()
	windows := hys.meta[str].window
	hys.rwMutex.RUnlock()
	sumTotal, sumFailue := 0, 0
	for k, stc := range windows {
		fmt.Println(k, stc.failure, stc.total)
		sumFailue += stc.failure
		sumTotal += stc.total
	}
	if sumTotal == 0 {
		return 0
	} else {
		return float32(sumFailue) / float32(sumTotal)
	}
}

func (hys *Hystrix) addCount(str string, success bool, reqTime time.Time) {
	//TODO: Will Lock be the bottleNeck?
	offset := reqTime.Unix() % int64(hys.maxWindow)
	hys.rwMutex.Lock()
	if !success {
		hys.meta[str].window[offset].failure++
	}
	hys.meta[str].window[offset].total++
	hys.rwMutex.Unlock()
}

func (hys *Hystrix) addNewRequest(str string) {
	//创建window, 设置status为CLOSED
	statistic := make([]statistic, hys.maxWindow)
	newReqMeta := reqMeta{CLOSED, statistic, time.Now()}
	hys.rwMutex.Lock()
	hys.meta[str] = &newReqMeta
	hys.rwMutex.Unlock()
}

func (hys *Hystrix) modifyWindow(str string, reqTime time.Time) {
	offset := reqTime.Unix() % int64(hys.maxWindow)
	round := reqTime.Unix() / int64(hys.maxWindow)
	//擦除超过一圈的位置
	hys.rwMutex.Lock()
	if round != hys.meta[str].window[offset].round {
		hys.meta[str].window[offset] = statistic{0, 0, round}
	}
	hys.rwMutex.Unlock()
}

func (hys *Hystrix) clearWindow(str string) {
	hys.meta[str].window = make([]statistic, hys.maxWindow)
}

func (hys *Hystrix) Do(str string, f func() error) (bool, error) {
	//初始化请求
	hys.rwMutex.RLock()
	_, ok := hys.meta[str]
	hys.rwMutex.RUnlock()
	if !ok {
		hys.addNewRequest(str)
	}
	//清理window
	reqTime := time.Now()
	hys.modifyWindow(str, reqTime)

	hys.rwMutex.RLock()
	status := hys.meta[str].status
	hys.rwMutex.RUnlock()

	fmt.Println("current status:", status)
	var err error = nil
	if status == OPEN {
		//不请求，直接返回err
		err = errors.New("reject, status is OPEN")
	} else {
		//CLOSE 或 HALF_OPEN 都发送请求
		if err = f(); err != nil {
			hys.addCount(str, false, reqTime)
		} else {
			hys.addCount(str, true, reqTime)
		}
	}
	//统计windows的请求失败率，判断是否熔断
	failureRatio := hys.getFailureRatio(str)
	fmt.Println("failureRatio:", failureRatio)

	hys.rwMutex.Lock()
	defer hys.rwMutex.Unlock()
	if status == CLOSED {
		if failureRatio > hys.failRatio {
			fmt.Println("circuit breaker triggered")
			hys.meta[str].status = OPEN
			hys.meta[str].sleepSince = time.Now()
		}
	} else if status == OPEN {
		//超过睡眠时间，改为试探状态
		if time.Since(hys.meta[str].sleepSince) > hys.sleepTime {
			hys.meta[str].status = HALF_OPEN
		}
	} else if status == HALF_OPEN {
		//根据试探结果改变状态
		if err != nil {
			hys.meta[str].status = OPEN
			hys.meta[str].sleepSince = time.Now()
		} else {
			hys.meta[str].status = CLOSED
			hys.clearWindow(str)
		}
	}
	fmt.Println("--------------------")
	return err != nil, err
}
