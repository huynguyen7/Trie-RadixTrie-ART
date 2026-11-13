package main

/**
	Skeleton for trie, radix trie and adaptive radix trie
*/

// Only accept alphabet a-z lowercase for now..
type Index interface {
	Insert(value string) bool // Insert a string, return true if not exist.
	Contains(value string) bool
	Delete(value string) bool // Return true if exist
	WalkPrefix(prefix string, visit func(key string)) bool // Return true if prefix exists
	Size() int
}
