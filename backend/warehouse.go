package main

import "fmt"

// GetWarehouseIndex returns the index of the keg in the warehouse array
func GetWarehouseIndex(keg int) (int, error) {
	switch keg {
	case 10:
		return 0, nil
	case 15:
		return 1, nil
	case 20:
		return 2, nil
	case 30:
		return 3, nil
	case 50:
		return 4, nil
	}

	return 0, fmt.Errorf("invalid keg") // should never happen
}
