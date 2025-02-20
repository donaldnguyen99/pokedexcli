package pokeapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/donaldnguyen99/pokedexcli/internal/pokecache"
)

type PokeAPIWrapper struct {
	MapConfig   config
	Cache       *pokecache.Cache
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
		Cache: pokecache.NewCache(cacheInterval),
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