package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

const App string = "2Services. Cache service"
const version string = "0.3.3"

func main() {
	log.Printf("%s. Version %s", App, version)

	redisAddress := flag.String("redis", "localhost:6379", "<host>:<port>")
	flag.Parse()
	log.Printf("DB address to connect to: %s", *redisAddress)

	sysStop := make(chan os.Signal, 1)
	defer close(sysStop)
	signal.Notify(sysStop, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGABRT, syscall.SIGINT)

	stop := make(chan interface{})
	go func() {
		<-sysStop
		close(stop)
	}()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", 9090))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	config := ReadConfig()
	config.dbAddress = *redisAddress
	err = config.Validate()
	if err != nil {
		log.Fatalf("Error while validating config: %s", err.Error())
	}

	producer := NewCache(config)

	if producer != nil {
		go func() {
			<-stop
			producer.Stop()
		}()
		log.Printf("Listening for incoming connections...")
		//a, b, c := producer.cacheDb.GetOrLock("qweqwe", 10*time.Second)
		//log.Printf("Ping: %v, %v, %v", a, b, c)

		err := producer.Serve(listener)
		if err != nil {
			log.Printf("Accepting incoming connection failed: %v", err)
		}
	}

	log.Printf("Exiting.")
}
