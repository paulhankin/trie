package trie

import (
	"fmt"
	"os"
)

type packedTable struct {
	cells []uint32
	cm    *CharMap
}

func (p *packedTable) SetKey(idx int, k uint8) {
	p.cells[idx] &^= 255
	p.cells[idx] |= uint32(k)
}

func (p *packedTable) SetValue(idx int, v uint32) {
	p.cells[idx] &= 255
	p.cells[idx] |= v << 8
}

func (p *packedTable) Value(idx int) uint32 {
	return p.cells[idx] >> 8
}

func (p *packedTable) Key(idx int) uint8 {
	return uint8(p.cells[idx] & 255)
}

type packedNode struct {
	tbl *packedTable
	idx int
}

type packedBuilder struct {
	scb *scBuilder
}

// NewPackedBuilder returns a builder for a packed suffix-compressed trie.
// This produces a highly compact form for the trie, at some cost in
// construction and slightly less performant navigation around the trie.
//
// The construction of the packed trie follows the PhD thesis of F.M.Liang,
// as used in the hyphenation dictionary of TeX.
// See: https://www.webcitation.org/5pqOfzlIA
func NewPackedBuilder(cm *CharMap) Builder {
	sn := simpleNode{
		s:  newSimple(cm.max()),
		cm: cm,
	}
	return packedBuilder{(*scBuilder)(&sn)}
}

// AddWord adds a word to the packed trie.
func (b packedBuilder) AddWord(s string) error {
	return b.scb.AddWord(s)
}

// Root returns the root of this packed trie.
func (b packedBuilder) Root() TrieNode {
	result := fromSimple(compress((*simpleNode)(b.scb)))
	fmt.Fprintf(os.Stderr, "Size of table=%d\n", len(result.tbl.cells))
	gaps := 0
	for idx := range result.tbl.cells {
		if result.tbl.Key(idx) == 0 {
			gaps += 1
		}
	}
	fmt.Fprintf(os.Stderr, "Gaps=%d\n", gaps)
	return result
}

var space [256]uint32

func (t *packedTable) ensureIndex(i, N int) int {
	for len(t.cells) < i+N {
		t.cells = append(t.cells, space[:N]...)
	}
	return i
}

func (t *packedTable) nextGap(i int) int {
	for i = i + 1; i < len(t.cells) && t.cells[i] != 0; i++ {
	}
	return i
}

// findIndex finds a gap in the table where the given node can legally be placed,
// producing extra space if not enough exists.
func (t *packedTable) findIndex(s *simple) int {
	/* TODO(paulhankin): Rather than scanning linearly through the table, keep track
	   of where gaps exist. Currently packing the table takes a few seconds for a
	   large dictionary. */
	N := len(s.children)
	offset := N + 1
	if s.isWord {
		offset = 0
	} else {
		for i := 0; i < N; i++ {
			if s.children[i] != nil {
				offset = i + 1
				break
			}
		}
	}
	for i := t.nextGap(-1) - offset; i < len(t.cells); i = t.nextGap(i+offset) - offset {
		if i < 0 {
			continue
		}
		ok := true
		for j := 0; j < N+1; j++ {
			if i+j >= len(t.cells) {
				return t.ensureIndex(i, N+1)
			}
			if t.Key(i+j) == uint8(j+1) {
				ok = false
				break
			}
			if t.Key(i+j) == 0 {
				continue
			}
			if j == 0 && s.isWord {
				ok = false
				break
			}
			if j > 0 && s.children[j-1] != nil {
				ok = false
				break
			}
		}
		if ok {
			return t.ensureIndex(i, N+1)
		}
	}
	return t.ensureIndex(len(t.cells), N+1)
}

func (t *packedTable) insertNode(s *simple, indexes map[*simple]int) int {
	if indexes[s] != 0 {
		return indexes[s] - 1
	}
	idx := t.findIndex(s)
	indexes[s] = idx + 1
	if s.isWord {
		t.SetKey(idx, uint8(1))
	}
	for i := range s.children {
		if s.children[i] != nil {
			t.SetKey(idx+i+1, uint8(i+2))
		}
	}
	for i := range s.children {
		ci := -1
		if s.children[i] != nil {
			ci = t.insertNode(s.children[i], indexes)
			t.SetValue(idx+i+1, uint32(ci))
		}
	}
	return idx
}

func fromSimple(s *simpleNode) packedNode {
	table := &packedTable{
		make([]uint32, 0, 10240),
		s.cm,
	}
	indexes := make(map[*simple]int)
	idx := table.insertNode(s.s, indexes)
	return packedNode{table, idx}
}

func (t *packedTable) indexAt(idx, c int) int {
	k := idx + c
	if k >= len(t.cells) || t.Key(k) != uint8(c+1) {
		return -1
	}
	return int(t.Value(k))
}

// IsPrefix reports whether any words can be completed from
// the current position in the trie.
func (p packedNode) IsPrefix() bool {
	return p.idx != -1
}

// IsWord reports whether the current position marks the
// end of a word.
func (p packedNode) IsWord() bool {
	return p.idx != -1 && p.tbl.indexAt(p.idx, 0) != -1
}

var badNode = packedNode{nil, -1}

// Follow descends the packed trie along the given character.
func (p packedNode) Follow(c byte) TrieNode {
	if p.idx == -1 || p.tbl.cm[c] == 0 {
		return badNode
	}
	idx := p.tbl.indexAt(p.idx, p.tbl.cm[c])
	if idx == -1 {
		return badNode
	}
	return packedNode{p.tbl, idx}
}
