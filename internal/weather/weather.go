package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// WeatherData represents parsed weather information
type WeatherData struct {
	Location    string
	Temperature string
	Condition   string
	Icon        string
	Humidity    string
	Wind        string
	UpdatedAt   time.Time
}

// DayForecast represents weather for a specific day
type DayForecast struct {
	Date        time.Time
	Icon        string
	MaxTempC    string
	MinTempC    string
	MaxTempF    string
	MinTempF    string
	WeatherCode string
	Condition   string
}

// WeatherForecast represents multi-day weather forecast
type WeatherForecast struct {
	Location  string
	Days      []DayForecast
	UpdatedAt time.Time
}

// WttrResponse represents the JSON response from wttr.in
type WttrResponse struct {
	CurrentCondition []CurrentCondition `json:"current_condition"`
	Weather          []DayWeather       `json:"weather"`
}

type CurrentCondition struct {
	TempC       string `json:"temp_C"`
	TempF       string `json:"temp_F"`
	Humidity    string `json:"humidity"`
	WindSpeedKmph string `json:"windspeedKmph"`
	WindDir     string `json:"winddirDegree"`
	WeatherCode string `json:"weatherCode"`
	WeatherDesc []struct {
		Value string `json:"value"`
	} `json:"weatherDesc"`
}

type DayWeather struct {
	Date     string `json:"date"`
	MaxTempC string `json:"maxtempC"`
	MinTempC string `json:"mintempC"`
	MaxTempF string `json:"maxtempF"`
	MinTempF string `json:"mintempF"`
	Hourly   []struct {
		Time        string `json:"time"`
		TempC       string `json:"tempC"`
		TempF       string `json:"tempF"`
		WeatherCode string `json:"weatherCode"`
		WeatherDesc []struct {
			Value string `json:"value"`
		} `json:"weatherDesc"`
	} `json:"hourly"`
}

// WeatherCache provides caching for weather data
type WeatherCache struct {
	data         map[string]WeatherData
	forecasts    map[string]WeatherForecast
	lastFetch    map[string]time.Time
	lastForecast map[string]time.Time
	cacheTTL     time.Duration
}

// NewWeatherCache creates a new weather cache with 2-hour TTL
func NewWeatherCache() *WeatherCache {
	return &WeatherCache{
		data:         make(map[string]WeatherData),
		forecasts:    make(map[string]WeatherForecast),
		lastFetch:    make(map[string]time.Time),
		lastForecast: make(map[string]time.Time),
		cacheTTL:     2 * time.Hour,
	}
}

// GetWeatherData fetches weather data for a location with caching
func (wc *WeatherCache) GetWeatherData(location string) (*WeatherData, error) {
	if location == "" {
		return nil, fmt.Errorf("no location specified")
	}

	// Check cache first
	cacheKey := strings.ToLower(location)
	if lastFetch, exists := wc.lastFetch[cacheKey]; exists {
		if time.Since(lastFetch) < wc.cacheTTL {
			if cachedData, exists := wc.data[cacheKey]; exists {
				return &cachedData, nil
			}
		}
	}

	// Fetch fresh data
	weatherData, err := fetchWeatherData(location)
	if err != nil {
		return nil, err
	}

	// Update cache
	wc.data[cacheKey] = *weatherData
	wc.lastFetch[cacheKey] = time.Now()

	return weatherData, nil
}

// GetWeatherForecast fetches 3-day weather forecast for a location with caching
func (wc *WeatherCache) GetWeatherForecast(location string) (*WeatherForecast, error) {
	if location == "" {
		return nil, fmt.Errorf("no location specified")
	}

	// Check cache first
	cacheKey := strings.ToLower(location)
	if lastForecast, exists := wc.lastForecast[cacheKey]; exists {
		if time.Since(lastForecast) < wc.cacheTTL {
			if cachedForecast, exists := wc.forecasts[cacheKey]; exists {
				return &cachedForecast, nil
			}
		}
	}

	// Fetch fresh forecast data
	forecast, err := fetchWeatherForecast(location)
	if err != nil {
		return nil, err
	}

	// Update cache
	wc.forecasts[cacheKey] = *forecast
	wc.lastForecast[cacheKey] = time.Now()

	return forecast, nil
}

// fetchWeatherData fetches weather data from wttr.in API
func fetchWeatherData(location string) (*WeatherData, error) {
	// URL encode the location
	encodedLocation := url.QueryEscape(location)
	
	// Use JSON format for detailed data
	url := fmt.Sprintf("https://wttr.in/%s?format=j1", encodedLocation)
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}
	
	var wttrResp WttrResponse
	if err := json.NewDecoder(resp.Body).Decode(&wttrResp); err != nil {
		return nil, fmt.Errorf("failed to decode weather response: %w", err)
	}
	
	// Parse the response into our WeatherData structure
	weatherData := &WeatherData{
		Location:  location,
		UpdatedAt: time.Now(),
	}
	
	// Extract current conditions
	if len(wttrResp.CurrentCondition) > 0 {
		current := wttrResp.CurrentCondition[0]
		weatherData.Temperature = current.TempC + "°C"
		weatherData.Humidity = current.Humidity + "%"
		weatherData.Wind = current.WindSpeedKmph + "km/h " + current.WindDir
		
		if len(current.WeatherDesc) > 0 {
			weatherData.Condition = current.WeatherDesc[0].Value
		}
		
		// Convert weather code to icon
		weatherData.Icon = getWeatherIcon(current.WeatherCode)
	}
	
	return weatherData, nil
}

// GetSimpleWeatherData fetches simple weather data using format string
func GetSimpleWeatherData(location string) (string, error) {
	if location == "" {
		return "", fmt.Errorf("no location specified")
	}
	
	// Use format 3 for simple "Location: 🌦 +11°C" output
	encodedLocation := url.QueryEscape(location)
	url := fmt.Sprintf("https://wttr.in/%s?format=3", encodedLocation)
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}
	
	buf := make([]byte, 1024)
	n, err := resp.Body.Read(buf)
	if err != nil && n == 0 {
		return "", fmt.Errorf("failed to read weather response: %w", err)
	}
	
	return strings.TrimSpace(string(buf[:n])), nil
}

// getWeatherIcon converts weather code to simple, single-width emoji icon
func getWeatherIcon(code string) string {
	switch code {
	case "113":
		return "☀"  // Clear/Sunny (no variation selector)
	case "116":
		return "⛅"  // Partly cloudy (single-width)
	case "119", "122":
		return "☁"  // Cloudy/Overcast (no variation selector)
	case "143":
		return "≡"  // Mist/Fog (simple horizontal lines)
	case "176", "263", "266":
		return "🌦"  // Light rain (no variation selector)
	case "179", "182", "185":
		return "❄"  // Light snow (no variation selector)
	case "200":
		return "⚡"  // Thundery outbreaks (lightning bolt)
	case "227":
		return "❄"  // Snow (no variation selector)
	case "230":
		return "❅"  // Blizzard (snowflake without variation selector)
	case "248":
		return "≡"  // Fog (simple horizontal lines)
	case "260":
		return "≡"  // Freezing fog (simple horizontal lines)
	case "293", "296", "299", "302", "305", "308", "311", "314", "317", "320", "323":
		return "🌧"  // Rain (no variation selector)
	case "326", "329", "332", "335", "338", "350", "353", "356", "359", "362", "365", "368", "371", "374", "377", "386", "389", "392", "395":
		return "❄"  // Snow (no variation selector)
	default:
		return "⛅"  // Default partly cloudy (single-width)
	}
}

// fetchWeatherForecast fetches 3-day weather forecast from wttr.in API
func fetchWeatherForecast(location string) (*WeatherForecast, error) {
	// URL encode the location
	encodedLocation := url.QueryEscape(location)
	
	// Use JSON format for detailed forecast data
	url := fmt.Sprintf("https://wttr.in/%s?format=j1", encodedLocation)
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather forecast: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}
	
	var wttrResp WttrResponse
	if err := json.NewDecoder(resp.Body).Decode(&wttrResp); err != nil {
		return nil, fmt.Errorf("failed to decode weather response: %w", err)
	}
	
	// Parse the response into our WeatherForecast structure
	forecast := &WeatherForecast{
		Location:  location,
		UpdatedAt: time.Now(),
		Days:      make([]DayForecast, 0, len(wttrResp.Weather)),
	}
	
	// Extract forecast for each day
	for _, dayWeather := range wttrResp.Weather {
		date, err := time.Parse("2006-01-02", dayWeather.Date)
		if err != nil {
			continue // Skip invalid dates
		}
		
		dayForecast := DayForecast{
			Date:     date,
			MaxTempC: dayWeather.MaxTempC,
			MinTempC: dayWeather.MinTempC,
			MaxTempF: dayWeather.MaxTempF,
			MinTempF: dayWeather.MinTempF,
		}
		
		// Get weather condition from first hourly entry (usually most representative)
		if len(dayWeather.Hourly) > 0 {
			hourly := dayWeather.Hourly[0]
			dayForecast.WeatherCode = hourly.WeatherCode
			dayForecast.Icon = getWeatherIcon(hourly.WeatherCode)
			
			if len(hourly.WeatherDesc) > 0 {
				dayForecast.Condition = hourly.WeatherDesc[0].Value
			}
		}
		
		forecast.Days = append(forecast.Days, dayForecast)
	}
	
	return forecast, nil
}