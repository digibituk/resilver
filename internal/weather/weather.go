package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type WeatherData struct {
	Temperature  float64 `json:"temperature"`
	ApparentTemp float64 `json:"apparentTemperature"`
	Humidity     int     `json:"humidity"`
	WeatherCode  int     `json:"weatherCode"`
	WindSpeed    float64 `json:"windSpeed"`
	Description  string  `json:"description"`
	Icon         string  `json:"icon"`
	IsDay        bool    `json:"isDay"`
}

type openMeteoResponse struct {
	Current struct {
		Temperature  float64 `json:"temperature_2m"`
		Humidity     int     `json:"relative_humidity_2m"`
		ApparentTemp float64 `json:"apparent_temperature"`
		WeatherCode  int     `json:"weather_code"`
		WindSpeed    float64 `json:"wind_speed_10m"`
		IsDay        int     `json:"is_day"`
	} `json:"current"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Fetch(lat, lon float64, units string) (WeatherData, error) {
	tempUnit := "celsius"
	windUnit := "kmh"
	if units == "fahrenheit" {
		tempUnit = "fahrenheit"
		windUnit = "mph"
	}

	url := fmt.Sprintf(
		"%s/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m,relative_humidity_2m,apparent_temperature,weather_code,wind_speed_10m,is_day&temperature_unit=%s&wind_speed_unit=%s",
		c.baseURL, lat, lon, tempUnit, windUnit,
	)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return WeatherData{}, fmt.Errorf("weather request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return WeatherData{}, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return WeatherData{}, fmt.Errorf("failed to read weather response: %w", err)
	}

	var raw openMeteoResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return WeatherData{}, fmt.Errorf("failed to parse weather response: %w", err)
	}

	isDay := raw.Current.IsDay == 1
	desc, icon := describeWeatherCode(raw.Current.WeatherCode, isDay)

	return WeatherData{
		Temperature:  raw.Current.Temperature,
		ApparentTemp: raw.Current.ApparentTemp,
		Humidity:     raw.Current.Humidity,
		WeatherCode:  raw.Current.WeatherCode,
		WindSpeed:    raw.Current.WindSpeed,
		Description:  desc,
		Icon:         icon,
		IsDay:        isDay,
	}, nil
}

type CachedClient struct {
	client *Client
	ttl    time.Duration
	mu     sync.RWMutex
	cached *WeatherData
	expiry time.Time
}

func NewCachedClient(baseURL string, ttl time.Duration) *CachedClient {
	return &CachedClient{
		client: NewClient(baseURL),
		ttl:    ttl,
	}
}

func (c *CachedClient) Fetch(lat, lon float64, units string) (WeatherData, error) {
	c.mu.RLock()
	if c.cached != nil && time.Now().Before(c.expiry) {
		data := *c.cached
		c.mu.RUnlock()
		return data, nil
	}
	c.mu.RUnlock()

	data, err := c.client.Fetch(lat, lon, units)
	if err != nil {
		return WeatherData{}, err
	}

	c.mu.Lock()
	c.cached = &data
	c.expiry = time.Now().Add(c.ttl)
	c.mu.Unlock()

	return data, nil
}

func describeWeatherCode(code int, isDay bool) (string, string) {
	type entry struct {
		desc string
		day  string
		night string
	}

	table := map[int]entry{
		0:  {"Clear sky", "day-sunny", "night-clear"},
		1:  {"Mainly clear", "day-cloudy", "night-alt-cloudy"},
		2:  {"Partly cloudy", "day-cloudy", "night-alt-cloudy"},
		3:  {"Overcast", "day-sunny-overcast", "night-alt-partly-cloudy"},
		45: {"Foggy", "day-fog", "night-fog"},
		48: {"Foggy", "day-fog", "night-fog"},
		51: {"Drizzle", "day-sprinkle", "night-sprinkle"},
		53: {"Drizzle", "day-showers", "night-showers"},
		55: {"Drizzle", "day-showers", "night-showers"},
		56: {"Freezing drizzle", "snowflake-cold", "snowflake-cold"},
		57: {"Freezing drizzle", "snowflake-cold", "snowflake-cold"},
		61: {"Light rain", "day-sprinkle", "night-sprinkle"},
		63: {"Moderate rain", "day-showers", "night-showers"},
		65: {"Heavy rain", "day-thunderstorm", "night-thunderstorm"},
		66: {"Freezing rain", "day-rain-mix", "night-rain-mix"},
		67: {"Freezing rain", "day-snow-thunderstorm", "night-snow-thunderstorm"},
		71: {"Light snow", "day-snow-wind", "night-snow-wind"},
		73: {"Moderate snow", "day-snow-wind", "night-snow-wind"},
		75: {"Heavy snow", "day-snow-thunderstorm", "night-snow-thunderstorm"},
		77: {"Snow grains", "day-sleet", "night-sleet"},
		80: {"Rain showers", "day-sprinkle", "night-sprinkle"},
		81: {"Rain showers", "day-showers", "night-showers"},
		82: {"Rain showers", "day-thunderstorm", "night-thunderstorm"},
		85: {"Snow showers", "day-rain-mix", "night-rain-mix"},
		86: {"Snow showers", "day-rain-mix", "night-rain-mix"},
		95: {"Thunderstorm", "day-thunderstorm", "night-thunderstorm"},
		96: {"Thunderstorm with hail", "day-sleet", "night-sleet"},
		99: {"Thunderstorm with hail", "day-sleet-storm", "night-sleet-storm"},
	}

	e, ok := table[code]
	if !ok {
		return "Unknown", "cloud"
	}
	if isDay {
		return e.desc, e.day
	}
	return e.desc, e.night
}
