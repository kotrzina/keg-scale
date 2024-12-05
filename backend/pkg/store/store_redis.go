package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/redis/go-redis/v9"
)

const (
	Events             = "events"
	WeightKey          = "weight"
	WeightAtKey        = "weight_at"
	ActiveKegKey       = "active_keg"
	ActiveKegAtKey     = "active_keg_at"
	IsLowKey           = "is_low"
	BeersLeftKey       = "beers_left"
	BeersTotal         = "beers_total"
	WarehouseKey       = "warehouse"
	LastOkKey          = "last_ok"
	OpenAtKey          = "open_at"
	CloseAtKey         = "close_at"
	IsOpenKey          = "is_open"
	ConversationPrefix = "conversation:"
)

type RedisStore struct {
	Client *redis.Client
	ctx    context.Context
}

func NewRedisStore(ctx context.Context, c *config.Config) *RedisStore {
	return &RedisStore{
		Client: redis.NewClient(&redis.Options{
			Addr: c.RedisAddr,
			DB:   c.RedisDB,
		}),
		ctx: ctx,
	}
}

func (s *RedisStore) AddEvent(event string) error {
	err := s.Client.RPush(s.ctx, Events, event).Err()
	if err != nil {
		return err
	}

	return s.Client.LTrim(s.ctx, Events, -500, -1).Err() // keep only the last 500 events
}

func (s *RedisStore) GetEvents() ([]string, error) {
	return s.Client.LRange(s.ctx, Events, 0, -1).Result()
}

func (s *RedisStore) SetWeight(weight float64) error {
	return s.Client.Set(s.ctx, WeightKey, weight, 0).Err()
}

func (s *RedisStore) GetWeight() (float64, error) {
	return s.Client.Get(s.ctx, WeightKey).Float64()
}

func (s *RedisStore) SetWeightAt(weightAt time.Time) error {
	return s.Client.Set(s.ctx, WeightAtKey, weightAt.Format(time.RFC3339), 0).Err()
}

func (s *RedisStore) GetWeightAt() (time.Time, error) {
	res, err := s.Client.Get(s.ctx, WeightAtKey).Result()
	if err != nil {
		return time.Time{}, err
	}

	return time.Parse(time.RFC3339, res)
}

func (s *RedisStore) SetActiveKeg(keg int) error {
	return s.Client.Set(s.ctx, ActiveKegKey, keg, 0).Err()
}

func (s *RedisStore) GetActiveKeg() (int, error) {
	return s.Client.Get(s.ctx, ActiveKegKey).Int()
}

func (s *RedisStore) SetActiveKegAt(openAt time.Time) error {
	return s.Client.Set(s.ctx, ActiveKegAtKey, openAt, 0).Err()
}

func (s *RedisStore) GetActiveKegAt() (time.Time, error) {
	return s.Client.Get(s.ctx, ActiveKegAtKey).Time()
}

func (s *RedisStore) SetIsLow(isLow bool) error {
	return s.Client.Set(s.ctx, IsLowKey, isLow, 0).Err()
}

func (s *RedisStore) GetIsLow() (bool, error) {
	return s.Client.Get(s.ctx, IsLowKey).Bool()
}

func (s *RedisStore) SetBeersLeft(beersLeft int) error {
	return s.Client.Set(s.ctx, BeersLeftKey, beersLeft, 0).Err()
}

func (s *RedisStore) GetBeersLeft() (int, error) {
	return s.Client.Get(s.ctx, BeersLeftKey).Int()
}

func (s *RedisStore) SetBeersTotal(beersTotal int) error {
	return s.Client.Set(s.ctx, BeersTotal, beersTotal, 0).Err()
}

func (s *RedisStore) GetBeersTotal() (int, error) {
	return s.Client.Get(s.ctx, BeersTotal).Int()
}

func (s *RedisStore) SetWarehouse(warehouse [5]int) error {
	val := fmt.Sprintf("%d,%d,%d,%d,%d", warehouse[0], warehouse[1], warehouse[2], warehouse[3], warehouse[4])
	return s.Client.Set(s.ctx, WarehouseKey, val, 0).Err()
}

func (s *RedisStore) GetWarehouse() ([5]int, error) {
	res, err := s.Client.Get(s.ctx, "warehouse").Result()
	if err != nil {
		return [5]int{0, 0, 0, 0, 0}, err
	}

	var warehouse [5]int
	parts := strings.Split(res, ",")

	if len(parts) != 5 {
		return [5]int{0, 0, 0, 0, 0}, fmt.Errorf("invalid warehouse format in the storage")
	}

	for i, part := range parts {
		x, err := strconv.Atoi(part)
		if err != nil {
			return [5]int{0, 0, 0, 0, 0}, fmt.Errorf("invalid warehouse format in the storage")
		}

		warehouse[i] = x

	}

	return warehouse, nil
}

func (s *RedisStore) SetLastOk(lastOk time.Time) error {
	return s.Client.Set(s.ctx, LastOkKey, lastOk, 0).Err()
}

func (s *RedisStore) GetLastOk() (time.Time, error) {
	return s.Client.Get(s.ctx, LastOkKey).Time()
}

func (s *RedisStore) SetOpenAt(openAt time.Time) error {
	return s.Client.Set(s.ctx, OpenAtKey, openAt, 0).Err()
}

func (s *RedisStore) GetOpenAt() (time.Time, error) {
	return s.Client.Get(s.ctx, OpenAtKey).Time()
}

func (s *RedisStore) SetCloseAt(closeAt time.Time) error {
	return s.Client.Set(s.ctx, CloseAtKey, closeAt, 0).Err()
}

func (s *RedisStore) GetCloseAt() (time.Time, error) {
	return s.Client.Get(s.ctx, CloseAtKey).Time()
}

func (s *RedisStore) SetIsOpen(isOpen bool) error {
	return s.Client.Set(s.ctx, IsOpenKey, isOpen, 0).Err()
}

func (s *RedisStore) GetIsOpen() (bool, error) {
	return s.Client.Get(s.ctx, IsOpenKey).Bool()
}

func (s *RedisStore) AddConversationMessage(id string, msg ConservationMessage) error {
	msg.ID = id
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal conversation message: %w", err)
	}

	key := fmt.Sprintf("%s%s", ConversationPrefix, id)
	err = s.Client.RPush(s.ctx, key, data).Err()
	if err != nil {
		return fmt.Errorf("failed to add conversation message: %w", err)
	}

	return s.Client.LTrim(s.ctx, key, -500, -1).Err() // keep only the last 500 events
}

func (s *RedisStore) GetConversation(id string) ([]ConservationMessage, error) {
	key := fmt.Sprintf("%s%s", ConversationPrefix, id)
	data, err := s.Client.LRange(s.ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}

	output := make([]ConservationMessage, len(data))
	for i, d := range data {
		var msg ConservationMessage
		err := json.Unmarshal([]byte(d), &msg)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal conversation message: %w", err)
		}

		output[i] = msg
	}

	return output, nil
}

func (s *RedisStore) ResetConversation(id string) error {
	return s.Client.Del(s.ctx, fmt.Sprintf("%s%s", ConversationPrefix, id)).Err()
}
