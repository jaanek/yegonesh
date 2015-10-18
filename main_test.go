package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"testing"
)

func (c *Command) String() string {
	return c.Name
}

func TestByScore_Sort(t *testing.T) {
	c1 := &Command{Name: "vim", Calls: 2}
	c2 := &Command{Name: "nano", Calls: 0}
	c3 := &Command{Name: "emacs", Calls: 1}

	result := ByScore{Commands{c1, c2, c3}}
	expected := ByScore{Commands{c2, c3, c1}}
	sort.Sort(result)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected ByScore sorted to equal: %v\ngot: %v", expected, result)
	}
}

func TestCheck(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected error check to have panicked")
		}
	}()
	check(errors.New("test"))
}

func TestFetchExecutables(t *testing.T) {
	dir, _ := ioutil.TempDir("", "yegonesh")
	defer os.RemoveAll(dir)
	files := []struct {
		name string
		perm os.FileMode
		add  bool
	}{
		{"emacs", 0111, true},
		{"nano", 0666, false},
		{"vim", 0744, true},
	}
	for _, f := range files {
		ioutil.WriteFile(dir+"/"+f.name, nil, f.perm)
	}
	out := fetchExecutables(dir)
	expected := []string{"emacs", "vim"}
	var result []string
	for c := range out {
		result = append(result, c)
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected executables to eq %v, got %v", expected, result)
	}
}

func TestScanPath(t *testing.T) {
	dir1, _ := ioutil.TempDir("", "yegonesh_dir1")
	dir2, _ := ioutil.TempDir("", "yegonesh_dir2")
	defer os.RemoveAll(dir1)
	defer os.RemoveAll(dir2)
	files := []struct {
		name string
		dir  string
	}{
		{"vim", dir1},
		{"emacs", dir2},
		{"nano", dir2},
	}
	for _, f := range files {
		ioutil.WriteFile(f.dir+"/"+f.name, nil, 0744)
	}
	out := scanPath(dir1 + ":" + dir2)
	expected := []string{"emacs", "nano", "vim"}
	var result []string
	for c := range out {
		result = append(result, c)
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected executables to eq %v, got %v", expected, result)
	}
}

func TestReadHistory(t *testing.T) {
	history := "2\tvim\n3\temacs\n1\tnano\n"

	dir, _ := ioutil.TempDir("", "yegonesh")
	defer os.RemoveAll(dir)
	name := dir + "/history.tsv"
	ioutil.WriteFile(name, []byte(history), 0644)

	expected := Commands{
		&Command{Name: "emacs", Calls: 3},
		&Command{Name: "vim", Calls: 2},
		&Command{Name: "nano", Calls: 1},
	}
	result := readHistory(name)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected executables to eq %v, got %v", expected, result)
	}

	result = readHistory("bogus_dir")
	if result != nil {
		t.Error("Expected unexistent directory to return nil")
	}
}

func TestWriteHistory(t *testing.T) {
	dir, _ := ioutil.TempDir("", "yegonesh")
	defer os.RemoveAll(dir)
	name := dir + "/history.tsv"
	cmds := Commands{
		&Command{Name: "emacs", Calls: 3},
		&Command{Name: "vim", Calls: 2},
	}
	writeHistory(name, cmds, "vim")
	writeHistory(name, cmds, "nano")
	f, _ := os.Open(name)
	defer f.Close()
	result, _ := ioutil.ReadAll(f)
	expected := []byte("3\temacs\n3\tvim\n1\tnano\n")

	if !bytes.Equal(expected, result) {
		t.Errorf("Expected history to eq %q, got %q", expected, result)
	}
}

func TestMultiplexMenuStreams(t *testing.T) {
	history := make(chan string, 3)
	cmds := make(chan string, 3)
	out := multiplexMenuStreams(history, cmds)

	history <- "vim"
	history <- "emacs"
	history <- "neovim"
	cmds <- "emacs"
	cmds <- "nano"
	cmds <- "vim"
	close(history)
	close(cmds)

	expected := []string{"vim", "emacs", "neovim", "nano"}
	var result []string
	for c := range out {
		result = append(result, c)
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected executables to eq %v, got %v", expected, result)
	}
}

func TestHistoryNameStream(t *testing.T) {
	cmds := Commands{
		&Command{Name: "emacs"},
		&Command{Name: "vim"},
		&Command{Name: "nano"},
	}
	out := historyNameStream(cmds)
	for _, c := range cmds {
		name := <-out
		if name != c.Name {
			t.Errorf("Expected executable %v, got %v", c.Name, name)
		}
	}
}
