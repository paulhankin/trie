package trie

import (
	"fmt"
	"os"
)

// An scBuilder constructs a suffix-compressed (simple) trie. Nodes that
// are the same (that is, they describe identical strings) are merged. */
type scBuilder simpleNode

// NewSuffixCompressedBuilder returns a builder for a suffix-compressed trie.
func NewSuffixCompressedBuilder(cm *CharMap) Builder {
	sn := simpleNode{
		s:  newSimple(cm.max()),
		cm: cm,
	}
	return (*scBuilder)(&sn)
}

// AddWord adds a word to the trie.
func (b *scBuilder) AddWord(s string) error {
	return (*simpleNode)(b).AddWord(s)
}

// Root returns the root of the constructed suffix-compressed trie.
func (b *scBuilder) Root() TrieNode {
	return compress((*simpleNode)(b))
}

type compressTable struct {
	hashes                        map[*simple]uint32
	done                          map[uint32][]*simple
	duplicates, nodes, collisions int
}

func (ct *compressTable) hash(t *simple) uint32 {
	if ct.hashes[t] == 0 {
		result := uint32(0)
		for i := range t.children {
			result = result * 167
			if t.children[i] != nil {
				result += uint32(i+1) * ct.hash(t.children[i])
			} else {
				result += uint32((i + 1) * 99)
			}
		}
		if t.isWord {
			result = result*7 + 1
		}
		if result == 0 {
			result = 42
		}
		ct.hashes[t] = result
	}
	return ct.hashes[t]
}

// Expensive equality operation!
func nodesEqual(s, t *simple) bool {
	if s == t {
		return true
	}
	if s.isWord != t.isWord {
		return false
	}
	for i := range s.children {
		if (s.children[i] == nil) != (t.children[i] == nil) {
			return false
		}
	}
	for i := range s.children {
		if s.children[i] != nil {
			if !nodesEqual(s.children[i], t.children[i]) {
				return false
			}
		}
	}
	return true
}

func (ct *compressTable) insertNode(t *simple) *simple {
	h := ct.hash(t)
	fresh := false
	if ct.done[h] == nil {
		ct.done[h] = make([]*simple, 0, 1)
		fresh = true
	}
	for _, s := range ct.done[h] {
		if nodesEqual(s, t) {
			ct.duplicates++
			return s
		}
	}
	ct.done[h] = append(ct.done[h], t)
	ct.nodes++
	if !fresh {
		ct.collisions++
	}
	return t
}

func (ct *compressTable) compressNode(t *simple) *simple {
	if t == nil {
		return nil
	}
	for i := range t.children {
		t.children[i] = ct.compressNode(t.children[i])
	}
	return ct.insertNode(t)
}

func compress(t *simpleNode) *simpleNode {
	ct := &compressTable{
		hashes: make(map[*simple]uint32),
		done:   make(map[uint32][]*simple),
	}
	s := ct.compressNode(t.s)
	fmt.Fprintf(os.Stderr, "Nodes:%d\nDuplicates:%d\nCollisions:%d\n", ct.nodes, ct.duplicates, ct.collisions)
	return &simpleNode{
		cm: t.cm,
		s:  s,
	}
}
