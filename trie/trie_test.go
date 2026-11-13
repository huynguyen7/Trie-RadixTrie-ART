package trie

import (
	"slices"
	"testing"
)

func TestEmptyTrie(t *testing.T) {
	tree := &Trie{}

	if tree.Size() != 0 {
		t.Fatalf("expected trie to have 0 element, got %d", tree.Size())
	}
}

func TestInsertAndContains(t *testing.T) {
	tree := &Trie{}

	added := tree.Insert("apple")

	if !added {
		t.Fatal("expected 'apple' to be inserted.")
	}

	if !tree.Contains("apple") {
		t.Error("expected 'apple' to exist.")
	}

	if tree.Contains("banana") {
		t.Error("expected 'banana' does not exist.")
	}

	if tree.Contains("app") {
		t.Error("expected 'app' does not exist as it just prefix.")
	}

	if tree.Size() != 1 {
		t.Errorf("expected trie to have 1 element, got %d", tree.Size())
	}
}

func TestDelete(t *testing.T) {
	tree := &Trie{}

	added := tree.Insert("apple")
	if !added {
		t.Fatal("expected 'apple' to be inserted.")
	}

	deleted := tree.Delete("apple")
	if !deleted {
		t.Fatal("expected 'apple' to be deleted.")
	}

	if tree.Size() != 0 {
		t.Errorf("expected trie to have 0 element, got %d", tree.Size())
	}
}

func TestWalkPrefix(t *testing.T) {
	tree := &Trie{}

	for _, key := range []string {"apple", "apply", "app", "banana"} {
		tree.Insert(key)
	}
	
	if tree.Size() != 4 {
		t.Fatalf("expected trie to have 4 elements, got %d", tree.Size())
	}

	tests := []struct {
		name   string
		prefix string
		expected []string
	}{
		{"stored prefix", "app", []string{"app", "apple", "apply"}},
		{"unstored prefix", "appl", []string{"apple", "apply"}},
		{"exact word", "apple", []string{"apple"}},
		{"other branch", "ban", []string{"banana"}},
		{"missing prefix", "orange", nil},
		{"longer than key", "apples", nil},
		{"empty prefix", "", []string{"app", "apple", "apply", "banana"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string

			matched := tree.WalkPrefix(tt.prefix, func(key string) {
				got = append(got, key)
			})
			slices.Sort(got)

			if !slices.Equal(got, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}

			expectedMatched := len(tt.expected) > 0
			if matched != expectedMatched {
				t.Errorf("expected matched=%v, got %v", expectedMatched, matched)
			}
		})
	}
}
