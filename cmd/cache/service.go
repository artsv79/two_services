package main

import (
	"github.com/artsv79/2services/api"
	"google.golang.org/grpc"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync/atomic"
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

	log.Printf("Received some")

	_ = input // input is not used for now

	ch := make(chan string, p.config.NumberOfRequests)

	var startedRequests int32
	for i := 0; i < p.config.NumberOfRequests; i++ {
		atomic.AddInt32(&startedRequests, 1)
		go func(i int) {
			url := p.config.URLs[rand.Int31n(int32(len(p.config.URLs)))]

			str, err := p.getUrlContent(url)
			if err != nil {
				log.Printf("Error getting URL: %v", err)
				ch <- err.Error()
			} else {
				ch <- str
			}

			if atomic.AddInt32(&startedRequests, -1) == 0 {
				close(ch)
			}
		}(i)
	}

	for str := range ch {
		err := server.Send(&api.RandomData{
			Str: str,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *CacheService) getUrlContent(url string) (string, error) {
	content, lock, err := p.cacheDb.GetOrLock(url, p.downloadTTL+100*time.Millisecond)
	if lock != nil {
		// really download content
		var ttl time.Duration
		content, err = p.downloadURLContent(url, p.downloadTTL)
		if err != nil {
			content = err.Error()
			// set minimal TTL
			ttl = time.Duration(p.config.MinTimeout) * time.Second
		} else {
			//set random TTL
			ttlRange := p.config.MaxTimeout - p.config.MinTimeout
			ttl = time.Duration(p.config.MinTimeout+rand.Intn(ttlRange+1)) * time.Second
		}

		p.cacheDb.StoreAndUnlock(url, content, ttl, lock)

		return content, nil
	} else if err != nil {
		return "", err
	} else {
		return content, nil
	}
}

func (p *CacheService) downloadURLContent(url string, timeout time.Duration) (string, error) {
	log.Printf("Downloading (%s)...", url)
	h := http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       timeout,
	}
	resp, err := h.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	contentLength := resp.ContentLength
	if contentLength > p.maxContentSize || contentLength < 0 {
		contentLength = p.maxContentSize
	}

	buf := make([]byte, contentLength)
	readLen, err := resp.Body.Read(buf)
	if (err == io.EOF || err == nil) && readLen >= 0 {
		log.Printf("Downloaded (%s), len: %d", url, readLen)
		return string(buf[:readLen]), nil
	} else {
		return "", err
	}
}
