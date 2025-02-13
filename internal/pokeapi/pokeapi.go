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

const (
	locationAreasPageBaseURL = "https://pokeapi.co/api/v2/location-area"
	locationAreasPageOffset  = 0
	locationAreasPageLimit   = 20
)

func GetLocationAreasPageURL(baseURL string, offset, limit int) string {
	return fmt.Sprintf("%s?offset=%d&limit=%d", baseURL, offset, limit)
}

func GetLocationAreasPageDefaultURL() string {
	return GetLocationAreasPageURL(
		locationAreasPageBaseURL,
		locationAreasPageOffset,
		locationAreasPageLimit,
	)
}

type LocationAreasPage struct {
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type LocationAreasManager struct {
	BaseURL string
	Offset  int
	Limit   int
	Cache   *pokecache.Cache
}

func NewLocationAreasManager() *LocationAreasManager {
	return &LocationAreasManager{
		BaseURL: locationAreasPageBaseURL,
		Offset:  locationAreasPageOffset,
		Limit:   locationAreasPageLimit,
		Cache:   pokecache.NewCache(5 * time.Minute),
	}
}

func (m *LocationAreasManager) GetLocationAreasPageURL() string {
	return fmt.Sprintf("%s?offset=%d&limit=%d", m.BaseURL, m.Offset, m.Limit)
}

func (m *LocationAreasManager) GetLocationAreasPage(
	fullURL string,
) (LocationAreasPage, error) {

	cachedData, ok := m.Cache.Get(fullURL)
	if ok {
		var locationArea LocationAreasPage
		err := json.Unmarshal(cachedData, &locationArea)
		if err != nil {
			return LocationAreasPage{}, fmt.Errorf("Error unmarshalling cached data: \n%v", err)
		}
		return locationArea, nil
	}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return LocationAreasPage{}, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return LocationAreasPage{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return LocationAreasPage{}, fmt.Errorf(
			"Unexpected status code: %d", 
			resp.StatusCode,
		)
	}

	var dataToCache []byte
	// Read the response body into a byte slice
	dataToCache, err = io.ReadAll(resp.Body)
	if err != nil {
		return LocationAreasPage{}, fmt.Errorf("Error reading response body: \n%v", err)
	}
	m.Cache.Add(fullURL, dataToCache)


	var locationArea LocationAreasPage
	decoder := json.NewDecoder(bytes.NewReader(dataToCache))
	if err := decoder.Decode(&locationArea); err != nil {
		return LocationAreasPage{}, fmt.Errorf("Error decoding JSON: \n%v", err)
	}
	return locationArea, nil
}
