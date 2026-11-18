package repl

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

func Start() error {
	scanner := bufio.NewScanner(os.Stdin)
	commands := getCommands()

	for {
		fmt.Print("grimhold > ")

		if !scanner.Scan() {
			fmt.Println("Bye!")
			os.Exit(0)
			return nil
		}

		args := strings.Fields(scanner.Text())
		if len(args) == 0 {
			continue
		}

		cmd, ok := commands[args[0]]
		if !ok {
			fmt.Printf("Unknown command %q. Type 'help' for any help.\n", args[0])
			continue
		}
		if err := cmd.Callback(args[1:]); err != nil {
			if errors.Is(err, errExit) {
				fmt.Println("Bye!")
				return nil
			}
			fmt.Println("Error:", err)
		}
	}
}
