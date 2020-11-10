package main

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/QXQZX/short-link/base62"
	"github.com/go-redis/redis"
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
	UrlHashKey         = "urlhash:%s:shortlink"
	ShortlinkDetailKey = "shortlink:%s:detail"
)

type URLDetail struct {
	URL                 string        `json:"url"`
	CreatedAt           string        `json:"created_at"`
	ExpirationInMinutes time.Duration `json:"expiration_in_minutes"`
}

type RedisClient struct {
	Client *redis.Client
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

func (r *RedisClient) Shorten(url string, exp int64) (string, error) {
	//convert url to sha1 hash
	h := toSha1(url)

	// select shortlink by hash(url)
	result, err := r.Client.Get(fmt.Sprintf(UrlHashKey, h)).Result()
	if err == redis.Nil {
		// not existed
	} else if err != nil {
		return "", err
	} else {
		if result == "{}" {
			// expiration, return {}, nothing to do
		} else {
			return result, nil
		}
	}

	// increase global counter
	err = r.Client.Incr(UrlIDkey).Err()
	if err != nil {
		return "", err
	}

	// get then encode global counter to base62
	id, err := r.Client.Get(UrlIDkey).Int64()
	if err != nil {
		return "", err
	}
	eid := base62.EncodeInt64(id)

	// store the url against the eid
	err = r.Client.Set(fmt.Sprintf(ShortlinkKey, eid), url, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	// store the shortlink against the hash(url)
	err = r.Client.Set(fmt.Sprintf(UrlHashKey, h), eid, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	detail, err := json.Marshal(&URLDetail{
		URL:                 url,
		CreatedAt:           time.Now().String(),
		ExpirationInMinutes: time.Duration(exp),
	})
	if err != nil {
		return "", err
	}

	// store the url detail against the eid
	err = r.Client.Set(fmt.Sprintf(ShortlinkDetailKey, eid), detail, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}
	return eid, nil
}

func (r *RedisClient) ShortlinkInfo(eid string) (interface{}, error) {
	result, err := r.Client.Get(fmt.Sprintf(ShortlinkDetailKey, eid)).Result()
	if err == redis.Nil {
		return "", StatusError{Code: 404, Err: errors.New("Unknown short URL")}
	} else if err != nil {
		return "", err
	} else {
		var URLdetail = URLDetail{}
		err := json.Unmarshal([]byte(result), &URLdetail)
		if err != nil {
			return URLDetail{}, err
		}
		return URLdetail, nil
	}
}

func (r *RedisClient) Unshorten(eid string) (string, error) {
	result, err := r.Client.Get(fmt.Sprintf(ShortlinkKey, eid)).Result()
	if err == redis.Nil {
		return "", StatusError{Code: 404, Err: err}
	} else if err != nil {
		return "", err
	} else {
		return result, nil
	}
}

func toSha1(url string) string {
	hash := sha1.New()
	hash.Write([]byte(url))
	bs := hash.Sum(nil)
	//SHA1 值经常以 16 进制输出，例如在 git commit 中。使用%x 来将散列结果格式化为 16 进制字符串。
	return fmt.Sprintf("%x\n", bs)
}
