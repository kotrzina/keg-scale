package main

import "math"

func calcBeersLeft(keg int, weight float64) int {
	var kegWeight float64 = 0 // ignore keg if not found
	switch keg {
	case 10:
		kegWeight = 6
	case 15:
		kegWeight = 7
	case 20:
		kegWeight = 7.25 // guess
	case 30:
		kegWeight = 7.5
	case 50:
		kegWeight = 10.1 // guess
	}

	if kegWeight > weight/1000 {
		return 0
	}

	return int(math.Floor((weight/1000 - kegWeight) * 2))
}
