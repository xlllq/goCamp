package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//Test: curl localhost:9553/service

type ctxKey string

func serviceHandler(w http.ResponseWriter, r *http.Request) {
	//pass "UserAgent" as value of context
	ctx := context.WithValue(r.Context(), ctxKey("UserAgent"), r.UserAgent())
	if err := service(ctx); err != nil {
		fmt.Println("Request failed:", err)
	} else {
		fmt.Println("service succeeded")
	}
}

func handleSignal(server *http.Server, cancelAllRequest context.CancelFunc, exit chan<- int) {
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT)
	//Block until siganl arrives
	sig := <-ch
	fmt.Println("received signal:", sig.String())
	cancelAllRequest()
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		exit <- 1
	}()
	if err := server.Shutdown(timeoutCtx); err != nil {
		fmt.Println(err)
	}
}

func main() {
	server := http.Server{
		Addr:    ":9553",
		Handler: http.DefaultServeMux,
	}
	//Create base context
	ctx, cancel := context.WithCancel(context.Background())
	server.BaseContext = func(l net.Listener) context.Context {
		return ctx
	}
	//Handle Linux Signal
	exit := make(chan int)
	go handleSignal(&server, cancel, exit)

	//Start HTTP Server
	http.HandleFunc("/service", serviceHandler)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Closing Server...")
	}
	<-exit
	fmt.Println("Server closed, program exits")
}
