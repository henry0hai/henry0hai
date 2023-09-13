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
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		log.Fatal("WEATHER_API_KEY not set in .env")
	}

	baseUrl := "https://api.weatherapi.com/v1/current.json"
	queryParams := url.Values{
		"key": []string{apiKey},
		"q":   []string{"Ho Chi Minh City"},
		"aqi": []string{"yes"},
	}

	resp, err := http.Get(baseUrl + "?" + queryParams.Encode())
	if err != nil {
		log.Fatal("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	var weather WeatherResponse
	err = json.Unmarshal(body, &weather)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	templateData, err := os.ReadFile("README.md.template")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	template := string(templateData)

	startIndex := strings.Index(template, "<!-- WEATHER:START -->")
	endIndex := strings.Index(template, "<!-- WEATHER:END -->")
	if startIndex == -1 || endIndex == -1 {
		fmt.Println("Cannot find weather section in README.md.template")
		os.Exit(1)
	}

	readme := template[:startIndex+len("<!-- WEATHER:START -->")] +
		fmt.Sprintf(
			`Current City: %s - %s

Condition: %s, <img src="https:%s"/>

Current temperature: %.2f °C, Feels like: %.2f °C, Humidity: %d%%

Wind: %.2f km/h, %d°, %s

Pressure: %.2f mb

Updated at: %s`,
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
			weather.Current.LastUpdated,
		) +
		template[endIndex:]

	err = os.WriteFile("README.md", []byte(readme), 0644)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
