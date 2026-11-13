package adaptiveradixtrie

import (
	"math/rand"
	"slices"
	"strings"
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
	child := tree.head.findChild('a')
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
	if tree.head.findChild('a') != nil {
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

// Used only by tests to check that storage actually grows and shrinks.
func childCapacity(current *trieNode) int {
	switch current.children.(type) {
	case *node4:
		return 4
	case *node16:
		return 16
	case *node48:
		return 48
	case *node256:
		return 256
	}
	return 0
}

func TestNodeGrowthAndShrink(t *testing.T) {
	tree := &Trie{}
	growth := []struct {
		count    int
		capacity int
	}{
		{1, 4}, {4, 4}, {5, 16}, {16, 16},
		{17, 48}, {48, 48}, {49, 256}, {256, 256},
	}
	for _, tt := range growth {
		for tree.Size() < tt.count {
			// Insert in reverse byte order to also check sorting.
			value := string([]byte{byte(255 - tree.Size())})
			if !tree.Insert(value) {
				t.Fatalf("expected %q to be inserted.", value)
			}
		}
		if childCapacity(&tree.head) != tt.capacity || tree.head.count != tt.count {
			t.Fatalf("expected %d children in node%d, got %d in %T",
				tt.count, tt.capacity, tree.head.count, tree.head.children)
		}
		var got []string
		tree.WalkPrefix("", func(key string) { got = append(got, key) })
		var expected []string
		for i := 256 - tt.count; i < 256; i++ {
			value := string([]byte{byte(i)})
			expected = append(expected, value)
			if !tree.Contains(value) || tree.Insert(value) {
				t.Fatalf("expected %q to exist without being inserted again.", value)
			}
		}
		if !slices.Equal(got, expected) || tree.Size() != tt.count {
			t.Fatalf("expected %d sorted words, got %q", tt.count, got)
		}
	}

	shrink := []struct {
		count    int
		capacity int
	}{
		{49, 256}, {48, 48}, {17, 48}, {16, 16},
		{5, 16}, {4, 4}, {1, 4}, {0, 0},
	}
	for _, tt := range shrink {
		for tree.Size() > tt.count {
			value := string([]byte{byte(tree.Size() - 1)})
			if !tree.Delete(value) || tree.Contains(value) {
				t.Fatalf("expected %q to be deleted.", value)
			}
		}
		if childCapacity(&tree.head) != tt.capacity || tree.head.count != tt.count {
			t.Fatalf("expected %d children in node%d, got %d in %T",
				tt.count, tt.capacity, tree.head.count, tree.head.children)
		}
		var got []string
		tree.WalkPrefix("", func(key string) { got = append(got, key) })
		var expected []string
		for i := 0; i < tt.count; i++ {
			value := string([]byte{byte(i)})
			expected = append(expected, value)
			if !tree.Contains(value) {
				t.Fatalf("expected %q to remain after shrinking.", value)
			}
		}
		if !slices.Equal(got, expected) {
			t.Fatalf("expected remaining words %q, got %q", expected, got)
		}
	}
	if !tree.Insert("apple") || !tree.Contains("apple") {
		t.Error("expected insertion to work after deleting all words.")
	}
}

func TestNode48Reuse(t *testing.T) {
	tree := &Trie{}
	for i := 0; i < 48; i++ {
		tree.Insert(string([]byte{byte(i)}))
	}
	tree.Delete(string([]byte{17}))
	tree.Insert(string([]byte{250}))

	if tree.Size() != 48 || childCapacity(&tree.head) != 48 {
		t.Fatal("expected node48 to reuse the free position.")
	}
	for i := 0; i < 256; i++ {
		expected := (i < 48 && i != 17) || i == 250
		if tree.Contains(string([]byte{byte(i)})) != expected {
			t.Errorf("incorrect lookup for byte %d after reusing position.", i)
		}
	}
}

func TestSplitChildInFullNode(t *testing.T) {
	for _, count := range []int{4, 16, 48, 256} {
		tree := &Trie{}
		for i := 0; i < count; i++ {
			tree.Insert(string([]byte{byte(i)}) + "apple")
		}
		// Splitting an existing child must not grow its parent.
		if !tree.Insert("\x00apply") {
			t.Fatal("expected new word to be inserted.")
		}
		if tree.head.count != count || childCapacity(&tree.head) != count {
			t.Fatalf("expected full node%d to keep its storage, got %T", count, tree.head.children)
		}
		for i := 0; i < count; i++ {
			value := string([]byte{byte(i)}) + "apple"
			if !tree.Contains(value) {
				t.Errorf("expected %q to remain after splitting.", value)
			}
		}
		if !tree.Contains("\x00apply") || tree.Size() != count+1 {
			t.Error("expected both words to exist under the split branch.")
		}
	}
}

func TestMixedOperations(t *testing.T) {
	tree := &Trie{}
	expected := make(map[string]bool)
	random := rand.New(rand.NewSource(1))
	values := []string{"", "app", "apple", "apply", "orders.eu", "orders.us"}
	for i := 0; i < 300; i++ {
		value := make([]byte, random.Intn(8)+1)
		for j := range value {
			value[j] = byte(random.Intn(256))
		}
		values = append(values, string(value))
	}

	for i := 0; i < 3000; i++ {
		value := values[random.Intn(len(values))]
		switch random.Intn(3) {
		case 0:
			if tree.Insert(value) != !expected[value] {
				t.Fatalf("incorrect insert result for %q at step %d", value, i)
			}
			expected[value] = true
		case 1:
			if tree.Delete(value) != expected[value] {
				t.Fatalf("incorrect delete result for %q at step %d", value, i)
			}
			delete(expected, value)
		case 2:
			if tree.Contains(value) != expected[value] {
				t.Fatalf("incorrect lookup result for %q at step %d", value, i)
			}
		}
		if tree.Size() != len(expected) {
			t.Fatalf("expected %d elements, got %d", len(expected), tree.Size())
		}

		if i%50 == 0 {
			prefix := value[:random.Intn(len(value)+1)]
			var got []string
			matched := tree.WalkPrefix(prefix, func(key string) { got = append(got, key) })
			var matches []string
			for key := range expected {
				if strings.HasPrefix(key, prefix) {
					matches = append(matches, key)
				}
			}
			slices.Sort(matches)
			if !slices.Equal(got, matches) || matched != (len(matches) > 0) {
				t.Fatalf("expected matches %q for %q, got %q", matches, prefix, got)
			}
		}
	}
}
