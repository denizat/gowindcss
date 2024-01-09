package main

import "maps"

func concatMaps[K comparable, V any](theMaps ...map[K]V) map[K]V {
	out := map[K]V{}
	for _, aMap := range theMaps {
		maps.Copy(out, aMap)
	}
	return out
}
