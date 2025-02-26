package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/donaldnguyen99/pokedexcli/internal/pokeapi"
	"golang.org/x/term"
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

func commandExit(terminal *term.Terminal, params ...string) error {
	fmt.Fprintln(terminal, "Closing the Pokedex... Goodbye!")
	return io.EOF
}

func commandHelp(terminal *term.Terminal, params ...string) error {
	if len(commands) == 0 {
		return fmt.Errorf("no commands available")
	}
	fmt.Fprintln(terminal, "Welcome to the Pokedex!")
	fmt.Fprintln(terminal, "Usage:")
	fmt.Fprintln(terminal, "")
	for _, command := range commands {
		fmt.Fprintf(terminal, "%s: %s\n", command.name, command.description)
	}
	return nil
}

func commandMapNextPage(terminal *term.Terminal, goToNextPage bool) error {
	var fullURL string
	var mapCommand string
	if goToNextPage {
		mapCommand = "map"
		if commands[mapCommand].api.MapConfig.Next == "" {
			fmt.Fprintln(terminal, "you're on the last page")
			return nil
		}
		fullURL = commands[mapCommand].api.MapConfig.Next
	} else {
		mapCommand = "mapb"
		if commands[mapCommand].api.MapConfig.Previous == "" {
			fmt.Fprintln(terminal, "you're on the first page")
			return nil
		}
		fullURL = commands[mapCommand].api.MapConfig.Previous
	}

	locationAreasPage, err := commands[mapCommand].api.GetNamedAPIResourceList(
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
		fmt.Fprintf(terminal, "%s\n", location.Name)
	}
	return nil
}

func commandMap(terminal *term.Terminal, params ...string) error {
	return commandMapNextPage(terminal, true)
}

func commandMapb(terminal *term.Terminal, params ...string) error {
	return commandMapNextPage(terminal, false)
}

func commandExplore(terminal *term.Terminal, params ...string) error {
	fullURL := pokeapi.GetLocationAreaURLByName(params[0])
	locationArea, err := commands["explore"].api.GetLocationArea(fullURL)
	if err != nil {
		return fmt.Errorf("error getting location area: %v", err)
	}
	fmt.Fprintf(terminal, "Exploring %s...\n", locationArea.Name)
	fmt.Fprintln(terminal, "Found Pokemon:")
	for _, encounter := range locationArea.PokemonEncounters {
		fmt.Fprintf(terminal, " - %s\n", encounter.Pokemon.Name)
	}
	return nil
}

func commandCatch(terminal *term.Terminal, params ...string) error {
	fullURL := pokeapi.GetPokemonURLByName(params[0])
	fmt.Fprintf(terminal, "Throwing a Pokeball at %s...\n", params[0])
	pokemon, err := commands["catch"].api.GetPokemon(fullURL)
	if err != nil {
		return fmt.Errorf("error getting pokemon: %v", err)
	}

	randInt := rand.Intn(1000)
	pokemonCatchRate := (pokemon.BaseExperience-36)*600/(635-36+1) + 400
	if randInt > pokemonCatchRate { // 36 - 608
		fmt.Fprintf(terminal, "%s was caught!\n", pokemon.Name)
		commands["catch"].api.CaughtPokemons[pokemon.Name] = pokemon
	} else {
		fmt.Fprintf(terminal, "%s escaped!\n", pokemon.Name)
	}
	return nil
}

func commandInspect(terminal *term.Terminal, params ...string) error {
	pokemon, ok := commands["inspect"].api.CaughtPokemons[params[0]]
	if !ok {
		fmt.Fprintln(terminal, "you have not caught that pokemon")
		return nil
	}
	fmt.Fprintf(terminal, "Name: %s\n", pokemon.Name)
	fmt.Fprintf(terminal, "Height: %d\n", pokemon.Height)
	fmt.Fprintf(terminal, "Weight: %d\n", pokemon.Weight)
	fmt.Fprintln(terminal, "Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Fprintf(terminal, "  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Fprintln(terminal, "Types:")
	for _, types := range pokemon.Types {
		fmt.Fprintf(terminal, "  - %s\n", types.Type.Name)
	}
	return nil
}

func commandPokedex(terminal *term.Terminal, params ...string) error {
	fmt.Fprintln(terminal, "Your pokedex:")
	for _, pokemon := range commands["pokedex"].api.CaughtPokemons {
		fmt.Fprintf(terminal, " - %s\n", pokemon.Name)
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
		fallthrough
	case "pokedex":
		if len(params) > 0 {
			return fmt.Errorf("%s does not take any arguments", commmand)
		}
	case "explore":
		fallthrough
	case "catch":
		fallthrough
	case "inspect":
		if len(params) != 1 {
			return fmt.Errorf("%s requires 1 argument", commmand)
		}
	default:
		return fmt.Errorf("invalid command %s", commmand)
	}
	return nil
}

type cliCommand struct {
	name           string
	description    string
	callback       func(*term.Terminal, ...string) error
	callbackParams []string
	api            *pokeapi.PokeAPIWrapper
}

// These globals aren't ideal, but they'll do for now.
var commands map[string]cliCommand
var pokeAPIWrapper *pokeapi.PokeAPIWrapper

func main() {
	err := repl()
	if err != nil {
		panic(err)
	}
	fmt.Print("")
}

func repl() error {

	pokeAPIWrapper = pokeapi.NewPokeAPIWrapper(5 * time.Second)
	locationAreasConfig := pokeapi.NewLocationAreasConfig(pokeAPIWrapper) // defaults to first page of first 20 locations
	pokeAPIWrapper.MapConfig.Next = locationAreasConfig.GetLocationAreasPageURL()
	commands = map[string]cliCommand{
		"help": {
			name:           "help",
			description:    "Displays a help message",
			callback:       commandHelp,
			callbackParams: nil,
			api:            nil,
		},
		"exit": {
			name:           "exit",
			description:    "Exit the Pokedex",
			callback:       commandExit,
			callbackParams: nil,
			api:            nil,
		},
		"map": {
			name:           "map",
			description:    "Displays the names of 20 locations in the Pokemon world or the next 20 locations.",
			callback:       commandMap,
			callbackParams: nil,
			api:            pokeAPIWrapper,
		},
		"mapb": {
			name:           "mapb",
			description:    "Displays the names of previous 20 locations in the Pokemon world.",
			callback:       commandMapb,
			callbackParams: nil,
			api:            pokeAPIWrapper,
		},
		"explore": {
			name:           "explore",
			description:    "Displays the names of the Pokemon in a specified location area.",
			callback:       commandExplore,
			callbackParams: []string{},
			api:            pokeAPIWrapper,
		},
		"catch": {
			name:           "catch",
			description:    "Catches a Pokemon in a current location area.",
			callback:       commandCatch,
			callbackParams: []string{},
			api:            pokeAPIWrapper,
		},
		"inspect": {
			name:           "inspect",
			description:    "Displays the details of a caught Pokemon.",
			callback:       commandInspect,
			callbackParams: []string{},
			api:            pokeAPIWrapper,
		},
		"pokedex": {
			name:           "pokedex",
			description:    "Displays the names of all the Pokemon in your Pokedex.",
			callback:       commandPokedex,
			callbackParams: nil,
			api:            pokeAPIWrapper,
		},
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return fmt.Errorf("stdin/stdout should be term")
	}
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	screen := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}
	terminal := term.NewTerminal(screen, "")
	terminal.SetPrompt(string(terminal.Escape.Red) + "Pokedex > " + string(terminal.Escape.Reset))

	// Start REPL
	for {

		text, err := terminal.ReadLine()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if text == "" {
			continue
		}
		words := cleanInput(text)

		if len(words) == 0 {
			continue
		}
		command, ok := commands[words[0]]
		if !ok {
			fmt.Fprintln(terminal, "Invalid command. Please try again.")
			continue
		}
		command.callbackParams = words[1:]
		err = verifyCallbackParams(command.name, command.callbackParams)
		if err != nil {
			fmt.Fprintf(terminal, "Error: %v\n", err)
			continue
		}

		if len(command.callbackParams) > 0 {
			err = command.callback(terminal, command.callbackParams...)
			if err != nil {
				fmt.Fprintf(terminal,
					"Error while executing %s command with arguments %s: %v\n",
					command.name,
					strings.Join(words[1:], " "),
					err,
				)
				continue
			}
		} else {
			err = command.callback(terminal)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				fmt.Fprintf(terminal,
					"Error while executing %s command: %v\n",
					command.name,
					err,
				)
				continue
			}
		}
	}
}
