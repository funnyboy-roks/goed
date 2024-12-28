package main

import (
	"bufio"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"

	// "log"
	"os"
)

type Mode int

const (
	ModeCommand Mode = iota
	ModeAppend
)

var buffer = []string{}
var mode = ModeCommand
var curr_line = -1

func (mode Mode) S() string {
	switch mode {
	case ModeCommand:
		return "COMMAND"
	case ModeAppend:
		return "INSERT"
	}
	panic("unreachable")
}

func exec(command string) bool {

	if mode == ModeCommand {
		switch command {
		case "a":
			mode = ModeAppend
		case "p":
			fmt.Println(buffer[curr_line])
		case ",p":
			for i := range buffer {
				fmt.Println(buffer[i])
			}
		case "n":
			fmt.Printf("%d\t%s\n", curr_line+1, buffer[curr_line])
		case ",n":
			for i := range buffer {
				fmt.Printf("%d\t%s\n", i+1, buffer[i])
			}
		case "q":
			return false
		default:
			n, err := strconv.Atoi(command)
			if err != nil {
				fmt.Println("?")
				break
			}

			if 0 < n && n <= len(buffer) {
				curr_line = n - 1
			} else {
				fmt.Println("?")
			}
		}
	} else if mode == ModeAppend {
		switch command {
		case ".":
			mode = ModeCommand
		default:
			curr_line += 1
			buffer = slices.Insert(buffer, curr_line, command)
		}
	}
	return true
}

func readFile(path string) {
    // open file
    f, err := os.Open(path)
    if err != nil {
        log.Fatal(err)
    }
    // remember to close the file at the end of the program
    defer f.Close()

    // read the file line by line using scanner
    scanner := bufio.NewScanner(f)

    for scanner.Scan() {
        buffer = append(buffer, scanner.Text());
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }

    curr_line = 1;
}

func main() {
    if len(os.Args) == 2 {
        readFile(os.Args[1]);
    }
	reader := bufio.NewReader(os.Stdin)
	for {
		// fmt.Printf("[%s] ", mode.S())
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if !exec(text) {
			break
		}
	}
}
