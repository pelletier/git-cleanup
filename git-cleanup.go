package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type branches struct {
	selected  string
	locals    []string
	remotes   []string
	protected []string
}

func parseBranchOutput(output string) branches {
	b := branches{
		protected: []string{"master"},
	}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if line[0] == '*' {
			b.selected = line[2:len(line)]
		} else if strings.HasPrefix(line, "remotes/") {
			name := strings.SplitN(line, "/", 3)[2]
			b.remotes = append(b.remotes, name)
		} else {
			b.locals = append(b.locals, line)
		}
	}

	return b
}

func (b *branches) toDelete() []string {
	toDelete := make([]string, 0, len(b.locals))
	for _, local := range b.locals {
		found := false
		for _, remote := range b.remotes {
			if local == remote {
				found = true
				break
			}
		}
		if !found {
			toDelete = append(toDelete, local)
		}
	}
	return toDelete
}

func gitDeleteBranch(branch string) {
	fmt.Println("Deleting", branch)
	cmd := exec.Command("git", "branch", "-D", branch)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not delete branch", branch, ":", err, "\n", out.String())
		os.Exit(3)
	}
}

func main() {
	cmd := exec.Command("git", "branch", "-a")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not list branches:", err)
		os.Exit(1)
	}
	output := out.String()
	branches := parseBranchOutput(output)
	toDelete := branches.toDelete()

	if len(toDelete) == 0 {
		fmt.Println("No branch to delete.")
		return
	}

	fmt.Println("About to delete the following branches:")
	for _, b := range toDelete {
		fmt.Println("-", b)
	}
	fmt.Print("Continue? [y/N] ")
	var continueChar string
	n, err := fmt.Scanln(&continueChar)
	if n != 1 || err != nil {
		fmt.Fprintln(os.Stderr, "Invalid input. Stopping.")
		os.Exit(2)
	}
	if continueChar != "y" {
		return
	}

	fmt.Println("")
	for _, b := range toDelete {
		gitDeleteBranch(b)
	}
}
