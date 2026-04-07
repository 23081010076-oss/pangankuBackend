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

type activeAlertItem struct {
	ID            string    `json:"id"`
	JenisMasalah  string    `json:"jenis_masalah"`
	KecamatanNama string    `json:"kecamatan_nama"`
	Status        string    `json:"status"`
	Prioritas     int       `json:"prioritas"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewAnalyticsHandler(db *gorm.DB) *AnalyticsHandler {
	return &AnalyticsHandler{db: db}
}

// GetDashboard - GET /api/v1/analytics/dashboard
func (h *AnalyticsHandler) GetDashboard(c *gin.Context) {
	periode := c.DefaultQuery("periode", "7d")
	days := 7
	switch periode {
	case "30d":
		days = 30
	case "90d":
		days = 90
	default:
		periode = "7d"
	}

	today := time.Now().Truncate(24 * time.Hour)

	var latestHargaDate time.Time
	h.db.Raw("SELECT COALESCE(MAX(tanggal), NOW()) FROM harga_pasars").Scan(&latestHargaDate)
	if latestHargaDate.IsZero() {
		latestHargaDate = today
	}

	rangeStart := latestHargaDate.AddDate(0, 0, -(days - 1))

	var totalKomoditas int64
	h.db.Model(&models.Komoditas{}).Count(&totalKomoditas)

	// Alert = laporan darurat aktif (baru/proses)
	var alertCount int64
	h.db.Model(&models.LaporanDarurat{}).
		Where("status IN ?", []string{"baru", "proses"}).
		Count(&alertCount)

	var activeAlerts []activeAlertItem
	h.db.Table("laporan_darurats l").
		Select("l.id::text as id, l.jenis_masalah, k.nama as kecamatan_nama, l.status, l.prioritas, l.created_at").
		Joins("JOIN kecamatans k ON k.id = l.kecamatan_id").
		Where("l.status IN ?", []string{"baru", "proses"}).
		Order("l.prioritas ASC, l.created_at DESC").
		Limit(5).
		Scan(&activeAlerts)

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
		// Ambil status terburuk per kecamatan menggunakan threshold terpusat
		var newStatus string
		switch {
		case pct >= ThresholdAman:
			newStatus = "aman"
		case pct >= ThresholdWaspada:
			newStatus = "waspada"
		default:
			newStatus = "kritis"
		}
		if existing == "" || (existing == "aman" && newStatus != "aman") || (existing == "waspada" && newStatus == "kritis") {
			kecamatanStatusMap[s.KecamatanID.String()] = newStatus
		}
	}

	kecamatanAman, kecamatanWaspada, kecamatanKritis := 0, 0, 0
	var listKecamatanAman, listKecamatanWaspada, listKecamatanKritis []string

	var allKecamatan []models.Kecamatan
	h.db.Find(&allKecamatan)
	kecNameMap := make(map[string]string)
	for _, k := range allKecamatan {
		kecNameMap[k.ID.String()] = k.Nama
	}

	for idStr, st := range kecamatanStatusMap {
		name := kecNameMap[idStr]
		if name == "" {
			name = "Unknown"
		}
		switch st {
		case "aman":
			kecamatanAman++
			listKecamatanAman = append(listKecamatanAman, name)
		case "waspada":
			kecamatanWaspada++
			listKecamatanWaspada = append(listKecamatanWaspada, name)
		case "kritis":
			kecamatanKritis++
			listKecamatanKritis = append(listKecamatanKritis, name)
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
			AND h.tanggal >= ?
			AND h.tanggal <= ?`,
			"%"+namaLike+"%", rangeStart, latestHargaDate).Scan(&r)
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
			AND h.tanggal <= ?
			GROUP BY DATE(h.tanggal)
			ORDER BY tanggal ASC`,
			"%"+namaLike+"%", rangeStart, latestHargaDate).Scan(&results)
		out := make([]float64, days)
		dateMap := make(map[string]float64)
		for _, r := range results {
			dateMap[r.Tanggal] = r.Avg
		}
		lastValue := 0.0
		for i := 0; i < days; i++ {
			d := rangeStart.AddDate(0, 0, i).Format("2006-01-02")
			if v, ok := dateMap[d]; ok {
				lastValue = v
				out[i] = v
				continue
			}
			out[i] = lastValue
		}
		return out
	}
	harga7HariBeras := dailyQuery("beras")
	harga7HariJagung := dailyQuery("jagung")
	harga7HariKedelai := dailyQuery("kedelai")
	harga7HariCabai := dailyQuery("cabai")
	harga7HariGula := dailyQuery("gula")
	harga7HariMinyak := dailyQuery("minyak")

	tanggalLabels := make([]string, days)
	for i := 0; i < days; i++ {
		tanggalLabels[i] = rangeStart.AddDate(0, 0, i).Format("02/01")
	}

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
		"periode":                periode,
		"tanggal_labels":         tanggalLabels,
		"total_komoditas":        totalKomoditas,
		"alert_count":            alertCount,
		"update_hari_ini":        updateHariIni,
		"kecamatan_aman":         kecamatanAman,
		"kecamatan_waspada":      kecamatanWaspada,
		"kecamatan_kritis":       kecamatanKritis,
		"list_kecamatan_aman":    listKecamatanAman,
		"list_kecamatan_waspada": listKecamatanWaspada,
		"list_kecamatan_kritis":  listKecamatanKritis,
		"avg_harga_beras":        avgHargaBeras,
		"avg_harga_jagung":       avgHargaJagung,
		"avg_harga_kedelai":      avgHargaKedelai,
		"avg_harga_cabai":        avgHargaCabai,
		"avg_harga_gula":         avgHargaGula,
		"avg_harga_minyak":       avgHargaMinyak,
		"harga_7hari_beras":      harga7HariBeras,
		"harga_7hari_jagung":     harga7HariJagung,
		"harga_7hari_kedelai":    harga7HariKedelai,
		"harga_7hari_cabai":      harga7HariCabai,
		"harga_7hari_gula":       harga7HariGula,
		"harga_7hari_minyak":     harga7HariMinyak,
		"distribusi_aktif":       distribusiAktif,
		"laporan_bulan_ini":      laporanBulanIni,
		"active_alerts":          activeAlerts,
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
			case stokPersen >= ThresholdAman:
				statusStok = "aman"
			case stokPersen >= ThresholdWaspada:
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
