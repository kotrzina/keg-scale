package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetWarehouseBeersLeft(t *testing.T) {
	assert.Equal(t, 0, GetWarehouseBeersLeft([5]int{0, 0, 0, 0, 0}), "Expected 0 beers left")
	assert.Equal(t, 20, GetWarehouseBeersLeft([5]int{1, 0, 0, 0, 0}), "Expected 20 beers left")
	assert.Equal(t, 60, GetWarehouseBeersLeft([5]int{1, 0, 1, 0, 0}), "Expected 60 beers left")
}
