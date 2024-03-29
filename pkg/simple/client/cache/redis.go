
package cache

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"k8s.io/klog"
)

type Client struct {
	client *redis.Client
}

func NewRedisClient(option *Options, stopCh <-chan struct{}) (Interface, error) {
	var r Client

	redisOptions := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", option.Host, option.Port),
		Password: option.Password,
		DB:       option.DB,
	}

	if stopCh == nil {
		klog.Fatalf("no stop channel passed, redis connections will leak.")
	}

	r.client = redis.NewClient(redisOptions)

	if err := r.client.Ping().Err(); err != nil {
		r.client.Close()
		return nil, err
	}

	// close redis in case of connection leak
	if stopCh != nil {
		go func() {
			<-stopCh
			if err := r.client.Close(); err != nil {
				klog.Error(err)
			}
		}()
	}

	return &r, nil
}

func (r *Client) Get(key string) (string, error) {
	return r.client.Get(key).Result()
}

func (r *Client) Keys(pattern string) ([]string, error) {
	return r.client.Keys(pattern).Result()
}

func (r *Client) Set(key string, value string, duration time.Duration) error {
	return r.client.Set(key, value, duration).Err()
}

func (r *Client) Del(keys ...string) error {
	return r.client.Del(keys...).Err()
}

func (r *Client) Exists(keys ...string) (bool, error) {
	existedKeys, err := r.client.Exists(keys...).Result()
	if err != nil {
		return false, err
	}

	return len(keys) == int(existedKeys), nil
}

func (r *Client) Expire(key string, duration time.Duration) error {
	return r.client.Expire(key, duration).Err()
}
