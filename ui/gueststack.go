package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/GeertJohan/go.rice"
)

type AddrRange [2]uint64

type GuestStack struct {
	BaseDir          string
	Ncpu             int
	stacks           []string
	total            []string
	relevant         []string
	relevantAddrPath string
	relevantAddrs    []AddrRange
	makeCmd          func() *exec.Cmd
	cmd              *exec.Cmd
}

func NewGuestStack(symbols map[string]SymbolTable) (*GuestStack, error) {
	gs := &GuestStack{BaseDir: "/sys/kernel/debug/gueststack"}
	files, err := ioutil.ReadDir(gs.BaseDir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "stack") {
			gs.stacks = append(gs.stacks, filepath.Join(gs.BaseDir, file.Name()))
		}
		if strings.HasPrefix(file.Name(), "total") {
			gs.total = append(gs.total, filepath.Join(gs.BaseDir, file.Name()))
		}
		if strings.HasPrefix(file.Name(), "stack") {
			gs.stacks = append(gs.stacks, filepath.Join(gs.BaseDir, file.Name()))
		}
		if file.Name() == "relevant_addr" {
			gs.relevantAddrPath = filepath.Join(gs.BaseDir, file.Name())
		}
	}
	for _, s := range symbols {
		gs.relevantAddrs = append(gs.relevantAddrs, AddrRange{s[0].Addr, s[len(s)-1].Addr})
	}

	dir, err := ioutil.TempDir(os.TempDir(), "gueststack-perf2")
	if err != nil {
		return nil, err
	}
	box := rice.MustFindBox("embed")
	perf2 := filepath.Join(dir, "perf2")
	if err := ioutil.WriteFile(perf2, box.MustBytes("perf2"), 0755); err != nil {
		return nil, err
	}

	gs.makeCmd = func() *exec.Cmd { return exec.Command(perf2) }
	gs.cmd = gs.makeCmd()
	return gs, nil
}

func (gs *GuestStack) Reset() error {
	if err := gs.Stop(); err != nil {
		return err
	}
	// empty stacks
	for _, stack := range gs.stacks {
		dump(stack, "reset")
	}
	fp, err := os.OpenFile(gs.relevantAddrPath, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer fp.Close()
	for _, rng := range gs.relevantAddrs {
		if _, err := fmt.Fprintf(fp, "%x-%x\n", rng[0], rng[1]); err != nil {
			return err
		}
	}
	return nil
}

type Stack struct {
	CPU  int
	RIP  uint64
	Addr []uint64
}

var stackHeader = regexp.MustCompile(`CPU:(\d+) RIP: ?([0-9a-z]+)`)

func ParseStack(r io.Reader, stacks []*Stack) ([]*Stack, error) {
	var stack *Stack
	for scanner := bufio.NewScanner(r); scanner.Scan(); {
		line := scanner.Text()
		if matches := stackHeader.FindStringSubmatch(line); matches != nil {
			if stack != nil {
				stacks = append(stacks, stack)
			}
			cpu, err := strconv.Atoi(matches[1])
			if err != nil {
				return nil, errors.New("CPU number is not int: " + err.Error())
			}
			rip, err := strconv.ParseUint(matches[2], 16, 64)
			if err != nil {
				return nil, errors.New("CPU number is not int: " + err.Error())
			}
			stack = &Stack{CPU: cpu, RIP: rip}
			continue
		}
		addr, err := strconv.ParseUint(line, 16, 64)
		if err != nil {
			return nil, errors.New("stack address unparsable" + err.Error())
		}
		stack.Addr = append(stack.Addr, addr)
	}
	if stack != nil {
		stacks = append(stacks, stack)
	}
	return stacks, nil
}

func (gs *GuestStack) Stack() ([]*Stack, error) {
	stacks := []*Stack{}
	for _, s := range gs.stacks {
		r, err := os.Open(s)
		if err != nil {
			return nil, err
		}
		defer r.Close()
		if stacks, err = ParseStack(r, stacks); err != nil {
			return nil, err
		}
	}
	return stacks, nil
}

func (gs *GuestStack) Stop() error {
	if gs.cmd.Process == nil {
		return nil
	}

	gs.cmd.Process.Signal(os.Interrupt)
	time.Sleep(10 * time.Millisecond)
	// I assume this must succeed, unless the process is in horrendeous state
	gs.cmd.Process.Kill()
	gs.cmd.Process.Wait()
	return nil
}

func (gs *GuestStack) Start() error {
	if err := gs.Stop(); err != nil {
		return err
	}
	gs.cmd = gs.makeCmd()
	in, err := gs.cmd.StdinPipe()
	if err != nil {
		fmt.Println("can't pipe:", err)
		return err
	}
	out, err := os.Open(os.DevNull)
	if err != nil {
		return err
	}
	gs.cmd.Stdout = out
	if err := gs.cmd.Start(); err != nil {
		return err
	}
	go func() {
		gs.cmd.Wait()
	}()
	_, err = io.WriteString(in, `
{
  "attr": {
    "sample_type": [
      "PERF_SAMPLE_IP"
    ],
    "wakeup_events": 1,
    "freq": true,
    "exclude_host": true,
    "exclude_idle": true,
    "sample_freq": 99,
    "config": "PERF_COUNT_HW_CPU_CYCLES",
    "type": "PERF_TYPE_HARDWARE"
  }
}`)
	if err != nil {
		return err
	}
	return in.Close()
}
