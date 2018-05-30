/*
  Package trie provides datastructures that can store sets of words, and
  methods to interrogate these sets.

  Three different implementations are provided. The first is a simple but
  memory-inefficient representation where each node in the trie is stored
  as a vector of pointers.

  The code to construct this trie looks like this:
      b := trie.NewSimpleBuilder(trie.NewAlphaCharMap())
      t, err := trie.ReadWords(b, reader)
  where reader is an io.Reader that stores all-capital-letter words, one
  per line.

  A second implementation compresses the simple trie by merging all equivalent
  nodes. (Two nodes are equivalent if the subtries that start at that node
  are the same). This compression is very effective for English. With a
  wordlist of around 69k English words, the uncompressed trie has 170k nodes
  whereas the suffix-compressed trie has 32k nodes.

  The code to construct the suffix-compressed trie is similar to the above:
      b := trie.NewSuffixCompressedBuilder(trie.NewAlphaCharMap())
      t, err := trie.ReadWords(b, reader)

  Finally, the third implementation packs the trie into an array of bytes,
  as described in the PhD thesis of Franklin Liang, which can be found at
  http://www.webcitation.org/5pqOfzlIA.

  This, at the cost of a slower construction time and slightly worse
  performance, further reduces the memory-use for the trie. For the dictionary
  as used above, the trie packs into around 300KB, where the original list
  of words was 650KB.

  Again, the code to construct this is similar:
      b := trie.NewPackedBuilder(trie.NewAlphaCharMap())
      t, err := trie.ReadWords(b, reader)

  Each implementation implements the TrieNode interface, allowing branches
  of the trie to be explored, and methods to test if the current prefix is
  either a word or if it can be extended to a word. For example:

  t.Follow('C').Follow('A').Follow('T').IsWord()
  -> true

  t.Follow('C').Follow('A').IsPrefix()
  -> true

  This package also provides a trie.Navigator class, which keeps track
  of the current prefix, and allows backtracking. It also has methods which
  can return or count all words which start with the current prefix.
*/
package trie

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// The TrieNode interface provides information about the set of words
// after a particular prefix (which is implicit).
type TrieNode interface {
	IsPrefix() bool         // Is there any continuation which will result in a word?
	IsWord() bool           // The the current position a word?
	Follow(c byte) TrieNode // Add c to the current prefix.
}

// The Builder interface is used to construct tries. Typically, the trie
// will be populated with words using the ReadWords function.
type Builder interface {
	AddWord(s string) error // AddWord adds the given string to the trie.
	Root() TrieNode         // Return the constructed trie once building is done.
}

// ReadWords reads a file, adding all the words to the trie builder.
func ReadWords(b Builder, in io.Reader) (TrieNode, error) {
	br := bufio.NewReader(in)
	for {
		line, err := br.ReadString('\n')
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			if addErr := b.AddWord(line); addErr != nil {
				fmt.Fprintf(os.Stderr, "Failed to add word %s\n", line)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return b.Root(), nil
}

/* simple is a naive trie node implementation. */
type simple struct {
	isWord   bool      // Is the current node at the end of a word?
	children []*simple // Subtries for each valid continuation.
}

// A CharMap is used to describe the bytes the trie accepts. 0 means the
// byte is not accepted, otherwise bytes with the same number are identified
// within the tree.

// For example, if A=1, B=2, ..., Z=26 and a=1, b=2, ..., z=26, then the trie
// can store and identify words without case sensitivity, and only accept words
// which contain letters.
type CharMap [256]int

type simpleNode struct {
	s  *simple
	cm *CharMap
}

func newSimple(n int) *simple {
	return &simple{
		children: make([]*simple, n),
	}
}

// NewAlphaCharMap creates a default CharMap, suitable for English. It allows A-Z,
// a-z, '-' and apostrophe. It identifies upper and lower case.
func NewAlphaCharMap() *CharMap {
	var cm CharMap
	for i := 'A'; i <= 'Z'; i++ {
		cm[int(i)] = int(i - 'A' + 1)
	}
	for i := 'a'; i <= 'z'; i++ {
		cm[int(i)] = int(i - 'a' + 1)
	}
	cm[int('-')] = 27
	cm[int('\'')] = 28
	return &cm
}

func (cm *CharMap) max() int {
	r := 0
	for i := 0; i < 256; i++ {
		if cm[i] > r {
			r = cm[i]
		}
	}
	return r
}

// NewSimpleBuilder is a builder for a simple trie.
func NewSimpleBuilder(cm *CharMap) Builder {
	return &simpleNode{
		s:  newSimple(cm.max()),
		cm: cm,
	}
}

// AddWord adds a word to a simple trie.
func (b *simpleNode) AddWord(s string) error {
	// First check that we can add every letter in the string.
	for _, c := range s {
		if b.cm[c] == 0 {
			return fmt.Errorf("Bad word %s", s)
		}
	}
	st := b.s
	for _, c := range s {
		i := b.cm[c] - 1
		if st.children[i] == nil {
			st.children[i] = newSimple(b.cm.max())
		}
		st = st.children[i]
	}
	st.isWord = true
	return nil
}

// Root provides the root of a simpleNode trie.
func (b *simpleNode) Root() TrieNode {
	return b
}

// IsPrefix reports whether there are any words that
// have the prefix at the current position.
func (s *simpleNode) IsPrefix() bool {
	return s != nil
}

// IsWord reports whether the current location marks
// the end of a word.
func (s *simpleNode) IsWord() bool {
	return s != nil && s.s.isWord
}

// Follow descends the trie, by appending the given byte
// to the prefix.
func (s *simpleNode) Follow(c byte) TrieNode {
	if s == nil || s.cm[c] == 0 || s.s.children[s.cm[c]-1] == nil {
		return nil
	}
	return &simpleNode{
		s:  s.s.children[s.cm[c]-1],
		cm: s.cm,
	}
}
