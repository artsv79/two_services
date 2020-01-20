package main

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
	"math/rand"
	"time"
)

type CacheDB interface {
	GetOrLock(url string, lockTTL time.Duration) (string, Lock, error)
	StoreAndUnlock(url string, content string, ttl time.Duration, lock Lock)
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
		ReadTimeout:        0,
		WriteTimeout:       0,
		PoolSize:           0,
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
}

type RedisCacheDB struct {
	redisClient     *redis.Client
	maxLockAttempts int
}

func (r *RedisCacheDB) GetOrLock(url string, lockTTL time.Duration) (string, Lock, error) {
	log.Printf("Getting URL: %s", url)
	//key := base64.URLEncoding.EncodeToString(url)
	key := fmt.Sprintf("(%s)", url)
	contentKey := fmt.Sprintf("(%s).Content", url)
	lockId := r.generateLockId()

	var iterationCount int
	for {
		val, err := r.redisClient.Get(contentKey).Result()
		if err == redis.Nil {
			log.Printf("Not found in db. getting lock for %s", url)
			ok, err := r.redisClient.SetNX(key, lockId, lockTTL).Result()
			if err != nil {
				return "", nil, err
			}
			if ok {
				log.Printf("Got lock %s for URL %s", lockId, url)
				return "", &RedisCacheLock{lockId: lockId, key: key, contentKey: contentKey}, nil
			}
			iterationCount++
			if iterationCount > r.maxLockAttempts {
				return "", nil, errors.New("too many attempts to acquire lock")
			}

			sleepDuration := lockTTL / 4
			log.Printf("Failed to get lock (%s), waiting %d seconds to get content or lo", lockId, int(sleepDuration.Seconds()))
			time.Sleep(sleepDuration)
		} else if err != nil {
			return "", nil, err
		} else {
			log.Printf("Found in db, len: %d", len(val))
			return val, nil, nil
		}
	}
}

func (r *RedisCacheDB) StoreAndUnlock(url string, content string, ttl time.Duration, lock Lock) {
	redisLock, ok := lock.(*RedisCacheLock)
	if ok {
		lockInDb, err := r.redisClient.Get(redisLock.key).Result()
		if err == nil && lockInDb == redisLock.lockId {
			pipe := r.redisClient.TxPipeline()
			pipe.Set(redisLock.key, "ready", ttl)
			pipe.Set(redisLock.contentKey, content, ttl)
			_, err := pipe.Exec()
			if err != nil {
				log.Printf("Error in storing (%s) content to DB: %v", url, err)
			} else {
				log.Printf("Content stored in DB for URL: %s. Len: %d", url, len(content))
			}
		} else if err == nil && lockInDb != redisLock.lockId {
			log.Printf(
				"Lock is lost. Discarding content (%s). Existing lockId: %s, my lockId: %s",
				url, lockInDb, redisLock.lockId)
		} else if err == redis.Nil {
			log.Printf(
				"Lock is lost. Discarding content (%s). Existing lockId: %s, my lockId: %s",
				url, lockInDb, redisLock.lockId)
		} else { // err != nil
			log.Printf("Error in verifying lock ownership (URL: %s, lockId: %s): %v", url, redisLock.lockId, err)
		}
	} else {
		// developer error
		panic("Wrong Lock object type")
	}
}

func (r *RedisCacheDB) generateLockId() string {
	var val int64
	for val == 0 {
		val = rand.Int63()
	}
	return fmt.Sprintf("%d", val)
}
