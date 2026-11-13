package main

import (
	"runtime"
	"testing"
)

func buildTree(newTree func() Index, keys []string) Index {
	tree := newTree()
	for _, key := range keys {
		tree.Insert(key)
	}
	return tree
}

// Input strings already exist. Measure the additional live heap for the tree.
func memoryUsage(newTree func() Index, keys []string) (Index, float64) {
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	tree := buildTree(newTree, keys)
	runtime.GC()
	runtime.ReadMemStats(&after)
	runtime.KeepAlive(tree)
	runtime.KeepAlive(keys)

	// Check answers before timing the operations.
	if tree.Size() != len(keys) {
		panic("incorrect key count")
	}
	for _, key := range keys {
		if !tree.Contains(key) || tree.Contains(key+"z") {
			panic("incorrect lookup result")
		}
	}

	bytes := int64(after.HeapAlloc) - int64(before.HeapAlloc)
	return tree, float64(bytes) / float64(len(keys))
}

func lookupTime(tree Index, keys []string) float64 {
	result := testing.Benchmark(func(b *testing.B) {
		position := 0
		var found bool
		for b.Loop() {
			found = tree.Contains(keys[position])
			position++
			if position == len(keys) {
				position = 0
			}
		}
		runtime.KeepAlive(found)
	})
	return float64(result.T.Nanoseconds()) / float64(result.N)
}

// Each batch inserts distinct keys into a fresh empty tree.
func insertTime(newTree func() Index, keys []string) float64 {
	result := testing.Benchmark(func(b *testing.B) {
		for b.Loop() {
			tree := buildTree(newTree, keys)
			runtime.KeepAlive(tree)
		}
	})
	operations := float64(result.N) * float64(len(keys))
	return float64(result.T.Nanoseconds()) / operations
}

// Refill the tree outside the timer so every deletion removes an existing key.
func deleteTime(newTree func() Index, keys []string) float64 {
	result := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			tree := buildTree(newTree, keys)
			b.StartTimer()
			for _, key := range keys {
				tree.Delete(key)
			}
			b.StopTimer()
			if tree.Size() != 0 {
				panic("deletion left keys in the tree")
			}
		}
	})
	operations := float64(result.N) * float64(len(keys))
	return float64(result.T.Nanoseconds()) / operations
}

// An empty prefix visits every key, giving both datasets the same match count.
func walkPrefixTime(tree Index, keyCount int) float64 {
	count := 0
	visit := func(string) { count++ }
	result := testing.Benchmark(func(b *testing.B) {
		for b.Loop() {
			count = 0
			tree.WalkPrefix("", visit)
		}
		if count != keyCount {
			panic("incorrect prefix match count")
		}
	})
	return float64(result.T.Nanoseconds()) / float64(result.N)
}
