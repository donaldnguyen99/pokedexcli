package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

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

func commandExit(...string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(...string) error {
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
		if commands[mapCommand].api.MapConfig.Next == "" {
			fmt.Println("you're on the last page")
			return nil
		}
		fullURL = commands[mapCommand].api.MapConfig.Next
	} else {
		mapCommand = "mapb"
		if commands[mapCommand].api.MapConfig.Previous == "" {
			fmt.Println("you're on the first page")
			return nil
		}
		fullURL = commands[mapCommand].api.MapConfig.Previous
	}

	locationAreasPage, err := pokeAPIWrapper.GetNamedAPIResourceList(
		fullURL,
	)
	if err != nil {
		return fmt.Errorf("error getting location areas page: %v", err)
	}

	commands[mapCommand].api.MapConfig.Next = locationAreasPage.Next
	commands[mapCommand].api.MapConfig.Previous = locationAreasPage.Previous

	for _, location := range locationAreasPage.Results {
		// urlSplit := strings.Split(location.URL, "/")
		// id := urlSplit[len(urlSplit)-2]
		fmt.Printf("%s\n", location.Name)
	}
	return nil
}

func commandMap(...string) error {
	return commandMapNextPage(true)
}

func commandMapb(...string) error {
	return commandMapNextPage(false)
}

func commandExplore(params ...string) error {
	fullURL := pokeapi.GetLocationAreaURLByName(params[0])
	locationArea, err := pokeAPIWrapper.GetLocationArea(fullURL)
	fmt.Printf("Exploring %s...\n", locationArea.Name)
	if err != nil {
		return fmt.Errorf("error getting location area: %v", err)
	}
	fmt.Println("Found Pokemon:")
	for _, encounter := range locationArea.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}
	return nil
}

func verifyCallbackParams(commmand string, params []string) error {
	switch commmand {
	case "help":
		fallthrough
	case "exit":
		fallthrough
	case "map":
		fallthrough
	case "mapb":
		if len(params) > 0 {
			return fmt.Errorf("help %s does not take any arguments", commmand)
		}
	case "explore":
		if len(params) != 1{
			return fmt.Errorf("explore %s requires 1 argument", commmand)
		}
	default:
		return fmt.Errorf("invalid command %s", commmand)
	}
	return nil
}

type cliCommand struct {
	name        string
	description string
	callback    func(...string) error
	callbackParams []string
	api      *pokeapi.PokeAPIWrapper
}


// These globals aren't ideal, but they'll do for now.
var commands map[string]cliCommand
var pokeAPIWrapper *pokeapi.PokeAPIWrapper

func main() {
	
	scanner :=bufio.NewScanner(os.Stdin)
	pokeAPIWrapper = pokeapi.NewPokeAPIWrapper(5 * time.Second)
	locationAreasConfig := pokeapi.NewLocationAreasConfig(pokeAPIWrapper) // defaults to first page of first 20 locations
	pokeAPIWrapper.MapConfig.Next = locationAreasConfig.GetLocationAreasPageURL()
	commands = map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
			callbackParams: nil,
			api:      nil,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
			callbackParams: nil,
			api:      nil,
		},
		"map": {
			name:        "map",
			description: "Displays the names of 20 locations in the Pokemon world or the next 20 locations.",
			callback:    commandMap,
			callbackParams: nil,
			api:         pokeAPIWrapper,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the names of previous 20 locations in the Pokemon world.",
			callback:    commandMapb,
			callbackParams: nil,
			api:         pokeAPIWrapper,
		},
		"explore": {
			name:        "explore",
			description: "Displays the names of the Pokemon in a specified location area.",
			callback:    commandExplore,
			callbackParams: []string{},
			api:         pokeAPIWrapper,
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
		if len(words) > 1 {
			command.callbackParams = words[1:]
			err := verifyCallbackParams(command.name, command.callbackParams)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}

			err = command.callback(command.callbackParams...)
			if err != nil {
				fmt.Printf(
					"Error while executing %s command with arguments %s: %v\n", 
					command.name, 
					strings.Join(words[1:], " "), 
					err,
				)
				continue
			}
		} else {
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
}