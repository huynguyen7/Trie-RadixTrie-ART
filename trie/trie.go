package trie

/*
	Just naive Trie implementation, built for comparison/benchmark to radix trie.
*/

type trieNode struct {
	children [26]*trieNode // pre-initialized to 26 slots for alphabet chars.
	terminal bool // Is end of search.
}

type Trie struct {
	head trieNode
	size int
}

// Return true when value does not exist.
func (t* Trie) Insert(value string) bool {
	node := &t.head
	for i:= 0; i < len(value); i++ {
		idx := value[i] - 'a'
		if node.children[idx] == nil {
			node.children[idx] = &trieNode{}
		}
		node = node.children[idx]
	}
	
	if node.terminal {
		return false // Already exist.
	}

	node.terminal = true // Mark end of word.
	t.size += 1 // size increased by one
	return true
}

func (t* Trie) Contains(value string) bool {
	node := &t.head
	for i := 0; i < len(value); i++ {
		idx := value[i] - 'a'
		if node.children[idx] == nil {
			return false
		}
		node = node.children[idx]
	}

	return node.terminal
}

// Return True if the string exists.
func (t* Trie) Delete(value string) bool {
	node := &t.head

	for i := 0; i < len(value); i++ {
		idx := value[i] - 'a'
		if node.children[idx] == nil {
			return false
		}
		node = node.children[idx]
	}
	if !node.terminal {
		return false
	}

	node.terminal = false
	t.size -= 1
	return true
}

func (t* Trie) WalkPrefix(prefix string, visit func(key string)) bool {
	node := &t.head

	// (1. ) Search for prefix
	for i := 0; i < len(prefix); i++ {
		idx := prefix[i] - 'a'
		if node.children[idx] == nil {
			return false
		}
		node = node.children[idx]
	}

	// (2. ) Explore post prefix search to find all possible paths.
	return walk(node, prefix, visit)
}

func walk(node *trieNode, prefix string, visit func(key string)) bool {
	matched := false
	if node.terminal {
		visit(prefix)
		matched = true
	}

	for i, child := range node.children {
		if child == nil {
			continue
		}
		next := prefix + string(byte('a' + i))
		if walk(child, next, visit) {
			matched = true
		}
	}

	return matched
}

func (t* Trie) Size() int {
	return t.size
}

