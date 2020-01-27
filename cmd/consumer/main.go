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
	argMin := flag.Int64("argMin", 0, "lower bound for randomly generated arg, used for requesting cache service. Int64")
	argMax := flag.Int64("argMax", 1000, "upper bound for randomly generated arg, used for requesting cache service. Int64")
	flag.Parse()
	if *argMin > *argMax {
		log.Fatalf("Invalid arguments: argMin should be less than argMax. Provided: argMin=%d, argMax=%d", *argMin, *argMax)
	}
	log.Printf("Cache service address to connect to: %s. Arg range: %d - %d", *cacheServiceAddress, *argMin, *argMax)

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

	consumer.Run(*argMin, *argMax)

	log.Printf("Exiting.")
}
