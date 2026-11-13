package adaptiveradixtrie

import "strings"

type trieNode struct {
	prefix   string
	children any  // nil, *node4, *node16, *node48 or *node256
	count    int  // Number of children, not number of stored words.
	terminal bool // is end of word
}

// Keys are compared byte by byte. A zero-value Trie is ready to use.
type Trie struct {
	head trieNode
	size int
}

// Return max common prefix length of 2 strings (a, b).
func commonPrefix(a, b string) int {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i += 1
	}

	return i
}

func (t *Trie) Insert(value string) bool { // Insert a string, return true if not exist.
	current := &t.head

	for len(value) > 0 { // while value len is > 0
		idx := value[0]
		child := current.findChild(idx)

		// No matching, store the entire block as once.
		if child == nil {
			current.setChild(idx, &trieNode{
				prefix:   value,
				terminal: true,
			})
			t.size += 1
			return true // Added whole block, return.
		}

		matched := commonPrefix(value, child.prefix)

		// If matching prefix length to child prefix length, then move pointer to the child AND continue with the remaining block
		if matched == len(child.prefix) {
			value = value[matched:] // Get the remaining
			current = child         // Move pointer to child
			continue
		}

		// Only parts are matching at this point (value match < child.prefix match), we then split the existing segment
		branch := &trieNode{
			prefix:   child.prefix[:matched], // Split to last match position.
			terminal: false,
		}
		current.setChild(idx, branch)
		child.prefix = child.prefix[matched:]
		branch.setChild(child.prefix[0], child)

		remaining := value[matched:]
		if len(remaining) == 0 {
			// inserted value ends at new branch.
			branch.terminal = true
		} else {
			branch.setChild(remaining[0], &trieNode{
				prefix:   remaining,
				terminal: true,
			})
		}

		t.size += 1
		return true
	}

	// We reached an existing node at the exact end of the value.
	if current.terminal {
		return false
	}

	current.terminal = true
	t.size += 1
	return true
}

func (t *Trie) Contains(value string) bool {
	current := &t.head

	for len(value) > 0 {
		child := current.findChild(value[0])
		if child == nil || !strings.HasPrefix(value, child.prefix) {
			return false
		}

		value = value[len(child.prefix):] // Update our value to continue the search.
		current = child
	}

	return current.terminal // If is end of word -> contains!
}

func (t *Trie) Delete(value string) bool { // Return true if exist
	current := &t.head
	if !deleteValue(current, value) {
		return false
	}

	t.size -= 1
	return true
}

func deleteValue(current *trieNode, value string) bool {
	if len(value) == 0 {
		if !current.terminal {
			return false
		}

		current.terminal = false
		return true
	}

	idx := value[0]
	child := current.findChild(idx)

	if child == nil || !strings.HasPrefix(value, child.prefix) {
		return false
	}

	if !deleteValue(child, value[len(child.prefix):]) {
		return false
	}

	// A stored word must retain its own ending node.
	if child.terminal {
		return true
	}

	if child.count == 0 {
		current.removeChild(idx) // Nothing remains in this branch.
	} else if child.count == 1 {
		// Find the only child and merge both segments.
		for i := 0; i < 256; i++ {
			onlyChild := child.findChild(byte(i))
			if onlyChild != nil {
				child.prefix += onlyChild.prefix
				child.terminal = onlyChild.terminal
				child.children = onlyChild.children
				child.count = onlyChild.count
				break
			}
		}
	}

	return true
}

func (t *Trie) WalkPrefix(prefix string, visit func(string)) bool {
	if t.size == 0 {
		return false
	}

	current := &t.head
	value := ""

	for len(prefix) > 0 {
		child := current.findChild(prefix[0])

		if child == nil {
			return false
		}

		// The requested prefix may end inside a node's segment.
		// Example: searching "app" reaches the segment "apple".
		if strings.HasPrefix(child.prefix, prefix) {
			walk(child, value+child.prefix, visit)
			return true
		}

		if !strings.HasPrefix(prefix, child.prefix) {
			return false
		}

		value += child.prefix
		prefix = prefix[len(child.prefix):]
		current = child
	}

	walk(current, value, visit)
	return true
}

func walk(current *trieNode, value string, visit func(string)) {
	if current.terminal {
		visit(value)
	}

	// Small nodes keep children sorted. Larger nodes use the byte as index.
	switch children := current.children.(type) {
	case *node4:
		for i := 0; i < current.count; i++ {
			child := children.children[i]
			walk(child, value+child.prefix, visit)
		}
	case *node16:
		for i := 0; i < current.count; i++ {
			child := children.children[i]
			walk(child, value+child.prefix, visit)
		}
	case *node48:
		for _, index := range children.index {
			if index != 0 {
				child := children.children[index-1]
				walk(child, value+child.prefix, visit)
			}
		}
	case *node256:
		for _, child := range children.children {
			if child != nil {
				walk(child, value+child.prefix, visit)
			}
		}
	}
}

func (t *Trie) Size() int {
	return t.size
}
