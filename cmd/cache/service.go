package main

import (
	"fmt"
	"github.com/artsv79/two_services/api"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"time"
)

//"github.com/go-redis/redis/v7"

type CacheService struct {
	*grpc.Server
	api.UnimplementedCacheServer
	config         Config
	cacheDb        CacheDB
	maxContentSize int64
	downloadTTL    time.Duration
}

func NewCache(config Config) *CacheService {
	p := &CacheService{
		config:         config,
		maxContentSize: 1024 * 1024,
		downloadTTL:    15 * time.Second,
	}

	p.cacheDb = NewRedisCacheDB(config.dbAddress)

	p.Server = grpc.NewServer()
	api.RegisterCacheServer(p.Server, p)

	return p
}

func (p *CacheService) GetRandomDataStream(input *api.RandomDataArg, server api.Cache_GetRandomDataStreamServer) error {

	log.Printf("Received request for arg: %d", input.Arg)

	url := fmt.Sprintf("%020d", input.Arg)
	ch, err := p.getUrlContent(url)
	if err != nil {
		log.Printf("Error getting URL: %v", err)
		str := fmt.Sprintf("%v", err)
		_ = server.Send(&api.RandomData{
			Str: str,
		})
		return nil
	}

	for str := range ch {
		err = server.Send(&api.RandomData{
			Str: str,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *CacheService) getUrlContent(url string) (chan string, error) {
	contentFromDb, lock, err := p.cacheDb.GetOrLock(url, p.downloadTTL+100*time.Millisecond)
	if lock != nil {
		// generate content
		var ttl time.Duration
		//set random TTL
		ttlRange := p.config.MaxTimeout - p.config.MinTimeout
		ttl = time.Duration(p.config.MinTimeout+rand.Intn(ttlRange+1)) * time.Second

		contentChan := make(chan string)

		go func() {
			defer close(contentChan)
			defer p.cacheDb.Unlock(lock, ttl)
			contentChan <- "generated:"
			dbChan, err := p.cacheDb.WriterForLocked(lock)
			if err != nil {
				log.Printf("ERROR: %v", err)
				return
			}
			for i := 0; i < p.config.StreamLen; i++ {
				generated := url
				contentChan <- generated
				dbChan <- generated
			}
		}()

		return contentChan, nil

	} else if err != nil {
		return nil, err
	} else {
		// return content from cache
		contentChan := make(chan string)
		go func() {
			contentChan <- "cached   :"
			for s := range contentFromDb {
				contentChan <- s
			}
			close(contentChan)
		}()
		return contentChan, nil
	}
}
