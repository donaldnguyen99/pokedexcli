package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func cleanInput(text string) []string {
	var textSlice []string
	for _, word := range strings.Split(text, " ") {
		if word == "" {
			continue
		}
		textSlice = append(textSlice, strings.ToLower(strings.Trim(word, " ")))
	}
	
	return textSlice
}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	if len(commands) == 0 {
		return fmt.Errorf("no commands available")
	}
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println("")
	for _, command := range commands {
		fmt.Printf("%s: %s\n", command.name, command.description)
	}
	return nil
}

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

var commands map[string]cliCommand

func main() {
	
	scanner :=bufio.NewScanner(os.Stdin)
	fmt.Printf("Pokedex > ")
	for ; scanner.Scan(); fmt.Printf("Pokedex > ") {
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}
		
		text := scanner.Text()
		words := cleanInput(text)
		
		commands = map[string]cliCommand{
			"help": {
				name:        "help",
				description: "Displays a help message",
				callback:    commandHelp,
			},
			"exit": {
				name:        "exit",
				description: "Exit the Pokedex",
				callback:    commandExit,
			},
		}

		if len(words) == 0 {
			continue
		}
		command, ok := commands[words[0]]
		if !ok {
			fmt.Println("Invalid command. Please try again.")
			continue
		}
		err := command.callback()
		if err != nil {
			fmt.Printf("Error while executing %s command: %v\n", command.name, err)
			continue
		}
	}
}