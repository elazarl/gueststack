package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func readSymbols(env []string, name string, args ...string) (SymbolTable, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = env
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	stderrstr := ""
	go func() {
		s, _ := ioutil.ReadAll(stderr)
		stderrstr = string(s)
	}()
	var symt SymbolTable
	rv := SymbolTable{}
	if symt, err = ReadSymbolTable(stdout); err != nil {
		return nil, err
	}
	for _, sym := range symt {
		if sym.Type != t && sym.Type != T {
			continue
		}
		rv = append(rv, sym)
	}
	if err := cmd.Wait(); err != nil {
		return nil, errors.New(stderrstr)
	}
	return rv, nil
}

func SymAppend(base, extra SymbolTable) SymbolTable {
	for _, sym := range extra {
		if sym.Type == t || sym.Type == T {
			base = append(base, sym)
		}
	}
	return base
}

func dump(path, data string) error {
	fd, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	_, err = fd.WriteString(data)
	return err
}

func main() {
	symbols := make(map[string]SymbolTable)
	elf_image := flag.String("elf", "", "relevant HVX image of the guest")
	kallsyms := flag.String("kallsyms", "", "guest /proc/kallsyms file")
	ssh := flag.String("ssh", "", "ssh command line arguments for connecting to guest")
	addr := flag.String("http", ":8080", "port to listen on")
	flag.Parse()
	if *ssh != "" {
		log.Println("Fetching kallsyms from guest via SSH")
		s, err := readSymbols(nil, "bash", "-c", fmt.Sprintf("%s sudo cat /proc/kallsyms", *ssh))
		if err != nil {
			log.Println("Cannot read", "/proc/kallsyms", "from host", err)
		} else {
			symbols["kallsyms"] = s
		}
	}
	if *kallsyms != "" {
		log.Println("Reading symbols from", *kallsyms)
		fd, err := os.Open(*kallsyms)
		if err != nil {
			log.Fatal(err)
		}
		s, err := ReadSymbolTable(fd)
		if err != nil {
			log.Fatal("Bad kallsyms file", err)
		}
		symbols[filepath.Base(*kallsyms)] = s
	}
	if *elf_image != "" {
		log.Println("Fetching symbols from a given ELF file")
		s, err := readSymbols(nil, "bash", "-c", fmt.Sprintf("nm -n %s", *elf_image))
		if err != nil {
			log.Println("Cannot read", *elf_image, "with nm -n", err)
		} else {
			symbols[filepath.Base(*elf_image)] = s
		}
	}
	for name, s := range symbols {
		log.Println("Fetched", len(s), "symbols from", name)
	}
	for _, s := range symbols {
		sort.Sort(s)
	}

	log.Println("Compiling the gueststack kernel module, and insmod'ing it")
	if err := MakeModule(); err != nil {
		log.Fatal(err)
	}
	log.Println("Reset /debug/gueststack")
	gs, err := NewGuestStack(symbols)
	if err != nil {
		log.Panic(err)
	}
	if err := gs.Reset(); err != nil {
		log.Panic(err)
	}
	log.Println("Extracting flamegraph.pl")
	fgServer, err := NewServer(symbols, *elf_image, gs)
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/flamegraph", fgServer)
	http.Handle("/flamegraph/raw", fgServer)
	http.Handle("/api/", &APIServer{gs})
	index, err := NewIndex(gs)
	if err != nil {
		log.Panic(err)
	}
	http.Handle("/", index)
	_, port, err := net.SplitHostPort(*addr)
	if err != nil {
		log.Panic(err)
	}
	if strings.HasPrefix(*addr, ":") {
		hostname, _ := os.Hostname()
		log.Println("Open your browser on", hostname+*addr)
	} else {
		log.Println("Open your browser on", *addr)
	}
	if strings.HasPrefix(*addr, "localhost") || strings.HasPrefix(*addr, "127.0.0.1") {
		log.Print("for remote server, use ssh -L", port, ":localhost:", port,
			", and browse to localhost:", port, "\n")
	}
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
}
