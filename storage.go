package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/mattheath/base62"
	"time"
)

type Storage interface {
	Shorten(url string, exp int64) (string, error)
	ShortlinkInfo(eid string) (interface{}, error)
	Unshorten(eid string) (string, error)
}

const (
	UrlIDkey           = "next.url.id"
	ShortlinkKey       = "shortlink:%s:url"
	UrlHashKey         = "urlhash:%s:shortlinkzL"
	ShortlinkDetailKey = "shortlink:%s:detail"
)

type RedisClient struct {
	Client *redis.Client
}

func (r *RedisClient) Shorten(url string, exp int64) (string, error) {
	//convert url to sha1 hash
	h := toSha1(url)

	result, err := r.Client.Get(fmt.Sprintf(UrlHashKey, h)).Result()
	if err == redis.Nil {
		// not existed
	} else if err != nil {
		return "", err
	} else {
		if result == "{}" {
			// expiration, return {}， nothing to do
		} else {
			return result, nil
		}
	}

	// increase global counter
	err = r.Client.Incr(UrlIDkey).Err()
	if err != nil {
		return "", err
	}

	//encode global counter to base62
	id, err := r.Client.Get(UrlIDkey).Int64()
	if err != nil {
		return "", err
	}
	eid := base62.EncodeInt64(id)

	//store the url against the eid
	err = r.Client.Set(fmt.Sprintf(ShortlinkKey, eid), url, time.Second*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	//store the url against the hash of it
	err = r.Client.Set(fmt.Sprintf(UrlHashKey, h), eid, time.Second*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	detail, err := json.Marshal(&URLDetail{
		URL:                 url,
		CreatedAt:           time.Now().String(),
		ExpirationInMinutes: time.Minute * time.Duration(exp),
	})
	if err != nil {
		return "", err
	}

	//store the url detail against the eid
	err = r.Client.Set(fmt.Sprintf(ShortlinkDetailKey, eid), detail, time.Second*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}
	return eid, nil
}

func (r *RedisClient) ShortlinkInfo(eid string) (interface{}, error) {
	panic("implement me")
}

func (r *RedisClient) Unshorten(eid string) (string, error) {
	panic("implement me")
}

type URLDetail struct {
	URL                 string        `json:"url"`
	CreatedAt           string        `json:"created_at"`
	ExpirationInMinutes time.Duration `json:"expiration_in_minutes"`
}

func NewRedisClient(addr, pwd string, db int) *RedisClient {
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       db,
	})
	if _, err := c.Ping().Result(); err != nil {
		panic(err)
	}
	return &RedisClient{Client: c}
}

func toSha1(url string) string {

}
