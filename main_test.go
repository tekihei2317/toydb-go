package main

import (
	"bufio"
	"io"
	"os/exec"
	"testing"
)

func runScript(commands []string) ([]string, error) {
	var output []string

	cmd := exec.Command("./toydb-go")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return output, err
	}

	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return output, err
	}

	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		return output, err
	}

	for _, command := range commands {
		io.WriteString(stdin, command+"\n")
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		output = append(output, scanner.Text())
	}

	if err := cmd.Wait(); err != nil {
		return output, err
	}

	return output, err
}

func TestInsertAndRetrieveRow(t *testing.T) {
	commands := []string{
		"insert 1 user1 person1@example.com",
		"select",
		".exit",
	}

	result, err := runScript(commands)

	if err != nil {
		t.Errorf("Error: %s\n", err.Error())
	}

	expected := []string{
		"db > Executed.",
		"db > (1, user1, person1@example.com)",
		"Executed.",
		"db > ",
	}

	for i, expectedLine := range expected {
		if result[i] != expectedLine {
			t.Errorf("expected %s, but got %s\n", expectedLine, result[i])
		}
	}
}
