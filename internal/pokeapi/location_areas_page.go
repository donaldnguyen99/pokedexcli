package pokeapi

import (
	"fmt"
)

const (
	locationAreasPageBaseURL = "https://pokeapi.co/api/v2/location-area"
	locationAreasPageOffset  = 0
	locationAreasPageLimit   = 20
)

func GetLocationAreasPageBaseURL() string {
	return locationAreasPageBaseURL
}

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

type LocationAreasConfig struct {
	BaseURL string
	Offset  int
	Limit   int
	api     *PokeAPIWrapper
}

func NewLocationAreasConfig(api *PokeAPIWrapper) *LocationAreasConfig {
	return &LocationAreasConfig{
		BaseURL: locationAreasPageBaseURL,
		Offset:  locationAreasPageOffset,
		Limit:   locationAreasPageLimit,
		api:     api,
	}
}

func (m *LocationAreasConfig) GetLocationAreasPageURL() string {
	return fmt.Sprintf("%s?offset=%d&limit=%d", m.BaseURL, m.Offset, m.Limit)
}
