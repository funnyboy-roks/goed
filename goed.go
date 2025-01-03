package main

import (
	"bufio"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strings"

	// "log"
	"os"
	os_exec "os/exec"
)

type Mode int

const (
	ModeCommand Mode = iota
	ModeAppend
)

var buffer = []string{}
var mode = ModeCommand
var curr_line = -1
var prompt = ""

func (mode Mode) S() string {
	switch mode {
	case ModeCommand:
		return "COMMAND"
	case ModeAppend:
		return "INSERT"
	}
	panic("unreachable")
}

func parseCommand(s string) (rangeLow int, rangeHigh int, command string, hasRange bool) {
	i := 0
	rangeLow = 0
	rangeHigh = 0
	hasRange = true

	for i < len(s) && '0' <= s[i] && s[i] <= '9' {
		rangeLow *= 10
		rangeLow += (int)(s[i] - '0')
		i += 1
	}

	if rangeLow != 0 {
		rangeLow -= 1
	}

	if i >= len(s) {
		rangeHigh = rangeLow
		command = ""
		return
	}

	if s[i] != ',' {
		if i == 0 {
			return curr_line, curr_line, s, false
		} else {
			return rangeLow, rangeLow, s[i:], false
		}
	}
	i += 1

	startI := i

	for i < len(s) && '0' <= s[i] && s[i] <= '9' {
		rangeHigh *= 10
		rangeHigh += (int)(s[i] - '0')
		i += 1
	}

	if rangeHigh != 0 {
		rangeHigh -= 1
	}

	if i == startI {
		rangeHigh = len(buffer) - 1
	}
	command = s[i:]

	return
}

func getSize() int {
	size := 0
	for i := range buffer {
		size += len(buffer[i])
	}
	return size
}

func exec(cmdStr string) bool {
	if mode == ModeCommand {
		low, high, command, hasRange := parseCommand(cmdStr)
		// fmt.Printf("(%d, %d, %s)\n", low, high, command)
        split := strings.Split(command, " ")
        prefix := split[0]
        args := split[1:]
		switch prefix {
		case "P":
			if hasRange {
				fmt.Println("?")
				break
			}
            prompt = "*"
		case "a":
			if hasRange {
				fmt.Println("?")
				break
			}
			mode = ModeAppend
		case "p":
			if len(buffer) == 0 {
				break
			}
			for i := low; i < high+1; i += 1 {
				fmt.Println(buffer[i])
			}
		case "d":
			buffer = slices.Delete(buffer, low, high+1)
			if low <= curr_line && curr_line <= high {
				curr_line = low - 1
			}
		case "#":
            break
		case "n":
			if len(buffer) == 0 {
				break
			}
			for i := low; i < high+1; i += 1 {
				fmt.Printf("%d\t%s\n", i+1, buffer[i])
			}
		case "q":
			if hasRange {
				fmt.Println("?")
				break
			}
			return false
		case "r":
			if hasRange {
				fmt.Println("?")
				break
			}
            rest := strings.Join(args, " ");
            if rest[0] == '!' {
                command := os_exec.Command("bash", "-c", rest[1:])
                command.Stderr = os.Stderr
                command.Stdin = os.Stdin
                output, err := command.Output()
                if err != nil {
                    fmt.Printf("Err: %s\n", err);
                    break
                }
                split := strings.Split(string(output), "\n")
                before_size := getSize()
                for i := range split {
                    curr_line += 1
                    buffer = slices.Insert(buffer, curr_line, split[i])
                }
                fmt.Printf("%d\n", getSize() - before_size);
                break
            }
			fmt.Println("?")
		case "":
			if hasRange {
				if low != high {
					fmt.Println("?")
					break
				}
				curr_line = low
			} else {
				curr_line += 1
				if curr_line >= len(buffer) {
					fmt.Println("?")
					break
				}
				fmt.Println(buffer[curr_line])
			}
		default:
            if command[0] == '!' {
                command := os_exec.Command("bash", "-c", command[1:])
                command.Stdout = os.Stdout
                command.Stderr = os.Stderr
                command.Stdin = os.Stdin
                if err := command.Run(); err != nil {
                    fmt.Printf("Err: %s\n", err);
                }
                fmt.Println("!")
                break
            }
            if command[0] == 's' && command[1] == '/' {
                cmd := command[2:]
                mid := strings.IndexByte(cmd, '/');
                needle := cmd[:mid]
                repl := cmd[mid + 1:]
                reg, err := regexp.CompilePOSIX(needle)
                if err != nil {
                    fmt.Println("?")
                    break
                }
                fmt.Printf("needle = %s\n", needle)
                fmt.Printf("reg = %s\n", reg)
                fmt.Printf("repl = %s\n", repl)
                for i := low; i < high+1; i += 1 {
                    buffer[i] = reg.ReplaceAllString(buffer[i], repl)
                }
                break
            }
			fmt.Println("?")
		}
	} else if mode == ModeAppend {
		switch cmdStr {
		case ".":
			mode = ModeCommand
		default:
			curr_line += 1
			buffer = slices.Insert(buffer, curr_line, cmdStr)
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
		buffer = append(buffer, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	curr_line = 0
	fmt.Println(getSize())
}

func main() {
	if len(os.Args) == 2 {
		readFile(os.Args[1])
	}
	reader := bufio.NewReader(os.Stdin)
	for {
        if mode == ModeCommand {
            fmt.Printf("%s", prompt)
        }
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if !exec(text) {
			break
		}
	}
}
