package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const App string = "2Services. Consumer service"
const version string = "0.2.4"

func main() {
	log.Printf("%s. Version %s", App, version)

	cacheServiceAddress := flag.String("cache", "localhost:9090", "<host>:<port>")
	argMin := flag.Int64("argMin", 0, "lower bound for randomly generated arg, used for requesting cache service. Int64")
	argMax := flag.Int64("argMax", 1000, "upper bound for randomly generated arg, used for requesting cache service. Int64")
	requestsNumber := flag.Int("requestsNum", 1000, "number of parallel requests to make to cache service")
	flag.Parse()
	if *argMin > *argMax {
		log.Fatalf("Invalid arguments: argMin should be less than argMax. Provided: argMin=%d, argMax=%d", *argMin, *argMax)
	}
	if *requestsNumber <= 0 {
		log.Fatalf("Invalid argument requestsNum. Should be >= 1")
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

	consumer.Run(*argMin, *argMax, *requestsNumber)

	log.Printf("Exiting.")
}
