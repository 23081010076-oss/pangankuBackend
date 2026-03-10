package algorithms_test

import (
	"math"
	"testing"

	"github.com/panganku/backend/internal/algorithms"
)

func TestMovingAverage(t *testing.T) {
	prices := []float64{10000, 10200, 10100, 10300, 10500}
	result := algorithms.MovingAverage(prices, 3)
	
	// Index 0 dan 1 harus 0 (data kurang dari window)
	if result[0] != 0 || result[1] != 0 {
		t.Error("Index < window harus 0")
	}
	
	// Test index 2: (10000 + 10200 + 10100) / 3 = 10100
	expected := (10000.0 + 10200.0 + 10100.0) / 3.0
	if math.Abs(result[2]-expected) > 0.01 {
		t.Errorf("SMA[2] = %f, want %f", result[2], expected)
	}
}

func TestPredictNext7Days(t *testing.T) {
	prices := make([]float64, 30)
	for i := range prices {
		prices[i] = 12000.0 + float64(i*50)
	}
	
	result := algorithms.PredictNext7Days(prices)
	
	if len(result) != 7 {
		t.Fatalf("Harus 7 prediksi, dapat %d", len(result))
	}
	
	for i, p := range result {
		if p <= 0 {
			t.Errorf("Prediksi[%d] = %f tidak valid", i, p)
		}
	}
}

func TestDetectAnomalies(t *testing.T) {
	prices := []float64{10000, 10100, 10050, 10200, 10150, 50000, 10100}
	// 50000 di index 5 harus terdeteksi
	
	anomalies := algorithms.DetectAnomalies(prices)
	
	found := false
	for _, idx := range anomalies {
		if idx == 5 {
			found = true
		}
	}
	
	if !found {
		t.Error("Harga 50000 harus terdeteksi sebagai anomali")
	}
}

func TestGetTrend(t *testing.T) {
	// Harga naik terus
	rising := []float64{10000, 10200, 10400, 10600, 10800, 11000, 11200}
	if algorithms.GetTrend(rising) != "NAIK" {
		t.Error("Harus NAIK")
	}
	
	// Harga stabil
	stable := []float64{10000, 10010, 9990, 10005, 9995, 10000, 10002}
	if algorithms.GetTrend(stable) != "STABIL" {
		t.Error("Harus STABIL")
	}
	
	// Harga turun
	falling := []float64{11000, 10800, 10600, 10400, 10200, 10000, 9800}
	if algorithms.GetTrend(falling) != "TURUN" {
		t.Error("Harus TURUN")
	}
}

func TestForecast(t *testing.T) {
	prices := make([]float64, 90)
	for i := range prices {
		prices[i] = 12000 + float64(i*10)
	}
	
	result := algorithms.Forecast(prices)
	
	if len(result.Predictions) != 7 {
		t.Errorf("Predictions harus 7, dapat %d", len(result.Predictions))
	}
	
	if result.Trend == "" {
		t.Error("Trend harus terisi")
	}
	
	if result.Anomalies == nil {
		t.Error("Anomalies harus terisi (bisa kosong)")
	}
}

func BenchmarkForecast(b *testing.B) {
	prices := make([]float64, 90)
	for i := range prices {
		prices[i] = 12000 + float64(i*10)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		algorithms.Forecast(prices)
	}
}
