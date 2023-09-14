package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type WeatherResponse struct {
	Location struct {
		Name string `json:"name"`
		TzID string `json:"tz_id"`
	} `json:"location"`
	Current struct {
		TempC       float64 `json:"temp_c"`
		FeelslikeC  float64 `json:"feelslike_c"`
		Humidity    int     `json:"humidity"`
		WindKph     float64 `json:"wind_kph"`
		WindDegree  int     `json:"wind_degree"`
		WindDir     string  `json:"wind_dir"`
		PressureMb  float64 `json:"pressure_mb"`
		LastUpdated string  `json:"last_updated"`
		Condition   struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
		} `json:"condition"`
	} `json:"current"`
}

func main() {
	apiKey := getEnvVar("WEATHER_API_KEY")
	baseUrl := "https://api.weatherapi.com/v1/current.json"

	queryParams := url.Values{
		"key": []string{apiKey},
		"q":   []string{"Ho Chi Minh City"},
		"aqi": []string{"yes"},
	}

	resp, err := http.Get(baseUrl + "?" + queryParams.Encode())
	if err != nil {
		log.Fatalf("Error making the request: %v", err)
	}
	defer resp.Body.Close()

	var weather WeatherResponse
	err = json.Unmarshal(readBody(resp), &weather)
	if err != nil {
		log.Fatalf("Error unmarshal the response: %v", err)
	}

	templateData, err := os.ReadFile("README.md.template")
	if err != nil {
		log.Fatalf("Error reading the template file: %v", err)
	}

	updateReadme(string(templateData), weather)
}

func getEnvVar(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s not set in .env", key)
	}
	return value
}

func readBody(resp *http.Response) []byte {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading the response body: %v", err)
	}
	return body
}

func updateReadme(template string, weather WeatherResponse) {
	startIndex := strings.Index(template, "<!-- WEATHER:START -->")
	endIndex := strings.Index(template, "<!-- WEATHER:END -->")

	if startIndex == -1 || endIndex == -1 {
		log.Fatal("Error find weather section in README.md.template")
	}

	location, _ := time.LoadLocation(weather.Location.TzID)

	layout := "2006-01-02 15:04"
	lastUpdated, _ := time.Parse(layout, weather.Current.LastUpdated)

	// Get the timezone offset in seconds.
	_, offset  := lastUpdated.In(location).Zone() 

	// Convert to hour
	offsetHours := offset / 3600

	formattedTime := fmt.Sprintf("%s (GMT%+02d)", weather.Current.LastUpdated, offsetHours)

	fmt.Printf("formattedTime: %v\n", formattedTime)

	readme := template[:startIndex+len("<!-- WEATHER:START -->")] +
		fmt.Sprintf(
			generateWeatherString(),
			weather.Location.Name,
			time.Now().Format("02/01/2006"),
			weather.Current.Condition.Text,
			weather.Current.Condition.Icon,
			weather.Current.TempC,
			weather.Current.FeelslikeC,
			weather.Current.Humidity,
			weather.Current.WindKph,
			weather.Current.WindDegree,
			weather.Current.WindDir,
			weather.Current.PressureMb,
			formattedTime,
		) +
		template[endIndex:]

	err := os.WriteFile("README.md", []byte(readme), 0644)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println("README.md updated successfully!")
}

func generateWeatherString() string {
	return `
**Current City**: %s - *%s*

**Condition**: %s, <img src="https:%s"/>

**Current temperature**: %.2f °C, **Feels like**: %.2f °C, **Humidity**: %d%%

**Wind**: %.2f km/h, %d°, *%s*

**Pressure**: %.2f mb

**Updated at**: %s`
}
