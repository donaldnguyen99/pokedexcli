package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/donaldnguyen99/pokedexcli/internal/pokeapi"
)

func cleanInput(text string) []string {
	var textSlice []string
	for _, word := range strings.Split(text, " ") {
		if word == "" {
			continue
		}
		textSlice = append(
			textSlice, 
			strings.ToLower(strings.Trim(word, " ")),
		)
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

func commandMapNextPage(goToNextPage bool) error {
	var fullURL string
	var mapCommand string
	if goToNextPage {
		mapCommand = "map"
		if commands[mapCommand].config.Next == "" {
			fmt.Println("you're on the last page")
			return nil
		}
		fullURL = commands[mapCommand].config.Next
	} else {
		mapCommand = "mapb"
		if commands[mapCommand].config.Previous == "" {
			fmt.Println("you're on the first page")
			return nil
		}
		fullURL = commands[mapCommand].config.Previous
	}

	locationAreasPage, err := locationAreasManager.GetLocationAreasPage(
		fullURL,
	)
	if err != nil {
		return fmt.Errorf("error getting location areas page: %v", err)
	}

	commands[mapCommand].config.Next = locationAreasPage.Next
	commands[mapCommand].config.Previous = locationAreasPage.Previous

	for _, location := range locationAreasPage.Results {
		// urlSplit := strings.Split(location.URL, "/")
		// id := urlSplit[len(urlSplit)-2]
		fmt.Printf("%s\n", location.Name)
	}
	return nil
}

func commandMap() error {
	return commandMapNextPage(true)
}

func commandMapb() error {
	return commandMapNextPage(false)
}

type cliCommand struct {
	name        string
	description string
	callback    func() error
	config      *config
}

type config struct {
	Next     string
	Previous string
}

// These globals aren't ideal, but they'll do for now.
var commands map[string]cliCommand
var locationAreasManager *pokeapi.LocationAreasManager

func main() {
	
	scanner :=bufio.NewScanner(os.Stdin)
	locationAreasManager = pokeapi.NewLocationAreasManager()
	map_config := &config{
		Next:     locationAreasManager.GetLocationAreasPageURL(),
		Previous: "",
	}
	commands = map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
			config:      nil,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
			config:      nil,
		},
		"map": {
			name:        "map",
			description: "Displays the names of 20 locations in the Pokemon world or the next 20 locations.",
			callback:    commandMap,
			config:      map_config,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the names of previous 20 locations in the Pokemon world.",
			callback:    commandMapb,
			config:      map_config,
		},
	}

	// Start REPL
	for fmt.Printf("Pokedex > "); scanner.Scan(); fmt.Printf("Pokedex > ") {
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}
		
		text := scanner.Text()
		words := cleanInput(text)

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
			fmt.Printf(
				"Error while executing %s command: %v\n", 
				command.name, 
				err,
			)
			continue
		}
	}
}