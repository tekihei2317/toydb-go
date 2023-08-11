package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"testing"
)

func runScripts(commands []string) ([]string, error) {
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

	result, err := runScripts(commands)

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

func TestPrintErrorWhenTableIsFull(t *testing.T) {
	var commands []string

	// 一列は 8+32+256=296バイト
	// 1ページに4096/296 = 13.8...行
	// ページ数は100なので、1300レコードまでは登録できる
	for i := 1; i <= 1301; i++ {
		commands = append(commands, fmt.Sprintf("insert %d user%d person%d@example.com", i, i, i))
	}
	commands = append(commands, ".exit")

	results, err := runScripts(commands)

	if err != nil {
		t.Errorf("Error: %s\n", err.Error())
	}

	// 後ろから一番目が.exitの時のプロンプトで、その直前に行数の上限を超えている
	errorLine := results[len(results)-2]

	if errorLine != "db > Error: Table is full." {
		t.Errorf("Table is not full. actual message: %s\n", errorLine)
	}
}
