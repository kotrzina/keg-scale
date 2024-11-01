package store

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"strings"
	"time"

	"github.com/kotrzina/keg-scale/pkg/config"
)

const (
	WeightKey    = "weight"
	WeightAtKey  = "weight_at"
	ActiveKegKey = "active_keg"
	IsLowKey     = "is_low"
	BeersLeftKey = "beers_left"
	WarehouseKey = "warehouse"
	LastOkKey    = "last_ok"
	OpenAtKey    = "open_at"
	CloseAtKey   = "close_at"
	IsOpenKey    = "is_open"
)

type RedisStore struct {
	Client *redis.Client
}

func NewRedisStore(config *config.Config) *RedisStore {
	return &RedisStore{
		Client: redis.NewClient(&redis.Options{
			Addr: config.RedisAddr,
			DB:   config.RedisDB,
		}),
	}
}

func (s *RedisStore) SetWeight(weight float64) error {
	return s.Client.Set(context.Background(), WeightKey, weight, 0).Err()
}

func (s *RedisStore) GetWeight() (float64, error) {
	return s.Client.Get(context.Background(), WeightKey).Float64()
}

func (s *RedisStore) SetWeightAt(weightAt time.Time) error {
	return s.Client.Set(context.Background(), WeightAtKey, weightAt.Format(time.RFC3339), 0).Err()
}

func (s *RedisStore) GetWeightAt() (time.Time, error) {
	res, err := s.Client.Get(context.Background(), WeightAtKey).Result()
	if err != nil {
		return time.Time{}, err
	}

	return time.Parse(time.RFC3339, res)
}

func (s *RedisStore) SetActiveKeg(keg int) error {
	return s.Client.Set(context.Background(), ActiveKegKey, keg, 0).Err()
}

func (s *RedisStore) GetActiveKeg() (int, error) {
	return s.Client.Get(context.Background(), ActiveKegKey).Int()
}

func (s *RedisStore) SetIsLow(isLow bool) error {
	return s.Client.Set(context.Background(), IsLowKey, isLow, 0).Err()
}

func (s *RedisStore) GetIsLow() (bool, error) {
	return s.Client.Get(context.Background(), IsLowKey).Bool()
}

func (s *RedisStore) SetBeersLeft(beersLeft int) error {
	return s.Client.Set(context.Background(), BeersLeftKey, beersLeft, 0).Err()
}

func (s *RedisStore) GetBeersLeft() (int, error) {
	return s.Client.Get(context.Background(), BeersLeftKey).Int()
}

func (s *RedisStore) SetWarehouse(warehouse [5]int) error {
	val := fmt.Sprintf("%d,%d,%d,%d,%d", warehouse[0], warehouse[1], warehouse[2], warehouse[3], warehouse[4])
	return s.Client.Set(context.Background(), WarehouseKey, val, 0).Err()
}

func (s *RedisStore) GetWarehouse() ([5]int, error) {
	res, err := s.Client.Get(context.Background(), "warehouse").Result()
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
	return s.Client.Set(context.Background(), LastOkKey, lastOk, 0).Err()
}
func (s *RedisStore) GetLastOk() (time.Time, error) {
	return s.Client.Get(context.Background(), LastOkKey).Time()
}
func (s *RedisStore) SetOpenAt(openAt time.Time) error {
	return s.Client.Set(context.Background(), OpenAtKey, openAt, 0).Err()
}
func (s *RedisStore) GetOpenAt() (time.Time, error) {
	return s.Client.Get(context.Background(), OpenAtKey).Time()
}
func (s *RedisStore) SetCloseAt(closeAt time.Time) error {
	return s.Client.Set(context.Background(), CloseAtKey, closeAt, 0).Err()
}
func (s *RedisStore) GetCloseAt() (time.Time, error) {
	return s.Client.Get(context.Background(), CloseAtKey).Time()
}
func (s *RedisStore) SetIsOpen(isOpen bool) error {
	return s.Client.Set(context.Background(), IsOpenKey, isOpen, 0).Err()
}
func (s *RedisStore) GetIsOpen() (bool, error) {
	return s.Client.Get(context.Background(), IsOpenKey).Bool()
}
