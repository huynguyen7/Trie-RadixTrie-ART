package main

import (
	"encoding/json"
	"flag"
	"os"
	"runtime"
	"testing"

	"example.com/adaptive-radix-trie/adaptive-radix-trie"
	"example.com/adaptive-radix-trie/radix-trie"
	"example.com/adaptive-radix-trie/trie"
)

type Index interface {
	Insert(string) bool
	Contains(string) bool
	Delete(string) bool
	WalkPrefix(string, func(string)) bool
	Size() int
}

type Implementation struct {
	Name string
	New  func() Index
}

var implementations = []Implementation{
	{"trie", func() Index { return &trie.Trie{} }},
	{"radix", func() Index { return &radixtrie.Trie{} }},
	{"art", func() Index { return &adaptiveradixtrie.Trie{} }},
}

type Result struct {
	Implementation string  `json:"implementation"`
	Dataset        string  `json:"dataset"`
	Keys           int     `json:"keys"`
	LookupNS       float64 `json:"lookup_ns"`
	InsertNS       float64 `json:"insert_ns"`
	DeleteNS       float64 `json:"delete_ns"`
	WalkPrefixNS   float64 `json:"walkprefix_ns"`
	BytesPerKey    float64 `json:"bytes_per_key"`
}

func main() {
	testing.Init() // Provides the benchmark duration flag used by Python.
	keyCount := flag.Int("keys", 10000, "Number of keys")
	flag.Parse()
	if *keyCount < 1 {
		panic("keys must be positive")
	}
	runtime.GOMAXPROCS(1)

	var results []Result
	for _, dataset := range []string{"shared", "scattered"} {
		keys := makeKeys(*keyCount, dataset)
		for _, implementation := range implementations {
			tree, bytesPerKey := memoryUsage(implementation.New, keys)
			results = append(results, Result{
				Implementation: implementation.Name,
				Dataset:        dataset,
				Keys:           len(keys),
				LookupNS:       lookupTime(tree, keys),
				InsertNS:       insertTime(implementation.New, keys),
				DeleteNS:       deleteTime(implementation.New, keys),
				WalkPrefixNS:   walkPrefixTime(tree, len(keys)),
				BytesPerKey:    bytesPerKey,
			})
		}
	}
	if err := json.NewEncoder(os.Stdout).Encode(results); err != nil {
		panic(err)
	}
}
