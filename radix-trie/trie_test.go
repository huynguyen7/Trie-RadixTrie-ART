package radixtrie

import (
	"slices"
	"testing"
)

func TestEmptyTrie(t *testing.T) {
	tree := &Trie{}

	if tree.Size() != 0 {
		t.Fatalf("expected trie to have 0 element, got %d", tree.Size())
	}
	if tree.Contains("") || tree.Contains("apple") {
		t.Error("expected empty trie to contain no words.")
	}
	if tree.Delete("apple") {
		t.Error("expected deleting missing word to return false.")
	}
	matched := tree.WalkPrefix("", func(key string) {
		t.Errorf("expected no matches, got %q", key)
	})
	if matched {
		t.Error("expected empty trie to have no prefix matches.")
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
	if tree.Insert("apple") {
		t.Error("expected duplicate 'apple' not to be inserted.")
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

	for _, key := range []string{"apple", "apply", "app", "banana"} {
		tree.Insert(key)
	}

	if tree.Size() != 4 {
		t.Fatalf("expected trie to have 4 elements, got %d", tree.Size())
	}

	tests := []struct {
		name     string
		prefix   string
		expected []string
	}{
		{"stored prefix", "app", []string{"app", "apple", "apply"}},
		{"unstored prefix", "appl", []string{"apple", "apply"}},
		{"exact word", "apple", []string{"apple"}},
		{"other branch", "ban", []string{"banana"}},
		{"different inside segment", "bax", nil},
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

func TestInsertOrder(t *testing.T) {
	tests := []struct {
		name   string
		values []string
	}{
		{"short word first", []string{"app", "apple", "apply"}},
		{"long word first", []string{"apple", "app", "apply"}},
		{"branch first", []string{"apple", "apply", "app"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := &Trie{}
			for _, value := range tt.values {
				if !tree.Insert(value) {
					t.Fatalf("expected %q to be inserted.", value)
				}
			}
			for _, value := range tt.values {
				if !tree.Contains(value) {
					t.Errorf("expected %q to exist.", value)
				}
				if tree.Insert(value) {
					t.Errorf("expected duplicate %q not to be inserted.", value)
				}
			}
			if tree.Contains("appl") {
				t.Error("expected 'appl' does not exist as it just prefix.")
			}
			if tree.Size() != 3 {
				t.Errorf("expected trie to have 3 elements, got %d", tree.Size())
			}
		})
	}
}

func TestDeleteAndMerge(t *testing.T) {
	tree := &Trie{}
	tree.Insert("apple")
	tree.Insert("apply")

	if tree.Delete("app") || tree.Delete("banana") {
		t.Fatal("expected deleting missing word to return false.")
	}
	if tree.Size() != 2 {
		t.Fatalf("expected trie to have 2 elements, got %d", tree.Size())
	}
	if !tree.Delete("apply") {
		t.Fatal("expected 'apply' to be deleted.")
	}
	if tree.Contains("apply") || !tree.Contains("apple") {
		t.Error("expected only 'apple' to remain.")
	}
	if tree.Delete("apply") {
		t.Error("expected deleting 'apply' again to return false.")
	}
	if tree.Size() != 1 {
		t.Errorf("expected trie to have 1 element, got %d", tree.Size())
	}

	// After deleting 'apply', 'appl' and 'e' should merge into 'apple'.
	child := tree.head.children['a'-'a']
	if child == nil || child.prefix != "apple" || !child.terminal {
		t.Fatal("expected one terminal 'apple' segment after merging.")
	}
	var got []string
	matched := tree.WalkPrefix("app", func(key string) {
		got = append(got, key)
	})
	if !matched || !slices.Equal(got, []string{"apple"}) {
		t.Errorf("expected [apple] after merging, got %v", got)
	}
	if !tree.Delete("apple") || tree.Size() != 0 {
		t.Fatal("expected empty trie after deleting 'apple'.")
	}
	if tree.head.children['a'-'a'] != nil {
		t.Error("expected empty branch to be removed.")
	}
}

func TestDeleteAndReinsertPrefix(t *testing.T) {
	tree := &Trie{}
	for _, value := range []string{"app", "apple", "apply"} {
		tree.Insert(value)
	}
	if !tree.Delete("app") || tree.Contains("app") {
		t.Fatal("expected 'app' to be deleted.")
	}
	if !tree.Contains("apple") || !tree.Contains("apply") {
		t.Error("expected longer words to remain after deleting 'app'.")
	}
	if tree.Size() != 2 {
		t.Errorf("expected trie to have 2 elements, got %d", tree.Size())
	}
	if !tree.Insert("app") || !tree.Contains("app") {
		t.Error("expected 'app' to exist after inserting again.")
	}
	if tree.Size() != 3 {
		t.Errorf("expected trie to have 3 elements, got %d", tree.Size())
	}
	if !tree.Delete("apple") || !tree.Delete("apply") || !tree.Contains("app") {
		t.Error("expected 'app' to remain after deleting longer words.")
	}
	if tree.Size() != 1 {
		t.Errorf("expected trie to have 1 element, got %d", tree.Size())
	}
}

func TestEmptyString(t *testing.T) {
	tree := &Trie{}
	tree.Insert("apple")
	if !tree.Insert("") || !tree.Contains("") {
		t.Fatal("expected empty string to be inserted.")
	}
	if tree.Insert("") || tree.Size() != 2 {
		t.Error("expected duplicate empty string not to increase size.")
	}
	var got []string
	tree.WalkPrefix("", func(key string) {
		got = append(got, key)
	})
	if !slices.Equal(got, []string{"", "apple"}) {
		t.Errorf("expected empty string and 'apple', got %q", got)
	}
	if !tree.Delete("") || tree.Contains("") {
		t.Error("expected empty string to be deleted.")
	}
	if !tree.Contains("apple") || tree.Size() != 1 {
		t.Error("expected only 'apple' to remain.")
	}
}
