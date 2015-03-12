package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type SymbolType string

const (
	t SymbolType = "t"
	T            = "T"
	a            = "a"
	A            = "A"
	b            = "b"
	g            = "g"
	i            = "i"
	I            = "I"
	N            = "N"
	p            = "p"
	r            = "r"
	R            = "R"
	s            = "s"
	S            = "S"
	u            = "u"
	U            = "U"
	V            = "V"
	v            = "v"
	W            = "W"
	w            = "w"
)

type Symbol struct {
	Addr uint64
	Type SymbolType
	Sym  string
}

type SymbolTable []Symbol

func (t SymbolTable) Len() int {
	return len(t)
}

func (t SymbolTable) Less(i, j int) bool {
	return t[i].Addr < t[j].Addr
}

func (t SymbolTable) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t SymbolTable) Search(addr uint64) Symbol {
	rv := sort.Search(t.Len(), func(i int) bool {
		return t[i].Addr > addr
	})
	if rv == 0 {
		return Symbol{addr, "", "UNKNOWN"}
	}
	return t[rv-1]
}

func MustParseAddr(s string) uint64 {
	addr, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		log.Fatalln("Error: bad addr in line", s, err)
	}
	return addr
}

func ReadSymbolTable(r io.Reader) (SymbolTable, error) {
	t := make(SymbolTable, 0)
	for scanner := bufio.NewScanner(r); scanner.Scan(); {
		parts := strings.Split(scanner.Text(), " ")
		if len(parts) != 3 {
			return nil, errors.New("Error: unexpected line" + fmt.Sprint(parts))
		}
		t = append(t, Symbol{MustParseAddr(parts[0]), SymbolType(parts[1]), parts[2]})
	}
	return t, nil
}

type twoSorter struct {
	key []int
	val []string
}

func (s twoSorter) Len() int {
	return len(s.key)
}

func (s twoSorter) Less(i, j int) bool {
	return s.key[i] < s.key[j]
}

func (s twoSorter) Swap(i, j int) {
	s.key[i], s.key[j] = s.key[j], s.key[i]
	s.val[i], s.val[j] = s.val[j], s.val[i]
}

func SymbolTableFromFile(path string) (SymbolTable, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	return ReadSymbolTable(fd)
}

/*func main() {
	prependAddr := flag.Bool("prependaddr", false, "prepend the hex address to the symbol")
	needsort := flag.Bool("sort", false, "prepend the hex address to the symbol")
	csv := flag.Bool("csv", false, "output as csv")
	symfile := flag.String("kallsyms", "", "symbol table to resolve addresses with, gen with nm -n")
	flag.Parse()
	if *symfile == "" {
		fmt.Fprintln(os.Stderr, "Must specify symbol file to resolve symbols with")
		return
	}
	if *needsort {
		symbols := make(map[string]int)
		for scanner := bufio.NewScanner(os.Stdin); scanner.Scan(); {
			addr := MustParseAddr(scanner.Text())
			sym := t.Search(addr).Sym
			if *prependAddr {
				sym = fmt.Sprintf("%x-%s", addr, t.Search(addr).Sym)
			}
			symbols[sym] += 1
		}
		keys := make([]string, 0, len(symbols))
		counts := make([]int, 0, len(symbols))
		total := 0
		for key, c := range symbols {
			keys = append(keys, key)
			counts = append(counts, c)
			total += c
		}
		sort.Sort(twoSorter{counts, keys})
		for i := range keys {
			if *csv {
				fmt.Printf("%f%%,%d,%s\n", 100*float64(counts[i])/float64(total), counts[i], keys[i])
			} else {
				fmt.Printf("%6.2f%%(%d samples) %s\n", 100*float64(counts[i])/float64(total), counts[i], keys[i])
			}
		}
	} else {
		for scanner := bufio.NewScanner(os.Stdin); scanner.Scan(); {
			line := scanner.Text()
			r := regexp.MustCompile(`[a-f0-9]{16}`)
			for {
				loc := r.FindStringIndex(line)
				if loc == nil {
					fmt.Println(line)
					break
				}
				fmt.Print(line[0:loc[0]])
				addr := MustParseAddr(line[loc[0]:loc[1]])
				sym := t.Search(addr).Sym
				if *prependAddr {
					sym = fmt.Sprintf("%x-%s", addr, t.Search(addr).Sym)
				}
				fmt.Print(sym)
				line = line[loc[1]:]
			}
		}
	}
}*/
