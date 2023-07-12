package service

func copyMap(in map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range in {
		out[k] = v
	}
	return out
}

func updateMap(target, updates map[string]string) {
	for k, v := range updates {
		target[k] = v
	}
}
