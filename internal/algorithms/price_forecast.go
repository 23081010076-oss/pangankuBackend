// Penjelasan file:
// Lokasi: internal/algorithms/price_forecast.go
// Bagian: algorithm
// File: price_forecast
// Fungsi utama: File ini berisi logika perhitungan atau algoritma pendukung fitur aplikasi.
package algorithms

import (
	"math"
)

// MovingAverage menghitung Simple Moving Average dengan memoization
// Kompleksitas: Time O(n), Space O(n) dengan memo vs O(n*window) tanpa memo
func MovingAverage(prices []float64, window int) []float64 {
	n := len(prices)
	result := make([]float64, n)
	memo := make(map[int]float64) // key=index, value=SMA di titik itu

	var sma func(i int) float64
	sma = func(i int) float64 {
		if i < window-1 {
			return 0
		}
		if v, ok := memo[i]; ok {
			return v
		}
		sum := 0.0
		for j := i - window + 1; j <= i; j++ {
			sum += prices[j]
		}
		memo[i] = sum / float64(window)
		return memo[i]
	}

	for i := range prices {
		result[i] = sma(i)
	}
	return result
}

// PredictNext7Days prediksi 7 hari ke depan dengan DP tabulation + EMA
// Kompleksitas: Time O(n), Space O(n)
func PredictNext7Days(prices []float64) []float64 {
	n := len(prices)
	dp := make([]float64, n+7)
	copy(dp, prices) // isi data historis

	alpha := 0.3 // EMA smoothing factor
	for i := n; i < n+7; i++ {
		// EMA: alpha * nilai_sebelumnya + (1-alpha) * rata-rata 7 hari
		sum := 0.0
		for j := i - 7; j < i; j++ {
			sum += dp[j]
		}
		avg7 := sum / 7.0
		dp[i] = alpha*dp[i-1] + (1-alpha)*avg7
	}
	return dp[n:] // return 7 prediksi
}

// DetectAnomalies return index hari yang harganya anomali (> 2 std dev dari mean)
func DetectAnomalies(prices []float64) []int {
	if len(prices) == 0 {
		return []int{}
	}

	n := float64(len(prices))
	
	// hitung mean
	sum := 0.0
	for _, p := range prices {
		sum += p
	}
	mean := sum / n
	
	// hitung standard deviation
	variance := 0.0
	for _, p := range prices {
		variance += (p - mean) * (p - mean)
	}
	stddev := math.Sqrt(variance / n)
	
	// deteksi anomali
	var anomalies []int
	for i, p := range prices {
		if math.Abs(p-mean) > 2*stddev {
			anomalies = append(anomalies, i)
		}
	}
	return anomalies
}

// GetTrend return "NAIK", "TURUN", atau "STABIL" berdasarkan linear regression
func GetTrend(prices []float64) string {
	if len(prices) < 2 {
		return "STABIL"
	}

	n := float64(len(prices))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0
	
	for i, p := range prices {
		x := float64(i)
		sumX += x
		sumY += p
		sumXY += x * p
		sumX2 += x * x
	}
	
	// slope dari least squares
	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return "STABIL"
	}
	
	slope := (n*sumXY - sumX*sumY) / denominator
	meanPrice := sumY / n
	
	if meanPrice == 0 {
		return "STABIL"
	}
	
	slopePct := (slope / meanPrice) * 100 // persentase per hari

	switch {
	case slopePct > 2:
		return "NAIK"
	case slopePct < -2:
		return "TURUN"
	default:
		return "STABIL"
	}
}

type ForecastResult struct {
	Predictions []float64 `json:"predictions"`
	Trend       string    `json:"trend"`
	Anomalies   []int     `json:"anomaly_indexes"`
}

// Forecast melakukan prediksi harga lengkap
func Forecast(prices []float64) ForecastResult {
	return ForecastResult{
		Predictions: PredictNext7Days(prices),
		Trend:       GetTrend(prices),
		Anomalies:   DetectAnomalies(prices),
	}
}

