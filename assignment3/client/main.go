package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

//客户端模拟高并发请求
var totalRequest int = 3000

func main() {
	rand.Seed(time.Now().Unix())
	totalSucceed := 0

	for i := 0; i < totalRequest; i++ {
		resp, err := http.Get("http://localhost:8668")
		if err != nil {
			fmt.Println("request error")
			continue
		}
		fmt.Println("response:", resp.StatusCode)
		if resp.StatusCode == 200 {
			totalSucceed++
		}
		resp.Body.Close()
		duration := rand.Int()%25 - i%1000/100
		time.Sleep(time.Millisecond * time.Duration(duration))
	}
	fmt.Println("totalSucceed:", totalSucceed, "/", totalRequest)
}
