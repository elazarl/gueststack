package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/GeertJohan/go.rice"
)

type Server struct {
	sync.Mutex
	Symbols         map[string]SymbolTable
	GuestStack      *GuestStack
	ELFPath         string
	addrCache       map[uint64]string
	funcCache       map[uint64]string
	flamegraph_path string
}

func (s *Server) SearchSymbol(addr uint64) (string, Symbol) {
	for name, s := range s.Symbols {
		if sym := s.Search(addr); sym.Type != "" {
			return name, sym
		}
	}
	return "", Symbol{addr, "", "UNKNOWN"}
}

func NewServer(symbols map[string]SymbolTable, ELFPath string, gs *GuestStack) (*Server, error) {
	dir, err := ioutil.TempDir(os.TempDir(), "gueststack-flamegraph")
	if err != nil {
		return nil, err
	}
	box := rice.MustFindBox("embed")
	flamegraphpl := filepath.Join(dir, "flamegraph.pl")
	if err := ioutil.WriteFile(flamegraphpl, box.MustBytes("flamegraph.pl"), 0755); err != nil {
		return nil, err
	}
	return &Server{sync.Mutex{}, symbols, gs, ELFPath, make(map[uint64]string), make(map[uint64]string), flamegraphpl}, nil
}

func (s *Server) collapsedStack(w io.Writer, stacks []*Stack) error {
	for _, stack := range stacks {
		if len(stack.Addr) == 0 {
			continue
		}
		for i := len(stack.Addr) - 1; i >= 0; i-- {
			name, sym := s.SearchSymbol(stack.Addr[i])
			if _, err := fmt.Fprintf(w, "%s:%s;", sym.Sym, name); err != nil {
				return err
			}
		}
		if s.ELFPath != "" && false {
			s.Lock()
			if _, ok := s.addrCache[stack.RIP]; !ok {
				cmd := exec.Command("addr2line", "-fe", s.ELFPath, fmt.Sprintf("%x", stack.RIP))
				output, err := cmd.CombinedOutput()
				if err != nil {
					s.Unlock()
					return err
				}
				parts := strings.Split(string(output), "\n")
				if strings.Contains(parts[0], "??") {
					s.funcCache[stack.RIP] = ""
					s.addrCache[stack.RIP] = fmt.Sprintf("%x", stack.RIP)
				} else {
					s.funcCache[stack.RIP] = parts[0]
					s.addrCache[stack.RIP] = filepath.Base(parts[1])
				}
			}
			fn := s.funcCache[stack.RIP]
			rip := s.addrCache[stack.RIP]
			s.Unlock()
			name, sym := s.SearchSymbol(stack.Addr[0])
			if fn != sym.Sym+":"+name && fn != "" {
				fmt.Fprint(w, fn, ";")
			}
			fmt.Fprint(w, rip, ";")
		}
		if _, err := fmt.Fprintln(w, " 1"); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	stacks, err := s.GuestStack.Stack()
	if err != nil {
		log.Panic(err)
	}
	total := float64(0)
	relevant := float64(0)
	for _, stack := range stacks {
		total++
		if len(stack.Addr) > 0 {
			relevant++
		}
	}
	if strings.HasSuffix(req.URL.Path, "/raw") {
		if len(stacks) == 0 {
			return
		}
		if err := s.collapsedStack(w, stacks); err != nil {
			log.Panic(err)
		}
		return
	}
	flamegraph := exec.Command(s.flamegraph_path, "-title",
		fmt.Sprintf("%%%.2f relevant %.0f/%.0f", 100*relevant/total, relevant, total))
	flamegraph.Stdout = w
	fgin, err := flamegraph.StdinPipe()
	if err != nil {
		log.Panic(err)
	}
	if err := flamegraph.Start(); err != nil {
		log.Panic(err)
	}
	if relevant == 0 {
		io.WriteString(w, "No Samples Yet")
		return
	}
	if err := s.collapsedStack(fgin, stacks); err != nil {
		log.Panic(err)
	}
	if err := fgin.Close(); err != nil {
		log.Panic(err)
	}
	if err := flamegraph.Wait(); err != nil {
		log.Panic(err)
	}
}

type APIServer struct {
	GuestStack *GuestStack
}

func (s *APIServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m := map[string]func() error{
		"/api/start": s.GuestStack.Start,
		"/api/stop":  s.GuestStack.Stop,
		"/api/reset": s.GuestStack.Reset,
	}
	f, ok := m[req.URL.Path]
	if !ok {
		http.Error(w, "illegal path "+req.URL.Path, 404)
		return
	}
	if err := f(); err != nil {
		http.Error(w, "error: "+err.Error(), 501)
		return
	}
	io.WriteString(w, "ok")
}

type Index struct {
	index *template.Template
	gs    *GuestStack
}

func NewIndex(gs *GuestStack) (*Index, error) {
	box := rice.MustFindBox("embed")
	tmpl, err := template.New("index").Parse(box.MustString("index.html"))
	if err != nil {
		return nil, err
	}
	return &Index{tmpl, gs}, nil
}

func (ix *Index) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Panic(err)
	}
	status := false
	if ix.gs.cmd.Process != nil && (ix.gs.cmd.ProcessState == nil || ix.gs.cmd.ProcessState.Exited() == false) {
		status = true
	}
	ix.index.Execute(w, struct {
		Hostname string
		Status   bool
	}{hostname, status})
}
