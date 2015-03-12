package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/GeertJohan/go.rice"
)

func unpack() (dir string, err error) {
	dir, err = ioutil.TempDir(os.TempDir(), "gueststack")
	if err != nil {
		return "", err
	}
	tgz := rice.MustFindBox("embed")
	cmd := exec.Command("tar", "xzf", "-")
	cmd.Dir = dir
	if cmd.Stdin, err = tgz.Open("module.tar.gz"); err != nil {
		return "", err
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return dir, nil
}

func compile(dir string) error {
	cmd := exec.Command("make")
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func insmod(dir string) error {
	cmd := exec.Command("insmod", "gueststack.ko")
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func MakeModule() error {
	modules, err := exec.Command("lsmod").CombinedOutput()
	if err != nil {
		return err
	}
	if strings.Contains(string(modules), "gueststack") {
		return nil
	}
	dir, err := unpack()
	log.Println("Extracted kernel module to", dir)
	if err != nil {
		return err
	}
	log.Println("Compiling kernel module", dir)
	if err := compile(dir); err != nil {
		return err
	}
	log.Println("Inserting kernel module", dir)
	if err := insmod(dir); err != nil {
		return err
	}
	return nil
}
