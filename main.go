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

const CURRENT_WEATHER_START_TEMPLATE = `<!-- CURRENT_WEATHER:START -->`
const CURRENT_WEATHER_END_TEMPLATE = `<!-- CURRENT_WEATHER:END -->`

const FORECAST_WEATHER_START_TEMPLATE = `<!-- FORECAST_WEATHER:START -->`
const FORECAST_WEATHER_END_TEMPLATE = `<!-- FORECAST_WEATHER:END -->`

type WeatherResponse struct {
	Location struct {
		Name string `json:"name"`
		TzID string `json:"tz_id"`
	} `json:"location"`
	Current struct {
		LastUpdated string `json:"last_updated"`
		ShareAttribute
	} `json:"current"`
	Forecast struct {
		ForecastDay []ForecastDay `json:"forecastday"`
	} `json:"forecast"`
}

type ShareAttribute struct {
	TempC      float64 `json:"temp_c"`
	FeelslikeC float64 `json:"feelslike_c"`
	Humidity   int     `json:"humidity"`
	WindKph    float64 `json:"wind_kph"`
	WindDegree int     `json:"wind_degree"`
	WindDir    string  `json:"wind_dir"`
	PressureMb float64 `json:"pressure_mb"`
	UV         float64 `json:"uv"`
	Cloud      int     `json:"cloud"`
	Condition  struct {
		Text string `json:"text"`
		Icon string `json:"icon"`
	} `json:"condition"`
}

type ForecastDay struct {
	Date string `json:"date"`
	Day  struct {
		MaxtempC      float64 `json:"maxtemp_c"`
		MintempC      float64 `json:"mintemp_c"`
		AvgtempC      float64 `json:"avgtemp_c"`
		MaxwindKph    float64 `json:"maxwind_kph"`
		TotalprecipMM float64 `json:"totalprecip_mm"`
		AvgvisKM      float64 `json:"avgvis_km"`
		Avghumidity   float64 `json:"avghumidity"`
		Condition     struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
		} `json:"condition"`
	} `json:"day"`
	Astro struct {
		Sunrise          string `json:"sunrise"`
		Sunset           string `json:"sunset"`
		Moonrise         string `json:"moonrise"`
		Moonset          string `json:"moonset"`
		MoonPhase        string `json:"moon_phase"`
		MoonIllumination string `json:"moon_illumination"`
	} `json:"astro"`
	Hour []Hour `json:"hour"`
}

type Hour struct {
	TimeEpoch    int64   `json:"time_epoch"`
	Time         string  `json:"time"`
	WillItRain   int     `json:"will_it_rain"`
	ChanceOfRain float64 `json:"chance_of_rain"`
	WillItSnow   int     `json:"will_it_snow"`
	ChanceOfSnow float64 `json:"chance_of_snow"`
	VisKM        float64 `json:"vis_km"`
	GustKph      float64 `json:"gust_kph"`
	ShareAttribute
}

type AirQuality struct {
	Co           float64 `json:"co"`
	No2          float64 `json:"no2"`
	O3           float64 `json:"o3"`
	So2          float64 `json:"so2"`
	Pm25         float64 `json:"pm2_5"`
	Pm10         float64 `json:"pm10"`
	UsEpaIndex   int     `json:"us_epa_index"`
	GbDefraIndex int     `json:"gb_defra_index"`
}

func main() {
	apiKey := getEnvVar("WEATHER_API_KEY")
	baseUrl := "https://api.weatherapi.com/v1/forecast.json"

	queryParams := url.Values{
		"key":  []string{apiKey},
		"q":    []string{"Ho Chi Minh City"},
		"days": []string{"1"},
		"aqi":  []string{"yes"},
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

	updateCurrentWeather(string(templateData), weather)

	currentData, err := os.ReadFile("README.md")
	if err != nil {
		log.Fatalf("Error reading the current file: %v", err)
	}
	updateForecastWeather(string(currentData), weather)
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

func updateCurrentWeather(template string, weather WeatherResponse) {
	startIndex := strings.Index(template, CURRENT_WEATHER_START_TEMPLATE)
	endIndex := strings.Index(template, CURRENT_WEATHER_END_TEMPLATE)

	if startIndex == -1 || endIndex == -1 {
		log.Fatal("Error find weather section in README.md.template")
	}

	if weather.Location.TzID == "" || len(weather.Forecast.ForecastDay) == 0 {
		log.Fatal("Error: Location or TzID is nil or No forecast data available")
	} else {
		location, _ := time.LoadLocation(weather.Location.TzID)

		layout := "2006-01-02 15:04"
		lastUpdated, _ := time.Parse(layout, weather.Current.LastUpdated)

		// Get the timezone offset in seconds.
		_, offset := lastUpdated.In(location).Zone()

		// Convert to hour
		offsetHours := offset / 3600

		formattedTime := fmt.Sprintf("%s (GMT%+02d)", weather.Current.LastUpdated, offsetHours)

		fmt.Printf("formattedTime: %v\n", formattedTime)

		readme := template[:startIndex+len(CURRENT_WEATHER_START_TEMPLATE)] +
			fmt.Sprintf(
				generateCurrentWeatherString(),
				weather.Location.Name,
				time.Now().Format("02/01/2006"),
				weather.Current.ShareAttribute.Condition.Text,
				weather.Current.ShareAttribute.Condition.Icon,
				weather.Current.ShareAttribute.TempC,
				weather.Current.ShareAttribute.FeelslikeC,
				weather.Current.ShareAttribute.Humidity,
				weather.Current.ShareAttribute.WindKph,
				weather.Current.ShareAttribute.WindDegree,
				weather.Current.ShareAttribute.WindDir,
				weather.Current.ShareAttribute.PressureMb,
				weather.Forecast.ForecastDay[0].Astro.Sunrise,
				weather.Forecast.ForecastDay[0].Astro.Sunset,
				weather.Forecast.ForecastDay[0].Astro.MoonPhase,
				weather.Forecast.ForecastDay[0].Astro.Moonrise,
				weather.Forecast.ForecastDay[0].Astro.Moonset,
				weather.Forecast.ForecastDay[0].Astro.MoonIllumination,
				formattedTime,
			) +
			template[endIndex:]

		err := os.WriteFile("README.md", []byte(readme), 0644)
		if err != nil {
			log.Fatalf("Error [updateCurrentWeather] writeFile: %v", err)
		}

		fmt.Println("README.md updated successfully for current weather!")
	}
}

// generates a string representing the current weather conditions.
func generateCurrentWeatherString() string {
	return `
**Current City**: %s - *%s*

**Condition**: %s, <img src="https:%s"/>

**Current temperature**: %.2f °C, **Feels like**: %.2f °C, **Humidity**: %d%%

**Wind**: %.2f km/h, %d°, *%s*

**Pressure**: %.2f mb

**Sunrise**: %s

**Sunset**: %s

**Moon Phase**: %s

**Moon Rise**: %s

**Moon Set**: %s

**Moon Illumination**: %s

**Updated at**: %s`
}

func getHoursAhead(hours int, weather WeatherResponse) WeatherResponse {
	lastUpdated, err := time.Parse("2006-01-02 15:04", weather.Current.LastUpdated)
	if err != nil {
		log.Fatalf("Error [getHoursAhead] time.Parse: %v", err)
	}

	hoursAheadTime := lastUpdated.Add(time.Duration(hours) * time.Hour)

	var hoursAhead WeatherResponse

	hourCount := 0

	for _, forecastDay := range weather.Forecast.ForecastDay {
		newForecastDay := ForecastDay{
			Date: forecastDay.Date,
			Hour: []Hour{},
		}

		for _, hour := range forecastDay.Hour {
			hourTime, err := time.Parse("2006-01-02 15:04", hour.Time)
			if err != nil {
				log.Fatalf("Error [getHoursAhead] time.Parse: %v", err)
			}

			if hourTime.After(lastUpdated) && hourTime.Before(hoursAheadTime) && hourCount < hours {
				newForecastDay.Hour = append(newForecastDay.Hour, hour)
				hourCount++
				fmt.Println("hourCount:", hourCount)
			}
		}

		if len(newForecastDay.Hour) > 0 {
			hoursAhead.Forecast.ForecastDay = append(hoursAhead.Forecast.ForecastDay, newForecastDay)
		}

		if hourCount == hours {
			return hoursAhead
		}
	}

	return hoursAhead
}

func updateForecastWeather(template string, weather WeatherResponse) {
	startIndex := strings.Index(template, FORECAST_WEATHER_START_TEMPLATE)
	endIndex := strings.Index(template, FORECAST_WEATHER_END_TEMPLATE)

	if startIndex == -1 || endIndex == -1 {
		log.Fatal("Error find weather section in README.md.template")
	}

	customHours := getHoursAhead(6, weather)

	if len(weather.Forecast.ForecastDay) == 0 {
		log.Fatal("Error: No forecast data available")
	}

	fmt.Printf("customHours: %v\n", len(customHours.Forecast.ForecastDay[0].Hour))

	// Generate the forecast weather table.
	table := generateForecastWeatherTable(customHours)

	readme := template[:startIndex+len(FORECAST_WEATHER_START_TEMPLATE)] +
		table +
		template[endIndex:]

	err := os.WriteFile("README.md", []byte(readme), 0644)
	if err != nil {
		log.Fatalf("Error - [updateForecastWeather] - writeFile: %v", err)
	}

	fmt.Println("README.md updated successfully for forecast weather!")

}

// func generateForecastWeatherTable(weather WeatherResponse) string {
// 	// Initialize the table with the headers
// 	table := "| Hour | Weather | Condition | Temperature | Wind | Change of Rain |\n| --- | --- | --- | --- | --- | --- |\n"

// 	for _, forecastDay := range weather.Forecast.ForecastDay {
// 		for _, hour := range forecastDay.Hour {
// 			hourTime, err := time.Parse("2006-01-02 15:04", hour.Time)
// 			if err != nil {
// 				log.Fatalf("Error: %v", err)
// 			}
// 			formattedTime := hourTime.Format("15:04")

// 			table += fmt.Sprintf("| %s | %s | <img src='https:%s'/> | %.2f °C | %.2f km/h | %v %% |\n",
// 				formattedTime,
// 				hour.ShareAttribute.Condition.Text,
// 				hour.ShareAttribute.Condition.Icon,
// 				hour.ShareAttribute.TempC,
// 				hour.ShareAttribute.WindKph,
// 				hour.ChanceOfRain,
// 			)
// 		}
// 	}

// 	return table
// }

func generateForecastWeatherTable(weather WeatherResponse) string {
	// Initialize the table with the headers
	table := `
<table>
		<tr>
			<th>Hour</th>
			<th>Weather</th>
			<th>Condition</th>
			<th>Temperature</th>
			<th>Feel Like</th>
			<th>Wind</th>
			<th>Chance of Rain</th>
		</tr>`

	for _, forecastDay := range weather.Forecast.ForecastDay {
		for _, hour := range forecastDay.Hour {
			hourTime, err := time.Parse("2006-01-02 15:04", hour.Time)
			if err != nil {
				log.Fatalf("Error [generateForecastWeatherTable] - time.Parse: %v", err)
			}
			formattedTime := hourTime.Format("15:04")

			chanceOfRain := fmt.Sprintf("%v %%", hour.ChanceOfRain)
			if hour.ChanceOfRain > 75 {
				chanceOfRain = fmt.Sprintf("🌧️ %v %%", hour.ChanceOfRain)
			}

			table += fmt.Sprintf(`
				<tr>
					<td>%s</td>
					<td>%s</td>
					<td><img src='https:%s'/></td>
					<td>%.2f °C</td>
					<td>%.2f °C</td>
					<td>%.2f km/h</td>
					<td>%s</td>
				</tr>`,
				formattedTime,
				hour.ShareAttribute.Condition.Text,
				hour.ShareAttribute.Condition.Icon,
				hour.ShareAttribute.TempC,
				hour.ShareAttribute.FeelslikeC,
				hour.ShareAttribute.WindKph,
				chanceOfRain,
			)
		}
	}

	table += "\n</table>\n"

	return table
}
