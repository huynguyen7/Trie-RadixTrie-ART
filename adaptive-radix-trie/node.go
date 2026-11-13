package adaptiveradixtrie

// Small nodes store the byte and its child at the same position.
type node4 struct {
	keys     [4]byte
	children [4]*trieNode
}

type node16 struct {
	keys     [16]byte
	children [16]*trieNode
}

type node48 struct {
	index    [256]byte // 0 means missing, otherwise child position + 1.
	children [48]*trieNode
}

type node256 struct {
	children [256]*trieNode // Byte is the child position directly.
}

func (current *trieNode) findChild(idx byte) *trieNode {
	switch children := current.children.(type) {
	case *node4:
		for i := 0; i < current.count; i++ {
			if children.keys[i] == idx {
				return children.children[i]
			}
		}
	case *node16:
		for i := 0; i < current.count; i++ {
			if children.keys[i] == idx {
				return children.children[i]
			}
		}
	case *node48:
		index := children.index[idx]
		if index != 0 {
			return children.children[index-1]
		}
	case *node256:
		return children.children[idx]
	}
	return nil
}

// Add a child, or replace it when splitting a segment.
func (current *trieNode) setChild(idx byte, child *trieNode) {
	if current.children == nil {
		current.children = &node4{}
	}

	switch children := current.children.(type) {
	case *node4:
		for i := 0; i < current.count; i++ {
			if children.keys[i] == idx {
				children.children[i] = child
				return // Replacing does not increase count.
			}
		}
		if current.count == 4 {
			current.resize(16)
			current.setChild(idx, child)
			return
		}
		// Shift larger bytes right to keep the children sorted.
		i := current.count
		for i > 0 && children.keys[i-1] > idx {
			children.keys[i] = children.keys[i-1]
			children.children[i] = children.children[i-1]
			i -= 1
		}
		children.keys[i] = idx
		children.children[i] = child
	case *node16:
		for i := 0; i < current.count; i++ {
			if children.keys[i] == idx {
				children.children[i] = child
				return
			}
		}
		if current.count == 16 {
			current.resize(48)
			current.setChild(idx, child)
			return
		}
		i := current.count
		for i > 0 && children.keys[i-1] > idx {
			children.keys[i] = children.keys[i-1]
			children.children[i] = children.children[i-1]
			i -= 1
		}
		children.keys[i] = idx
		children.children[i] = child
	case *node48:
		index := children.index[idx]
		if index != 0 {
			children.children[index-1] = child
			return
		}
		if current.count == 48 {
			current.resize(256)
			current.setChild(idx, child)
			return
		}
		// Deletion can leave a free position anywhere in this array.
		for i, next := range children.children {
			if next == nil {
				children.children[i] = child
				children.index[idx] = byte(i + 1)
				break
			}
		}
	case *node256:
		if children.children[idx] != nil {
			children.children[idx] = child
			return
		}
		children.children[idx] = child
	}
	current.count += 1
}

func (current *trieNode) removeChild(idx byte) {
	if current.findChild(idx) == nil {
		return
	}

	switch children := current.children.(type) {
	case *node4:
		for i := 0; i < current.count; i++ {
			if children.keys[i] == idx {
				copy(children.keys[i:], children.keys[i+1:current.count])
				copy(children.children[i:], children.children[i+1:current.count])
				children.keys[current.count-1] = 0
				children.children[current.count-1] = nil
				break
			}
		}
	case *node16:
		for i := 0; i < current.count; i++ {
			if children.keys[i] == idx {
				copy(children.keys[i:], children.keys[i+1:current.count])
				copy(children.children[i:], children.children[i+1:current.count])
				children.keys[current.count-1] = 0
				children.children[current.count-1] = nil
				break
			}
		}
	case *node48:
		index := children.index[idx]
		children.children[index-1] = nil
		children.index[idx] = 0
	case *node256:
		children.children[idx] = nil
	}
	current.count -= 1

	if current.count == 0 {
		current.children = nil
		return
	}
	// Move back to smaller storage when the children fit.
	switch current.children.(type) {
	case *node16:
		if current.count == 4 {
			current.resize(4)
		}
	case *node48:
		if current.count == 16 {
			current.resize(16)
		}
	case *node256:
		if current.count == 48 {
			current.resize(48)
		}
	}
}

// Copy child pointers to new storage. The child nodes stay where they are.
func (current *trieNode) resize(size int) {
	next := &trieNode{}
	switch size {
	case 4:
		next.children = &node4{}
	case 16:
		next.children = &node16{}
	case 48:
		next.children = &node48{}
	case 256:
		next.children = &node256{}
	}
	for i := 0; i < 256; i++ {
		child := current.findChild(byte(i))
		if child != nil {
			next.setChild(byte(i), child)
		}
	}
	current.children = next.children
}
