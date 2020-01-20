package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cacheServiceAddress := flag.String("cache", "localhost:9090", "<host>:<port>")
	flag.Parse()
	log.Printf("Cache service address to connect to: %s", *cacheServiceAddress)

	sysStop := make(chan os.Signal, 1)
	defer close(sysStop)
	signal.Notify(sysStop, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGABRT, syscall.SIGINT)

	stop := make(chan interface{})
	go func() {
		<-sysStop
		close(stop)
	}()

	consumer := NewConsumer(stop)
	err := consumer.Dial(*cacheServiceAddress)
	if err != nil {
		log.Fatalf("Failed to connect to %s, error: %v", *cacheServiceAddress, err)
	}

	consumer.Run()

	log.Printf("Exiting.")
}
