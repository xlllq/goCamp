package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var rpcPerSecond struct {
	thisCicle time.Time
	num       int
}

var mutex sync.Mutex = sync.Mutex{}

//模拟rpc请求, 负载越高失败概率越高
func simulateRPC() error {
	mutex.Lock()
	defer mutex.Unlock()
	if time.Since(rpcPerSecond.thisCicle).Seconds() < 1 {
		rpcPerSecond.num++
		//QPS > 100 时必然失败
		if rand.Int()%100+rpcPerSecond.num > 100 {
			return errors.New("rpc failed")
		}
		return nil
	}
	rpcPerSecond.thisCicle = time.Now()
	rpcPerSecond.num = 0
	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	_, err := Hys.Do("rpc1", simulateRPC)
	if err != nil {
		w.WriteHeader(404)
		fmt.Println(err)
	}
}

func main() {
	rand.Seed(time.Now().Unix())
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8668", nil)
}
