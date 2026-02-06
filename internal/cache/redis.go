package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	DefaultTTL = 24 * time.Hour
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(ctx context.Context, redisURL string) (*RedisCache, error) {
	opts, err := redis.ParseURL(fmt.Sprintf("redis://%s", redisURL))
	if err != nil {
		// Try direct connection if URL parsing fails
		opts = &redis.Options{
			Addr: redisURL,
		}
	}

	client := redis.NewClient(opts)

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

// Set stores a value in cache with the default TTL
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}) error {
	return c.SetWithTTL(ctx, key, value, DefaultTTL)
}

// SetWithTTL stores a value in cache with a custom TTL
func (c *RedisCache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// Get retrieves a value from cache
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// Delete removes a key from cache
func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// DeleteByPattern removes all keys matching a pattern
func (c *RedisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

// Exists checks if a key exists in cache
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	return n > 0, err
}

// Cache key builders for D2 catalog
func D2ItemBaseKey(code string) string {
	return fmt.Sprintf("d2:item_base:%s", code)
}

func D2ItemBasesKey() string {
	return "d2:item_bases:all"
}

func D2UniqueItemKey(indexID int) string {
	return fmt.Sprintf("d2:unique:%d", indexID)
}

func D2UniqueItemsKey() string {
	return "d2:uniques:all"
}

func D2SetBonusKey(name string) string {
	return fmt.Sprintf("d2:set_bonus:%s", name)
}

func D2SetBonusesKey() string {
	return "d2:set_bonuses:all"
}

func D2SetItemKey(indexID int) string {
	return fmt.Sprintf("d2:set_item:%d", indexID)
}

func D2SetItemsKey() string {
	return "d2:set_items:all"
}

func D2RunewordKey(name string) string {
	return fmt.Sprintf("d2:runeword:%s", name)
}

func D2RunewordsKey() string {
	return "d2:runewords:all"
}

func D2RuneKey(code string) string {
	return fmt.Sprintf("d2:rune:%s", code)
}

func D2RunesKey() string {
	return "d2:runes:all"
}

func D2GemKey(code string) string {
	return fmt.Sprintf("d2:gem:%s", code)
}

func D2GemsKey() string {
	return "d2:gems:all"
}

func D2PropertyKey(code string) string {
	return fmt.Sprintf("d2:property:%s", code)
}

func D2PropertiesKey() string {
	return "d2:properties:all"
}

func D2AffixKey(name string, affixType string) string {
	return fmt.Sprintf("d2:affix:%s:%s", affixType, name)
}

func D2AffixesKey(affixType string) string {
	return fmt.Sprintf("d2:affixes:%s:all", affixType)
}

func D2ItemTypeKey(code string) string {
	return fmt.Sprintf("d2:item_type:%s", code)
}

func D2ItemTypesKey() string {
	return "d2:item_types:all"
}
