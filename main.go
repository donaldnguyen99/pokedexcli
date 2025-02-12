package main

import (
	"bufio"
	"fmt"
	"encoding/json"
	"net/http"
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

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var locationArea struct {
		Next     string     `json:"next"`
		Previous string     `json:"previous"`
		Results  []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"results"`
	}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&locationArea); err != nil {
		return err
	}
	commands[mapCommand].config.Next = locationArea.Next
	commands[mapCommand].config.Previous = locationArea.Previous

	for _, location := range locationArea.Results {
		urlSplit := strings.Split(location.URL, "/")
		id := urlSplit[len(urlSplit)-2]
		fmt.Printf("%s. %s\n", id, location.Name)
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

var commands map[string]cliCommand

func main() {
	
	scanner :=bufio.NewScanner(os.Stdin)
	map_config := &config{
		Next:     "https://pokeapi.co/api/v2/location-area?offset=0&limit=20",
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
	fmt.Printf("Pokedex > ")
	for ; scanner.Scan(); fmt.Printf("Pokedex > ") {
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
			fmt.Printf("Error while executing %s command: %v\n", command.name, err)
			continue
		}
	}
}