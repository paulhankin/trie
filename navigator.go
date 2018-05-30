package trie

// A Navigator allows a trie to be navigated, by keeping previous state
// (for example the prefix and the nodes back up the trie).
type Navigator struct {
	nodes  []TrieNode
	prefix []byte
}

// NewNavigator creates a navigator for the provided trie.
func NewNavigator(tn TrieNode) *Navigator {
	result := &Navigator{make([]TrieNode, 0, 20), make([]byte, 0, 20)}
	result.nodes = append(result.nodes, tn)
	return result
}

// Iterate through the trie, sending all valid words along the channel.
func (n *Navigator) allValidWords(words chan string) {
	if !n.IsPrefix() {
		return
	}
	if n.IsWord() {
		words <- n.Word()
	}
	for c := byte('A'); c <= byte('Z'); c++ {
		n.Push(c)
		n.allValidWords(words)
		n.Pop()
	}
}

// ValidWordsChan returns a channel that recieves all words from the current
// place in the trie. The navigator should not be used until all words have been
// read from the channel.
func (n *Navigator) ValidWordsChan() chan string {
	c := make(chan string)
	go func() {
		n.allValidWords(c)
		close(c)
	}()
	return c
}

// All returns all words that start from the current place in the trie.
func (n *Navigator) All(result []string) []string {
	for w := range n.ValidWordsChan() {
		result = append(result, w)
	}
	return result
}

// Count returns the number of words at the current location.
func (n *Navigator) Count() int {
	result := 0
	for _ = range n.ValidWordsChan() {
		result++
	}
	return result
}

// IsWord reports whether we are at the end of a word.
func (n *Navigator) IsWord() bool {
	return n.lastNode().IsWord()
}

func (n *Navigator) lastNode() TrieNode {
	return n.nodes[len(n.nodes)-1]
}

// IsPrefix reports whether the there are any words with the
// prefix represented by the current postion in the trie.
func (n *Navigator) IsPrefix() bool {
	return n.lastNode().IsPrefix()
}

// Word returns the current prefix.
func (n *Navigator) Word() string {
	return string(n.prefix)
}

// Push descends the trie along the given character.
func (n *Navigator) Push(c byte) {
	n.prefix = append(n.prefix, c)
	n.nodes = append(n.nodes, n.lastNode().Follow(c))
}

// PushString descends the trie, by pushing the characters
// of the given string.
func (n *Navigator) PushString(s string) {
	for _, c := range s {
		n.Push(byte(c))
	}
}

// Pop removes the last character of the current prefix, returning back
// up the trie.
func (n *Navigator) Pop() {
	n.prefix = n.prefix[:len(n.prefix)-1]
	n.nodes = n.nodes[:len(n.nodes)-1]
}

// Reset returns to the root of the trie.
func (n *Navigator) Reset() {
	n.prefix = n.prefix[:0]
	n.nodes = n.nodes[:1]
}
