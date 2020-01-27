package main

import (
	"context"
	"fmt"
	"github.com/artsv79/two_services/api"
	"google.golang.org/grpc"
	"io"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type Stat struct {
	content string
	error   error
}

type ConsumerService struct {
	*grpc.ClientConn

	stop           <-chan interface{}
	client         api.CacheClient
	requestTimeout time.Duration

	generated, cached, errors int32
}

func NewConsumer(stop <-chan interface{}) *ConsumerService {
	c := &ConsumerService{
		stop:           stop,
		requestTimeout: 30 * time.Second,
	}
	return c
}

func (c *ConsumerService) Dial(address string) error {
	ch := make(chan error)
	defer close(ch)

	maxAttempts := 5
	attempts := 1

	for {
		ctx, cancelDial := context.WithTimeout(context.TODO(), 40*time.Second)
		go func() {
			var err error

			log.Printf("Trying to connect to cache service %s ...", address)
			c.ClientConn, err = grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Printf("Error connecting to Cache service: %v", err)
			} else {
				c.client = api.NewCacheClient(c.ClientConn)
				log.Printf("Connected to %s", address)
			}
			ch <- err
		}()

		select {
		case <-c.stop:
			cancelDial()
			return ctx.Err()
		case res := <-ch:
			if res == nil || attempts >= maxAttempts {
				return res
			}
		}
		attempts++
		log.Printf("Trying once again (%d out of %d)", attempts, maxAttempts)
	}
}

func (c *ConsumerService) Run(argLowerBound, argUpperBound int64, numberOfRequests int) {
	defer c.Close()

	rand.Seed(time.Now().UnixNano())

	var groupSync sync.WaitGroup

	startTime := time.Now()

	for i := 0; i < numberOfRequests; i++ {

		var randomArg int64 = argLowerBound
		if argLowerBound < argUpperBound {
			randomArg = rand.Int63n(argUpperBound-argLowerBound) + argLowerBound
		}

		groupSync.Add(1)
		go c.requestOnce(&groupSync, randomArg)
	}

	log.Printf("Waiting for all %d requests to complete...", numberOfRequests)
	groupSync.Wait()
	finishTime := time.Now()
	log.Printf("All %d requests finished. Time spent: %v", numberOfRequests, finishTime.Sub(startTime))
	log.Printf("Results: generated: %d, cached: %d, errors: %d.   Total: %d", c.generated, c.cached, c.errors, c.generated+c.cached+c.errors)
}

func (c *ConsumerService) requestOnce(groupSync *sync.WaitGroup, requestArg int64) {
	defer groupSync.Done()

	log.Printf("Requesting stream...")

	answerChan := make(chan string)
	ctx, cancelStream := context.WithTimeout(context.TODO(), c.requestTimeout)

	go func() {
		defer close(answerChan)

		stream, err := c.client.GetRandomDataStream(ctx, &api.RandomDataArg{Arg: requestArg})
		if err != nil {
			log.Printf("Error readin from Cache service: %v", err)
		} else {
			for {
				ret, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Printf("Error reading stream: %v", err)
					break
				}
				answerChan <- ret.Str
			}
		}
	}()

	var answer string
	var contentSource string
	var requestArgStr string
	var headerReceived bool = false
	var totalReceived int = 0

iteratingAnswers:
	for {
		select {
		case <-c.stop:
			cancelStream()
			log.Printf("Stop signal received, canceling request...")
			break iteratingAnswers
		case answerPart, ok := <-answerChan:
			if ok {
				// answer will be stream of strings, which are parts of one big string.
				// this big string normally should have format {"generated"|"cached   "}:<requestArg><requestArg>...
				// where <requestArg> repeated N times (configured in cache service - StreamLen)
				// "generated:000000000000000045320000000000000000453200000000000000004532..."
				// "cached   :000000000000000045320000000000000000453200000000000000004532..."

				totalReceived += len(answerPart)
				if !headerReceived {
					// we are interested only in initial part of the answer,
					// where "generated" or "cached" word placed (10 chars) and which contains requestArg (+ 20 chars).
					answer += answerPart
				}
				if len(answer) >= 30 {
					headerReceived = true
					contentSource = answer[:10]
					requestArgStr = answer[10:30]
				}
			} else {
				log.Printf("RECEIVED: %s (%d), \"%.30s ...\"", contentSource, totalReceived, answer)
				var receivedArg int64
				n, err := fmt.Sscanf(requestArgStr, "%d", &receivedArg)
				if n == 1 {
					if receivedArg != requestArg {
						log.Printf("ERROR: received wrong request: mine = %d, received = %d.       ----------------------------------- ####", requestArg, receivedArg)
					}
				} else {
					log.Printf("ERROR: Error parsing receivedArg from content: %v", err)
				}
				break iteratingAnswers
			}
		}
	}

	switch contentSource {
	case "generated:":
		atomic.AddInt32(&c.generated, 1)
	case "cached   :":
		atomic.AddInt32(&c.cached, 1)
	default:
		atomic.AddInt32(&c.errors, 1)
	}
}
