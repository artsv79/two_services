package main

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
	"math/rand"
	"runtime/debug"
	"time"
)

type CacheDB interface {
	GetOrLock(url string, lockTTL time.Duration) (chan string, Lock, error)
	// don't close this writer chan. Use Unlock instead fo mark the and of stream
	WriterForLocked(lock Lock) (chan string, error)
	Unlock(lock Lock, ttl time.Duration) error
}

type Lock interface {
}

func NewRedisCacheDB(address string) CacheDB {
	p := &RedisCacheDB{
		maxLockAttempts: 100,
	}
	p.redisClient = redis.NewClient(&redis.Options{
		Network:            "",
		Addr:               address,
		Dialer:             nil,
		OnConnect:          nil,
		Password:           "",
		DB:                 0,
		MaxRetries:         0,
		MinRetryBackoff:    0,
		MaxRetryBackoff:    0,
		DialTimeout:        0,
		ReadTimeout:        40 * time.Second,
		WriteTimeout:       0,
		PoolSize:           1000,
		MinIdleConns:       0,
		MaxConnAge:         0,
		PoolTimeout:        0,
		IdleTimeout:        0,
		IdleCheckFrequency: 0,
		TLSConfig:          nil,
	})
	return p
}

type RedisCacheLock struct {
	lockId     string
	key        string
	contentKey string
	writeChan  chan string
}

type RedisCacheDB struct {
	redisClient     *redis.Client
	maxLockAttempts int
}

func (r *RedisCacheDB) GetOrLock(url string, lockTTL time.Duration) (chan string, Lock, error) {
	log.Printf("Getting URL: %s", url)
	//key := base64.URLEncoding.EncodeToString(url)
	key := fmt.Sprintf("(%s)", url)
	//contentKey := fmt.Sprintf("(%s).Content", url)
	lockId := r.generateLockId()

	for {
		redisStream, err := r.redisClient.XRead(&redis.XReadArgs{
			Streams: []string{key, "0-0"},
			Count:   0,
			Block:   -1,
		}).Result()
		if err == redis.Nil {

			firstId := 1
			// Empty. No stream for this key
			id, err := r.redisClient.XAdd(&redis.XAddArgs{
				Stream:       key,
				MaxLenApprox: 100000,
				ID:           fmt.Sprintf("0-%d", firstId),
				Values:       map[string]interface{}{"initial": "true"},
			}).Result()
			if err != nil {
				log.Printf("Failed to create stream for %s: %v", key, err)
				//TODO asv: implelent
				//return nil, nil, err
			} else {
				// Good - we got the lock for this URL
				log.Printf("Got lock %s (id=%s) for URL %s", lockId, id, key)
				//TODO asv: handle timouts
				ok, err := r.redisClient.Expire(key, lockTTL/4).Result()
				log.Printf("Setting TTL=%v for %s: %t (%v)", lockTTL/4, key, ok, err)

				wChan := make(chan string, 1000)
				go func() {
					recentId := firstId + 1 // next id for redis "XADD", will have form of "0-2", "0-3", and so forth
					for item := range wChan {
						_, err := r.redisClient.XAdd(&redis.XAddArgs{
							Stream: key,
							ID:     fmt.Sprintf("0-%d", recentId),
							Values: map[string]interface{}{"msg": item},
						}).Result()
						if err != nil {
							log.Printf("ERROR adding item to db for %s (lock: %s): %v", key, lockId, err)
						}
						recentId++
					}
					id, err = r.redisClient.XAdd(&redis.XAddArgs{
						Stream: key,
						ID:     fmt.Sprintf("0-%d", recentId),
						Values: map[string]interface{}{"finalizer": "true"},
					}).Result()
					if err != nil {
						log.Printf("ERROR adding finalizer to db for %s (lock: %s): %v", key, lockId, err)
					}
				}()
				return nil, &RedisCacheLock{lockId: lockId, key: key, writeChan: wChan}, nil
			}
		} else if err != nil {
			// some error - just return it.
			return nil, nil, err
		} else {
			// We got values.
			log.Printf("Got values for %s. Start reading...", key)
			readTimeout := lockTTL / 4

			readerChan := make(chan string)

			go func() {
				defer close(readerChan)
				defer log.Printf("Closing channel for reading %s", key)
				processMessages := func(_redisStreams []redis.XStream, previousLatestId string) (lastReceived bool, latestId string) {
					latestId = previousLatestId
					lastReceived = false

					if len(_redisStreams) > 0 {
						for _, message := range _redisStreams[0].Messages {
							latestId = message.ID
							if message.Values["initial"] == "true" {
								//skip
							} else if message.Values["finalizer"] == "true" {
								// skip, but return "end" flag
								lastReceived = true
								break
							} else if msg := message.Values["msg"]; msg != nil {
								stringValue := fmt.Sprintf("%v", msg)
								readerChan <- stringValue
							}
						}
					}
					return lastReceived, latestId
				}
				iterationCount := 0
				latestId := "0-0"
				for {
					isLast, latestId := processMessages(redisStream, latestId)
					if isLast {
						log.Printf("Read all values fro %s", key)
						break
					}
					redisStream, err = r.redisClient.XRead(&redis.XReadArgs{
						Streams: []string{key, latestId},
						Count:   0,
						Block:   readTimeout,
					}).Result()
					if err == redis.Nil {
						iterationCount++
						if iterationCount >= 4 {
							log.Printf("Timeout on waiting stream. Have read up to %s", latestId)
							break
						}
					} else if err != nil {
						log.Printf("Failed to read stream %s: %v", key, err)
						break
					} else {
						iterationCount = 0
					}
				}
			}()

			return readerChan, nil, nil
		}
	}
}

func (r *RedisCacheDB) WriterForLocked(lock Lock) (chan string, error) {
	if lock != nil {
		switch redisLock := lock.(type) {
		case *RedisCacheLock:
			return redisLock.writeChan, nil
		default:
			return nil, errors.New("dev ERROR. Wrong lock object type")
		}
	} else {
		return nil, errors.New("dev ERROR. Lock object is nil")
	}
}

func (r *RedisCacheDB) Unlock(lock Lock, ttl time.Duration) error {
	if lock != nil {
		switch redisLock := lock.(type) {
		case *RedisCacheLock:
			ok, err := r.redisClient.Expire(redisLock.key, ttl).Result()
			log.Printf("Setting TTL=%v for %s: %t (%v)", ttl, redisLock.key, ok, err)
			// redisLock.writeChan possible could be closed at this time.
			defer func() {
				if panicMessage := recover(); panicMessage != nil {
					stack := debug.Stack()
					log.Printf("RECOVERED FROM UNHANDLED PANIC: %v\nSTACK: %s", panicMessage, stack)
				}
			}()
			close(redisLock.writeChan)
			return nil
		default:
			return errors.New("dev ERROR. Wrong lock object type")
		}
	} else {
		return errors.New("dev ERROR. Lock object is nil")
	}
}

func (r *RedisCacheDB) generateLockId() string {
	var val int64
	for val == 0 {
		val = rand.Int63()
	}
	return fmt.Sprintf("%d", val)
}
