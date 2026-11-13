package radixtrie

import "strings"

type trieNode struct {
	prefix string
	children [26]*trieNode
	terminal bool // is end of word
}

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
		idx := value[0] - 'a'
		child := current.children[idx]

		// No matching, store the entire block as once.
		if child == nil {
			current.children[idx] = &trieNode {
				prefix: value,
				terminal: true,
			}
			t.size += 1
			return true // Added whole block, return.
		}

		matched := commonPrefix(value, child.prefix)

		// If matching prefix length to child prefix length, then move pointer to the child AND continue with the remaining block
		if matched == len(child.prefix) {
			value = value[matched:] // Get the remaining
			current = child // Move pointer to child
			continue
		}

		// Only parts are matching at this point (value match < child.prefix match), we then split the existing segment
		branch := &trieNode {
			prefix: child.prefix[:matched], // Split to last match position.
			terminal: false,
		}
		current.children[idx] = branch
		child.prefix = child.prefix[matched:]
		branch.children[child.prefix[0]-'a'] = child

		remaining := value[matched:]
		if len(remaining) == 0 {
			// inserted value ends at new branch.
			branch.terminal = true
		} else {
			branch.children[remaining[0]-'a'] = &trieNode {
				prefix: remaining,
				terminal: true,
			}
		}

		t.size += 1
		return true
	}
	
	// We reached existing node at the exact end of the value , so nothing to add.
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
		child := current.children[value[0]-'a']
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

	index := value[0] - 'a'
	child := current.children[index]

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

	count := 0
	var onlyChild *trieNode

	for _, next := range child.children {
		if next != nil {
			count++
			onlyChild = next
		}
	}

	if count == 0 {
		// Nothing remains in this branch.
		current.children[index] = nil
	} else if count == 1 {
		// Compress the branch again.
		child.prefix += onlyChild.prefix
		child.terminal = onlyChild.terminal
		child.children = onlyChild.children
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
		child := current.children[prefix[0]-'a']

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

	for _, child := range current.children {
		if child != nil {
			walk(child, value+child.prefix, visit)
		}
	}
}

func (t *Trie) Size() int {
	return t.size
}

