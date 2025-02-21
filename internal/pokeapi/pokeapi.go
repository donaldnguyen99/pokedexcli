package pokeapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/donaldnguyen99/pokedexcli/internal/pokecache"
)

const (
	baseURL = "https://pokeapi.co/api/v2"
)

type PokeAPIWrapper struct {
	BaseURL	    string
	MapConfig   config
	Cache       *pokecache.Cache
	CaughtPokemons map[string]Pokemon
}

type config struct {
	Next     string
	Previous string
}

type NamedAPIResourceList struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []NamedAPIResource `json:"results"`
}

type NamedAPIResource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func NewPokeAPIWrapper(cacheInterval time.Duration) *PokeAPIWrapper {
	return &PokeAPIWrapper{
		BaseURL: baseURL,
		MapConfig: config{
			Next: "",
			Previous: "",
		},
		Cache: pokecache.NewCache(cacheInterval),
		CaughtPokemons: make(map[string]Pokemon),
	}
}

func getStructFromURL[T any](fullURL string, cache *pokecache.Cache) (T, error) {
	cachedData, ok := cache.Get(fullURL)
	if ok {
		var result T
		err := json.Unmarshal(cachedData, &result)
		if err != nil {
			var noop T
			return noop, fmt.Errorf("error unmarshalling cached data: \n%v", err)
		}
		return result, nil
	}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		var noop T
		return noop, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		var noop T
		return noop, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var noop T
		return noop, fmt.Errorf(
			"unexpected status code: %d", 
			resp.StatusCode,
		)
	}

	var dataToCache []byte
	// Read the response body into a byte slice
	dataToCache, err = io.ReadAll(resp.Body)
	if err != nil {
		var noop T
		return noop, fmt.Errorf("error reading response body: \n%v", err)
	}
	cache.Add(fullURL, dataToCache)


	var result T
	decoder := json.NewDecoder(bytes.NewReader(dataToCache))
	if err := decoder.Decode(&result); err != nil {
		var noop T
		return noop, fmt.Errorf("error decoding JSON: \n%v", err)
	}
	return result, nil
}

func (p *PokeAPIWrapper) GetNamedAPIResourceList(fullURL string) (NamedAPIResourceList, error) {
	n, err := getStructFromURL[NamedAPIResourceList](fullURL, p.Cache)
	if err != nil {
		return NamedAPIResourceList{}, fmt.Errorf(
			"failed to get named API resource list from URL %s: %w", 
			fullURL, err,
		)
	}
	return n, nil
}

func (p *PokeAPIWrapper) GetLocationArea(fullURL string) (LocationArea, error) {
	l, err := getStructFromURL[LocationArea](fullURL, p.Cache)
	if err != nil {
		return LocationArea{}, fmt.Errorf(
			"failed to get location area from URL %s: %w",  fullURL, err,
		)
	}
	return l, nil
}

func (p *PokeAPIWrapper) GetPokemon(fullURL string) (Pokemon, error) {
	pokemon, err := getStructFromURL[Pokemon](fullURL, p.Cache)
	if err != nil {
		return Pokemon{}, fmt.Errorf(
			"failed to get pokemon from URL %s: %w",  fullURL, err,
		)
	}
	return pokemon, nil
}

func (p *PokeAPIWrapper) GetAllPokemon() ([]Pokemon, error) {
	pokemonList, err := p.GetNamedAPIResourceList("https://pokeapi.co/api/v2/pokemon?limit=100000&offset=0")
	if err != nil {
		return []Pokemon{}, fmt.Errorf(
			"failed to get all pokemon: %w", err,
		)
	}
	pokemons := make([]Pokemon, len(pokemonList.Results))
	for i, pokemon := range pokemonList.Results {
		pokemons[i], err = p.GetPokemon(pokemon.URL)
		if err != nil {
			return []Pokemon{}, fmt.Errorf(
				"failed to get pokemon %s: %w", pokemon.Name, err,
			)
		}
	}
	return pokemons, nil
}

func FindMinMaxBaseExperience(pokemons []Pokemon) (int, int) {
	min := math.MaxInt
	max := 0
	for _, pokemon := range pokemons {
		if pokemon.BaseExperience < min {
			min = pokemon.BaseExperience
		}
		if pokemon.BaseExperience > max {
			max = pokemon.BaseExperience
		}
	}
	return min, max
}
