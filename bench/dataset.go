package main

import "math/rand"

// Keys have 16 prefix letters and a unique six-letter suffix.
func makeKeys(count int, shape string) []string {
	random := rand.New(rand.NewSource(1))
	keys := make([]string, count)

	for i := range keys {
		key := []byte("abcdefghijklmnopaaaaaa")
		if shape == "scattered" {
			for j := 0; j < 16; j++ {
				key[j] = byte('a' + random.Intn(26))
			}
		}

		// Encode i in base 26, using a-z.
		number := i
		for j := len(key) - 1; j >= 16; j-- {
			key[j] = byte('a' + number%26)
			number /= 26
		}
		keys[i] = string(key)
	}

	random.Shuffle(len(keys), func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})
	return keys
}
