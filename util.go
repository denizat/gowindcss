package main

import (
	"maps"
	"strconv"
	"strings"
)

func concatMaps[K comparable, V any](theMaps ...map[K]V) map[K]V {
	out := map[K]V{}
	for _, aMap := range theMaps {
		maps.Copy(out, aMap)
	}
	return out
}

func parseFraction(s string) (float64, error) {
	parts := strings.Split(s, "/")
	numerator, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return numerator, err
	}
	denominator, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return denominator, err
	}
	return numerator / denominator, nil
}

type NilByteWriter struct{}

func (n NilByteWriter) WriteByte(_ byte) error {
	return nil
}
