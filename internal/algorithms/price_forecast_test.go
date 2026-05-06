// Penjelasan file:
// Lokasi: internal/algorithms/price_forecast_test.go
// Bagian: algorithm
// File: price_forecast_test
// Fungsi utama: File ini berisi logika perhitungan atau algoritma pendukung fitur aplikasi.
package algorithms_test

import (
	"math"
	"testing"

	"github.com/panganku/backend/internal/algorithms"
)

func TestMovingAverage(t *testing.T) {
	// Skenario Tabel 5: SMA-7 hari
	// 7 nilai: [12.000, 12.100, 12.050, 12.200, 12.150, 12.300, 12.250]
	prices := []float64{12000, 12100, 12050, 12200, 12150, 12300, 12250}
	result := algorithms.MovingAverage(prices, 7)

	// Harus 12.150 di elemen terakhir
	expected := 12150.0
	if math.Abs(result[len(prices)-1]-expected) > 0.01 {
		t.Errorf("SMA[6] = %f, want %f", result[len(prices)-1], expected)
	}
}

func TestPredictNext7Days(t *testing.T) {
	// Skenario Tabel 5: Smoothing berbobot alpha=0.3
	// Kita mensimulasikan nilai akhir sedemikian hingga H(t) = 12.300 dan mean(7 hari terakhir) = 12.150
	prices := []float64{12000, 12100, 12050, 12200, 12150, 12250, 12300} 
	// mean(7) = (12000+12100+12050+12200+12150+12250+12300)/7 = 85050/7 = 12150
	// H(t) = 12300 (nilai paling terakhir di array sebelum di prediksi)

	result := algorithms.PredictNext7Days(prices)
	
	// Prediksi di hari pertama (t+1) = 0.3 * H(t) + 0.7 * H_7
	// = 0.3 * 12300 + 0.7 * 12150 = 3690 + 8505 = 12195
	// *Catatan: Tabel dokumen mencantumkan 12.255 (kesalahan komutatif 0.7x12.300+0.3x12.150),
	// Algoritma sistem yang valid secara matematis menghasilkan 12195
	expected := 12195.0
	if math.Abs(result[0]-expected) > 0.01 {
		t.Errorf("Prediksi hari pertama = %f, want %f", result[0], expected)
	}
}

func TestDetectAnomalies(t *testing.T) {
	// Skenario Tabel 5: Deteksi anomali (2 sigma)
	// Kita buat set data stabil (variasi sangat kecil), dan injek 1 anomali masif
	prices := []float64{12200, 12210, 12190, 12200, 12200, 13100}

	anomalies := algorithms.DetectAnomalies(prices)

	found := false
	for _, idx := range anomalies {
		if idx == 5 { // Index 5 bernilai 13100
			found = true
		}
	}

	if !found {
		t.Error("Harga 13100 harus terdeteksi sebagai anomali menurut kriteria 2 sigma")
	}
}

func TestGetTrend(t *testing.T) {
	// Skenario Tabel 5: Pola meningkat 3% selama 30 hari
	// TrendPercent > 1.5% -> NAIK
	rising := make([]float64, 30)
	base := 10000.0
	// Buat meningkat total ~30% untuk triggernya, misal setiap hari naik stabil rata-rata > 1.5% harian
	for i := 0; i < 30; i++ {
		base += 200 // +2%: (200/10000 = 2%) per hari konstan
		rising[i] = base
	}

	if algorithms.GetTrend(rising) != "NAIK" {
		t.Error("Harus NAIK (Pola meningkat)")
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

