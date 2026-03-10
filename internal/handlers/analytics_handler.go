package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/panganku/backend/internal/models"
	"gorm.io/gorm"
)

type AnalyticsHandler struct {
	db *gorm.DB
}

func NewAnalyticsHandler(db *gorm.DB) *AnalyticsHandler {
	return &AnalyticsHandler{db: db}
}

// GetDashboard - GET /api/v1/analytics/dashboard
func (h *AnalyticsHandler) GetDashboard(c *gin.Context) {
	today := time.Now().Truncate(24 * time.Hour)

	var totalKomoditas int64
	h.db.Model(&models.Komoditas{}).Count(&totalKomoditas)

	// Alert = laporan darurat aktif (baru/proses)
	var alertCount int64
	h.db.Model(&models.LaporanDarurat{}).
		Where("status IN ?", []string{"baru", "proses"}).
		Count(&alertCount)

	// Update hari ini = harga yang di-input hari ini
	var updateHariIni int64
	h.db.Model(&models.HargaPasar{}).
		Where("tanggal >= ?", today).
		Count(&updateHariIni)

	// Kecamatan aman (stok >= 70% kapasitas)
	var allStok []models.StokPangan
	h.db.Find(&allStok)

	kecamatanStatusMap := make(map[string]string)
	for _, s := range allStok {
		if s.KapasitasKg == 0 {
			continue
		}
		pct := s.StokKg / s.KapasitasKg * 100
		existing := kecamatanStatusMap[s.KecamatanID.String()]
		// Ambil status terburuk per kecamatan
		var newStatus string
		switch {
		case pct >= 70:
			newStatus = "aman"
		case pct >= 30:
			newStatus = "waspada"
		default:
			newStatus = "kritis"
		}
		if existing == "" || (existing == "aman" && newStatus != "aman") || (existing == "waspada" && newStatus == "kritis") {
			kecamatanStatusMap[s.KecamatanID.String()] = newStatus
		}
	}

	kecamatanAman, kecamatanWaspada, kecamatanKritis := 0, 0, 0
	for _, st := range kecamatanStatusMap {
		switch st {
		case "aman":
			kecamatanAman++
		case "waspada":
			kecamatanWaspada++
		case "kritis":
			kecamatanKritis++
		}
	}

	// Rata-rata harga komoditas 7 hari terakhir
	type avgResult struct{ Avg float64 }
	avgQuery := func(namaLike string) float64 {
		var r avgResult
		h.db.Raw(`
			SELECT COALESCE(AVG(h.harga_per_kg), 0) as avg
			FROM harga_pasars h
			JOIN komoditas k ON k.id = h.komoditas_id
			WHERE LOWER(k.nama) LIKE ?
			AND h.tanggal >= ?`,
			"%"+namaLike+"%", time.Now().AddDate(0, 0, -7)).Scan(&r)
		return r.Avg
	}
	avgHargaBeras := avgQuery("beras")
	avgHargaJagung := avgQuery("jagung")
	avgHargaKedelai := avgQuery("kedelai")
	avgHargaCabai := avgQuery("cabai")
	avgHargaGula := avgQuery("gula")
	avgHargaMinyak := avgQuery("minyak")

	// Data harga harian 7 hari terakhir per komoditas
	type dailyResult struct {
		Tanggal string
		Avg     float64
	}
	dailyQuery := func(namaLike string) []float64 {
		var results []dailyResult
		h.db.Raw(`
			SELECT TO_CHAR(DATE(h.tanggal), 'YYYY-MM-DD') as tanggal,
			       COALESCE(AVG(h.harga_per_kg), 0) as avg
			FROM harga_pasars h
			JOIN komoditas k ON k.id = h.komoditas_id
			WHERE LOWER(k.nama) LIKE ?
			AND h.tanggal >= ?
			GROUP BY DATE(h.tanggal)
			ORDER BY tanggal ASC`,
			"%"+namaLike+"%", time.Now().AddDate(0, 0, -6)).Scan(&results)
		out := make([]float64, 7)
		dateMap := make(map[string]float64)
		for _, r := range results {
			dateMap[r.Tanggal] = r.Avg
		}
		for i := 0; i < 7; i++ {
			d := time.Now().AddDate(0, 0, -(6 - i)).Format("2006-01-02")
			out[i] = dateMap[d]
		}
		return out
	}
	harga7HariBeras := dailyQuery("beras")
	harga7HariJagung := dailyQuery("jagung")
	harga7HariKedelai := dailyQuery("kedelai")
	harga7HariCabai := dailyQuery("cabai")
	harga7HariGula := dailyQuery("gula")
	harga7HariMinyak := dailyQuery("minyak")

	// Distribusi aktif
	var distribusiAktif int64
	h.db.Model(&models.Distribusi{}).
		Where("status IN ?", []string{"terjadwal", "proses"}).
		Count(&distribusiAktif)

	// Total laporan bulan ini
	startBulan := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	var laporanBulanIni int64
	h.db.Model(&models.LaporanDarurat{}).
		Where("created_at >= ?", startBulan).
		Count(&laporanBulanIni)

	c.JSON(200, gin.H{
		"total_komoditas":    totalKomoditas,
		"alert_count":        alertCount,
		"update_hari_ini":    updateHariIni,
		"kecamatan_aman":     kecamatanAman,
		"kecamatan_waspada":  kecamatanWaspada,
		"kecamatan_kritis":   kecamatanKritis,
		"avg_harga_beras":    avgHargaBeras,
		"avg_harga_jagung":   avgHargaJagung,
		"avg_harga_kedelai":  avgHargaKedelai,
		"avg_harga_cabai":    avgHargaCabai,
		"avg_harga_gula":     avgHargaGula,
		"avg_harga_minyak":   avgHargaMinyak,
		"harga_7hari_beras":   harga7HariBeras,
		"harga_7hari_jagung":  harga7HariJagung,
		"harga_7hari_kedelai": harga7HariKedelai,
		"harga_7hari_cabai":   harga7HariCabai,
		"harga_7hari_gula":    harga7HariGula,
		"harga_7hari_minyak":  harga7HariMinyak,
		"distribusi_aktif":   distribusiAktif,
		"laporan_bulan_ini":  laporanBulanIni,
	})
}

// GetStatusPangan - GET /api/v1/analytics/status-pangan
func (h *AnalyticsHandler) GetStatusPangan(c *gin.Context) {
	var kecamatanList []models.Kecamatan
	h.db.Find(&kecamatanList)

	type StatusPangan struct {
		KecamatanID   string  `json:"kecamatan_id"`
		KecamatanNama string  `json:"kecamatan_nama"`
		Lat           float64 `json:"lat"`
		Lng           float64 `json:"lng"`
		StatusStok    string  `json:"status_stok"`
		StokPersen    float64 `json:"stok_persen"`
		HargaTrend    string  `json:"harga_trend"`
		JumlahLaporan int64   `json:"jumlah_laporan_aktif"`
	}

	var result []StatusPangan
	for _, kec := range kecamatanList {
		// Hitung status stok rata-rata kecamatan ini
		var stokList []models.StokPangan
		h.db.Where("kecamatan_id = ?", kec.ID).Find(&stokList)

		totalStok, totalKapasitas := 0.0, 0.0
		for _, s := range stokList {
			totalStok += s.StokKg
			totalKapasitas += s.KapasitasKg
		}

		statusStok := "tidak_ada_data"
		stokPersen := 0.0
		if totalKapasitas > 0 {
			stokPersen = totalStok / totalKapasitas * 100
			switch {
			case stokPersen >= 70:
				statusStok = "aman"
			case stokPersen >= 30:
				statusStok = "waspada"
			default:
				statusStok = "kritis"
			}
		}

		// Trend harga terbaru kecamatan ini (rata-rata semua komoditas)
		var trendData []struct {
			Tanggal    time.Time
			HargaPerKg float64
		}
		h.db.Raw(`
			SELECT tanggal, AVG(harga_per_kg) as harga_per_kg
			FROM harga_pasars
			WHERE kecamatan_id = ? AND tanggal >= ?
			GROUP BY tanggal
			ORDER BY tanggal ASC
		`, kec.ID, time.Now().AddDate(0, 0, -30)).Scan(&trendData)

		hargaTrend := "STABIL"
		if len(trendData) >= 2 {
			prices := make([]float64, len(trendData))
			for i, td := range trendData {
				prices[i] = td.HargaPerKg
			}
			// Sederhana: bandingkan rata-rata 7 hari pertama vs 7 hari terakhir
			n := len(prices)
			half := n / 2
			if half > 0 {
				avgFirst, avgLast := 0.0, 0.0
				for i := 0; i < half; i++ {
					avgFirst += prices[i]
				}
				for i := n - half; i < n; i++ {
					avgLast += prices[i]
				}
				avgFirst /= float64(half)
				avgLast /= float64(half)
				changePct := (avgLast - avgFirst) / avgFirst * 100
				switch {
				case changePct > 5:
					hargaTrend = "NAIK"
				case changePct < -5:
					hargaTrend = "TURUN"
				}
			}
		}

		// Laporan darurat aktif di kecamatan
		var jumlahLaporan int64
		h.db.Model(&models.LaporanDarurat{}).
			Where("kecamatan_id = ? AND status IN ?", kec.ID, []string{"baru", "proses"}).
			Count(&jumlahLaporan)

		result = append(result, StatusPangan{
			KecamatanID:   kec.ID.String(),
			KecamatanNama: kec.Nama,
			Lat:           kec.Lat,
			Lng:           kec.Lng,
			StatusStok:    statusStok,
			StokPersen:    stokPersen,
			HargaTrend:    hargaTrend,
			JumlahLaporan: jumlahLaporan,
		})
	}

	c.JSON(200, gin.H{"data": result})
}

func (h *AnalyticsHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/analytics/dashboard", h.GetDashboard)
	r.GET("/analytics/status-pangan", h.GetStatusPangan)
}
