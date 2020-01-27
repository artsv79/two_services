package main

import (
	"fmt"
	"github.com/artsv79/two_services/api"
	"google.golang.org/grpc"
	"io"
	"log"
	"math/rand"
	"net/http"
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
	str, err := p.getUrlContent(url)
	if err != nil {
		log.Printf("Error getting URL: %v", err)
		str = fmt.Sprintf("%v", err)
	}

	err = server.Send(&api.RandomData{
		Str: str,
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *CacheService) getUrlContent(url string) (string, error) {
	content, lock, err := p.cacheDb.GetOrLock(url, p.downloadTTL+100*time.Millisecond)
	if lock != nil {
		// generate content
		var ttl time.Duration
		//set random TTL
		ttlRange := p.config.MaxTimeout - p.config.MinTimeout
		ttl = time.Duration(p.config.MinTimeout+rand.Intn(ttlRange+1)) * time.Second

		strBuf := make([]byte, 0, 20*p.config.StreamLen)
		for i := 0; i < p.config.StreamLen; i++ {
			strBuf = append(strBuf, url...)
		}
		content = string(strBuf)

		p.cacheDb.StoreAndUnlock(url, content, ttl, lock)

		return "generated:" + content, nil
	} else if err != nil {
		return "", err
	} else {
		// return content from cache
		return "cached   :" + content, nil
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
