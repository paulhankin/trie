// Binary boggle finds all words in a boggle grid.
// It is written as an example use of the trie package.
//
// Example usage:
// boggle -d ~/go/src/github.com/paulhankin/trie/2of12.txt -grid ABCD/TNSE/HARP/ELLO
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/paulhankin/trie"
)

// Flags
var (
	dictFlag = flag.String("d", "./go/src/github.com/paulhankin/trie/2of12.txt", "Path of word dictionary")
	gridFlag = flag.String("grid", "ABCD/TNSE/HARP/ELLO", "Boggle grid, upper case, rows separated by slash (/)")
)

func LoadDict(path string) (trie.TrieNode, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	b := trie.NewPackedBuilder(trie.NewAlphaCharMap())
	tr, err := trie.ReadWords(b, f)
	f.Close()
	return tr, err
}

func Solve(b [4][4]byte, i, j int, d trie.TrieNode, current []byte, found map[string]struct{}) {
	if i < 0 || i >= len(b) || j < 0 || j >= len(b[0]) || b[i][j] == 0 {
		return
	}
	d = d.Follow(b[i][j])
	if !d.IsPrefix() {
		return
	}
	current = append(current, b[i][j])
	// Zap the character that we've used.
	old := b[i][j]
	b[i][j] = 0
	if len(current) >= 2 && d.IsWord() {
		found[string(current)] = struct{}{}
	}
	for dij := 0; dij < 9; dij++ {
		Solve(b, i+dij%3-1, j+dij/3-1, d, current, found)
	}
	// Restore the zapped character, and pop it from the current word.
	b[i][j] = old
	current = current[:len(current)-1]
	// Invariant: b and current are restored to the state they were in when
	// we were called.
}

func usage() {
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Parse()
	rows := strings.Split(*gridFlag, "/")
	if len(rows) != 4 {
		usage()
	}
	var grid [4][4]byte
	for i, row := range rows {
		if len(row) != 4 {
			usage()
		}
		copy(grid[i][:], row)
	}

	dict, err := LoadDict(*dictFlag)
	if err != nil {
		fmt.Printf("Failed to load dictionary: %v\n", err)
		os.Exit(1)
	}
	found := make(map[string]struct{})
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			Solve(grid, i, j, dict, nil, found)
		}
	}
	var r []string
	for k := range found {
		r = append(r, k)
	}
	sort.Strings(r)
	fmt.Println(strings.Join(r, ", "))
	fmt.Printf("\nFound %d words\n", len(found))
}
