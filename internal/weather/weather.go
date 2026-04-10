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
}

type openMeteoResponse struct {
	Current struct {
		Temperature  float64 `json:"temperature_2m"`
		Humidity     int     `json:"relative_humidity_2m"`
		ApparentTemp float64 `json:"apparent_temperature"`
		WeatherCode  int     `json:"weather_code"`
		WindSpeed    float64 `json:"wind_speed_10m"`
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
		"%s/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m,relative_humidity_2m,apparent_temperature,weather_code,wind_speed_10m&temperature_unit=%s&wind_speed_unit=%s",
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

	desc, icon := describeWeatherCode(raw.Current.WeatherCode)

	return WeatherData{
		Temperature:  raw.Current.Temperature,
		ApparentTemp: raw.Current.ApparentTemp,
		Humidity:     raw.Current.Humidity,
		WeatherCode:  raw.Current.WeatherCode,
		WindSpeed:    raw.Current.WindSpeed,
		Description:  desc,
		Icon:         icon,
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

func describeWeatherCode(code int) (string, string) {
	switch code {
	case 0:
		return "Clear sky", "☀️"
	case 1:
		return "Mainly clear", "🌤️"
	case 2:
		return "Partly cloudy", "⛅"
	case 3:
		return "Overcast", "☁️"
	case 45, 48:
		return "Foggy", "🌫️"
	case 51, 53, 55:
		return "Drizzle", "🌦️"
	case 56, 57:
		return "Freezing drizzle", "🌧️"
	case 61:
		return "Light rain", "🌧️"
	case 63:
		return "Moderate rain", "🌧️"
	case 65:
		return "Heavy rain", "🌧️"
	case 66, 67:
		return "Freezing rain", "🌧️"
	case 71:
		return "Light snow", "🌨️"
	case 73:
		return "Moderate snow", "🌨️"
	case 75:
		return "Heavy snow", "❄️"
	case 77:
		return "Snow grains", "🌨️"
	case 80, 81, 82:
		return "Rain showers", "🌧️"
	case 85, 86:
		return "Snow showers", "🌨️"
	case 95:
		return "Thunderstorm", "⛈️"
	case 96, 99:
		return "Thunderstorm with hail", "⛈️"
	default:
		return "Unknown", "❓"
	}
}
