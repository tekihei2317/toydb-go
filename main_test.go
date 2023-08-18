package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func beforeEach() {
	os.Remove("test.db")
}

func runScripts(commands []string) ([]string, error) {
	var output []string

	cmd := exec.Command("./toydb-go", "test.db")
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
		fmt.Println("start")
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
	beforeEach()

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
	// 1ページしか実装していないため、一旦スキップ
	beforeEach()
}

func TestInsertMaximumLengthString(t *testing.T) {
	beforeEach()

	longUsername := strings.Repeat("a", 32)
	longEmail := strings.Repeat("a", 256)

	commands := []string{
		fmt.Sprintf("insert 1 %s %s", longUsername, longEmail),
		"select",
		".exit",
	}
	results, err := runScripts(commands)

	if err != nil {
		t.Errorf("Error: %s\n", err.Error())
	}

	expected := []string{
		"db > Executed.",
		fmt.Sprintf("db > (1, %s, %s)", longUsername, longEmail),
		"Executed.",
		"db > ",
	}

	for i, expectedLine := range expected {
		if results[i] != expectedLine {
			t.Errorf("expected %s, but got %s\n", expectedLine, results[i])
		}
	}
}

func TestInsertTooLongString(t *testing.T) {
	beforeEach()

	longUsername := strings.Repeat("a", 33)
	longEmail := strings.Repeat("b", 257)

	commands := []string{
		fmt.Sprintf("insert 1 %s %s", longUsername, longEmail),
		"select",
		".exit",
	}
	results, err := runScripts(commands)

	if err != nil {
		t.Errorf("Error: %s\n", err.Error())
	}

	expected := []string{
		"db > String is too long.",
		"db > Executed.",
		"db > ",
	}

	for i, expectedLine := range expected {
		if results[i] != expectedLine {
			t.Errorf("expected \"%s\", but got \"%s\"\n", expectedLine, results[i])
		}
	}
}

func TestNegativeId(t *testing.T) {
	beforeEach()

	commands := []string{
		"insert -1 tekihei tekihei@example.com",
		"select",
		".exit",
	}

	results, err := runScripts(commands)

	if err != nil {
		t.Errorf("Error: %s\n", err.Error())
	}

	expected := []string{
		"db > ID must be positive.",
		"db > Executed.",
		"db > ",
	}

	for i, expectedLine := range expected {
		if results[i] != expectedLine {
			t.Errorf("expected \"%s\", but got \"%s\"\n", expectedLine, results[i])
		}
	}
}

func TestKeepsDataAfterClosingConnection(t *testing.T) {
	beforeEach()

	results, err := runScripts([]string{
		"insert 1 user1 person1@example.com",
		".exit",
	})
	check(err)

	expected := []string{
		"db > Executed.",
		"db > ",
	}
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("expected %v, but got %v", expected, results)
	}

	results2, err := runScripts([]string{
		"select",
		".exit",
	})
	check(err)

	expected2 := []string{
		"db > (1, user1, person1@example.com)",
		"Executed.",
		"db > ",
	}
	if !reflect.DeepEqual(results2, expected2) {
		t.Errorf("expected %v, but got %v", expected2, results)
	}
}

func TestPrintOneNodeBtree(t *testing.T) {
	beforeEach()

	scripts := []string{}
	for _, i := range []int{3, 1, 2} {
		scripts = append(scripts, fmt.Sprintf("insert %d user%d person%d@example.com", i, i, i))
	}
	scripts = append(scripts, ".btree", ".exit")
	results, err := runScripts(scripts)
	check(err)

	expected := []string{
		"db > Executed.",
		"db > Executed.",
		"db > Executed.",
		"db > Tree:",
		"leaf (size 3)",
		"  - 0 : 1",
		"  - 1 : 2",
		"  - 2 : 3",
		"db > ",
	}
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("expected\n%v, but got\n%v", strings.Join(expected, "\n"), strings.Join(results, "\n"))
	}
}
