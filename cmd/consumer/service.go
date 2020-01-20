package main

import (
	"context"
	"github.com/artsv79/2services/api"
	"google.golang.org/grpc"
	"io"
	"log"
	"sync"
	"time"
)

type ConsumerService struct {
	*grpc.ClientConn

	stop             <-chan interface{}
	client           api.CacheClient
	numberOfRequests int
	requestTimeout   time.Duration
}

func NewConsumer(stop <-chan interface{}) *ConsumerService {
	c := &ConsumerService{
		stop:             stop,
		numberOfRequests: 1000,
		requestTimeout:   30 * time.Second,
	}
	return c
}

func (c *ConsumerService) Dial(address string) error {
	ch := make(chan error)
	defer close(ch)

	maxAttempts := 5
	attempts := 1

	for {
		ctx, cancelDial := context.WithTimeout(context.TODO(), 10*time.Second)
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

func (c *ConsumerService) Run() {
	defer c.Close()

	var groupSync sync.WaitGroup

	for i := 0; i < c.numberOfRequests; i++ {
		groupSync.Add(1)
		go c.requestOnce(&groupSync)
	}

	log.Printf("Waiting for all requests to complete...")
	groupSync.Wait()
	log.Printf("All %d requests finished", c.numberOfRequests)
}

func (c *ConsumerService) requestOnce(groupSync *sync.WaitGroup) {
	defer groupSync.Done()

	log.Printf("Requesting stream...")

	answerChan := make(chan string)
	ctx, cancelStream := context.WithTimeout(context.TODO(), c.requestTimeout)

	go func() {
		defer close(answerChan)

		stream, err := c.client.GetRandomDataStream(ctx, &api.RandomDataArg{})
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

iteratingAnswers:
	for {
		select {
		case <-c.stop:
			cancelStream()
			log.Printf("Stop signal received, canceling request...")
			break iteratingAnswers
		case answer, ok := <-answerChan:
			if ok {
				log.Printf("RECV: (%d), \"%.10s ...\"", len(answer), answer)
			} else {
				break iteratingAnswers
			}
		}
	}
}
