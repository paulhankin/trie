package trie

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

var input = `
CAT
CAB
DOG
CABBAGE
CRIBBAGE
`

func newNav() *Navigator {
	b := NewPackedBuilder(NewAlphaCharMap())
	t, err := ReadWords(b, strings.NewReader(input))
	if err != nil {
		panic("error")
	}
	return NewNavigator(t)
}

func newEnglish() *Navigator {
	f, err := os.Open("./2of12.txt")
	if err != nil {
		panic("failed to open dict")
	}
	defer func() { f.Close() }()
	b := NewPackedBuilder(NewAlphaCharMap())
	tr, err := ReadWords(b, f)
	if err != nil {
		panic("error")
	}
	return NewNavigator(tr)
}

func TestPrint(t *testing.T) {
	n := newNav()
	total := 0
	fmt.Printf("ALL WORDS!\n")
	for w := range n.ValidWordsChan() {
		fmt.Printf("%s, ", w)
		total += len(w)
	}
	fmt.Printf("\nDONE!\n")
	if total != 24 {
		t.Fail()
	}
}

func TestCAT(t *testing.T) {
	n := newNav()
	n.Push('C')
	if !n.IsPrefix() {
		t.Fail()
	}
	n.Push('A')
	if !n.IsPrefix() {
		t.Fail()
	}
	n.Push('T')
	if !n.IsPrefix() {
		t.Fail()
	}
	if !n.IsWord() {
		t.Fail()
	}
}

func TestBackup(t *testing.T) {
	n := newNav()
	n.Push('C')
	n.Push('A')
	n.Push('T')
	n.Pop()
	n.Push('B')
	if !n.IsWord() {
		t.Fail()
	}
	if n.Word() != "CAB" {
		t.Fail()
	}
}

func TestReset(t *testing.T) {
	n := newNav()
	n.PushString("CAT")
	if n.Word() != "CAT" {
		t.Fail()
	}
	n.Reset()
	n.PushString("DOG")
	if n.Word() != "DOG" {
		t.Fail()
	}
	if !n.IsWord() {
		t.Fail()
	}
}

func TestGarbage(t *testing.T) {
	n := newNav()
	n.PushString("CAZYXW")
	if n.Word() != "CAZYXW" {
		t.Fail()
	}
	if n.IsWord() {
		t.Fail()
	}
}

func TestAll(t *testing.T) {
	n := newNav()
	n.PushString("CA")
	result := make([]string, 0, 1)
	result = n.All(result)
	if len(result) != 3 {
		t.Errorf("Wrong words in %v", result)
	}
	if n.Count() != len(result) {
		t.Errorf("Counted %d words, but found %d.", n.Count(), len(result))
	}
}

func TestFull(t *testing.T) {
	n := newEnglish()
	n.PushString("STRING")
	if !n.IsWord() {
		t.Errorf("STRING isn't a word?")
	}
	// There should be a few words starting with STRING.
	wc := n.Count()
	if wc < 3 || wc > 20 {
		t.Errorf("Words found: %d", wc)
	}
	n.Reset()
	wc = n.Count()
	fmt.Printf("Words: %d\n", n.Count())
	if wc < 20000 {
		t.Errorf("Not enough words found.")
	}
}
