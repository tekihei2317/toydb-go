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
		"- leaf (size 3)",
		"  - 1",
		"  - 2",
		"  - 3",
		"db > ",
	}
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("expected\n%v, but got\n%v", strings.Join(expected, "\n"), strings.Join(results, "\n"))
	}
}

func TestPrintBtreeOfDepthTwo(t *testing.T) {
	beforeEach()

	scripts := []string{}
	for i := range make([]int, 14) {
		scripts = append(scripts, fmt.Sprintf("insert %d user%d person%d@example.com", i+1, i+1, i+1))
	}
	scripts = append(scripts, ".btree", ".exit")

	results, err := runScripts(scripts)
	check(err)

	expected := []string{
		"db > Tree:",
		"- internal (size 1)",
		"  - leaf (size 7)",
		"    - 1",
		"    - 2",
		"    - 3",
		"    - 4",
		"    - 5",
		"    - 6",
		"    - 7",
		"  - key 7",
		"  - leaf (size 7)",
		"    - 8",
		"    - 9",
		"    - 10",
		"    - 11",
		"    - 12",
		"    - 13",
		"    - 14",
		"db > ",
	}
	results = results[14:]
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("expected\n%v, but got\n%v", strings.Join(expected, "\n"), strings.Join(results, "\n"))
	}
}

func assertEqualSlice(t *testing.T, results []string, expected []string) {
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("expected\n%v, but got\n%v", strings.Join(expected, "\n"), strings.Join(results, "\n"))
	}
}

func TestPrintAllRowsInMultiLevelTree(t *testing.T) {
	beforeEach()

	scripts := []string{}
	for i := 1; i <= 15; i++ {
		scripts = append(scripts, fmt.Sprintf("insert %d user%d person%d@example.com", i, i, i))
	}
	scripts = append(scripts, "select", ".exit")

	results, err := runScripts(scripts)
	check(err)

	expected := []string{
		"db > (1, user1, person1@example.com)",
		"(2, user2, person2@example.com)",
		"(3, user3, person3@example.com)",
		"(4, user4, person4@example.com)",
		"(5, user5, person5@example.com)",
		"(6, user6, person6@example.com)",
		"(7, user7, person7@example.com)",
		"(8, user8, person8@example.com)",
		"(9, user9, person9@example.com)",
		"(10, user10, person10@example.com)",
		"(11, user11, person11@example.com)",
		"(12, user12, person12@example.com)",
		"(13, user13, person13@example.com)",
		"(14, user14, person14@example.com)",
		"(15, user15, person15@example.com)",
		"Executed.",
		"db > ",
	}
	assertEqualSlice(t, results[15:], expected)
}

func TestPrintFourLeafNodeBtree(t *testing.T) {
	beforeEach()

	scripts := []string{
		"insert 18 user18 person18@example.com",
		"insert 7 user7 person7@example.com",
		"insert 10 user10 person10@example.com",
		"insert 29 user29 person29@example.com",
		"insert 23 user23 person23@example.com",
		"insert 4 user4 person4@example.com",
		"insert 14 user14 person14@example.com",
		"insert 30 user30 person30@example.com",
		"insert 15 user15 person15@example.com",
		"insert 26 user26 person26@example.com",
		"insert 22 user22 person22@example.com",
		"insert 19 user19 person19@example.com",
		"insert 2 user2 person2@example.com",
		"insert 1 user1 person1@example.com",
		"insert 21 user21 person21@example.com",
		"insert 11 user11 person11@example.com",
		"insert 6 user6 person6@example.com",
		"insert 20 user20 person20@example.com",
		"insert 5 user5 person5@example.com",
		"insert 8 user8 person8@example.com",
		"insert 9 user9 person9@example.com",
		"insert 3 user3 person3@example.com",
		"insert 12 user12 person12@example.com",
		"insert 27 user27 person27@example.com",
		"insert 17 user17 person17@example.com",
		"insert 16 user16 person16@example.com",
		"insert 13 user13 person13@example.com",
		"insert 24 user24 person24@example.com",
		"insert 25 user25 person25@example.com",
		"insert 28 user28 person28@example.com",
		".btree",
		".exit",
	}

	results, err := runScripts(scripts)
	check(err)

	expected := []string{
		"db > Tree:",
		"- internal (size 3)",
		"  - leaf (size 7)",
		"    - 1",
		"    - 2",
		"    - 3",
		"    - 4",
		"    - 5",
		"    - 6",
		"    - 7",
		"  - key 7",
		"  - leaf (size 8)",
		"    - 8",
		"    - 9",
		"    - 10",
		"    - 11",
		"    - 12",
		"    - 13",
		"    - 14",
		"    - 15",
		"  - key 15",
		"  - leaf (size 7)",
		"    - 16",
		"    - 17",
		"    - 18",
		"    - 19",
		"    - 20",
		"    - 21",
		"    - 22",
		"  - key 22",
		"  - leaf (size 8)",
		"    - 23",
		"    - 24",
		"    - 25",
		"    - 26",
		"    - 27",
		"    - 28",
		"    - 29",
		"    - 30",
		"db > ",
	}
	assertEqualSlice(t, results[len(scripts)-2:], expected)
}
