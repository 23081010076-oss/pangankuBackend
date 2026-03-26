package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type WeatherHandler struct {
	client *http.Client
}

func NewWeatherHandler() *WeatherHandler {
	return &WeatherHandler{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type openMeteoResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Current   struct {
		Time              string  `json:"time"`
		Temperature2m     float64 `json:"temperature_2m"`
		RelativeHumidity2 float64 `json:"relative_humidity_2m"`
		Precipitation     float64 `json:"precipitation"`
		WeatherCode       int     `json:"weather_code"`
		WindSpeed10m      float64 `json:"wind_speed_10m"`
	} `json:"current"`
}

func (h *WeatherHandler) GetCurrentWeather(c *gin.Context) {
	lat := parseCoordOrDefault(c.Query("lat"), envOrDefaultFloat("WEATHER_DEFAULT_LAT", -7.1183))
	lng := parseCoordOrDefault(c.Query("lng"), envOrDefaultFloat("WEATHER_DEFAULT_LNG", 112.4167))
	timezone := c.DefaultQuery("timezone", "Asia/Jakarta")

	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m,relative_humidity_2m,precipitation,weather_code,wind_speed_10m&timezone=%s",
		lat,
		lng,
		timezone,
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyiapkan request cuaca"})
		return
	}

	resp, err := h.client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Gagal mengambil data cuaca"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Layanan cuaca sedang tidak tersedia"})
		return
	}

	var payload openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Gagal membaca data cuaca"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"source": "open-meteo",
		"location": gin.H{
			"lat":      payload.Latitude,
			"lng":      payload.Longitude,
			"timezone": payload.Timezone,
		},
		"current": gin.H{
			"time":                   payload.Current.Time,
			"temperature_c":          payload.Current.Temperature2m,
			"humidity_percent":       payload.Current.RelativeHumidity2,
			"precipitation_mm":       payload.Current.Precipitation,
			"weather_code":           payload.Current.WeatherCode,
			"wind_speed_kmh":         payload.Current.WindSpeed10m,
			"weather_description_id": weatherCodeDescription(payload.Current.WeatherCode),
		},
	})
}

func (h *WeatherHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/weather/current", h.GetCurrentWeather)
}

func parseCoordOrDefault(raw string, fallback float64) float64 {
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	return v
}

func envOrDefaultFloat(key string, fallback float64) float64 {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	return v
}

func weatherCodeDescription(code int) string {
	switch code {
	case 0:
		return "Cerah"
	case 1, 2, 3:
		return "Berawan"
	case 45, 48:
		return "Berkabut"
	case 51, 53, 55, 56, 57:
		return "Gerimis"
	case 61, 63, 65, 66, 67, 80, 81, 82:
		return "Hujan"
	case 71, 73, 75, 77:
		return "Salju"
	case 95, 96, 99:
		return "Badai Petir"
	default:
		return "Tidak diketahui"
	}
}
