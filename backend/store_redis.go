package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"strings"
)

const (
	ActiveKegKey       = "active_keg"
	MeasurementListKey = "measurements"
	IsLowKey           = "is_low"
	BeersLeftKey       = "beers_left"
	WarehouseKey       = "warehouse"
)

type RedisStore struct {
	Client *redis.Client
}

func NewRedisStore(config *Config) *RedisStore {
	return &RedisStore{
		Client: redis.NewClient(&redis.Options{
			Addr: config.RedisAddr,
			DB:   config.RedisDB,
		}),
	}
}

func (s *RedisStore) SetActiveKeg(keg int) error {
	return s.Client.Set(context.Background(), ActiveKegKey, keg, 0).Err()
}

func (s *RedisStore) GetActiveKeg() (int, error) {
	return s.Client.Get(context.Background(), ActiveKegKey).Int()
}

func (s *RedisStore) AddMeasurement(m Measurement) error {
	ctx := context.Background()

	var buf bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&buf) // Will write to network.
	if err := enc.Encode(m); err != nil {
		return err
	}

	if err := s.Client.RPush(ctx, MeasurementListKey, buf.String()).Err(); err != nil {
		return err
	}

	return s.Client.LTrim(ctx, MeasurementListKey, -1000, -1).Err() // keep last 1000 items
}

func (s *RedisStore) GetMeasurements() ([]Measurement, error) {
	ctx := context.Background()

	vals, err := s.Client.LRange(ctx, MeasurementListKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var measurements []Measurement
	for _, val := range vals {
		var m Measurement
		dec := gob.NewDecoder(bytes.NewBufferString(val))
		if err := dec.Decode(&m); err != nil {
			return nil, err
		}
		measurements = append(measurements, m)
	}

	return measurements, nil
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
