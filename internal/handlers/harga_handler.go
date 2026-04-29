// Penjelasan file:
// Lokasi: internal/handlers/harga_handler.go
// Bagian: handler
// File: harga_handler
// Fungsi utama: File ini menangani request HTTP, membaca input, dan mengirim response API.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/panganku/backend/internal/algorithms"
	"github.com/panganku/backend/internal/middleware"
	"github.com/panganku/backend/internal/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Struct handler ini menyimpan dependency yang dibutuhkan untuk melayani endpoint fitur ini.
type HargaHandler struct {
	db  *gorm.DB
	rdb *redis.Client
}

// Constructor ini membuat instance handler baru beserta dependency yang diperlukan.
func NewHargaHandler(db *gorm.DB, rdb *redis.Client) *HargaHandler {
	return &HargaHandler{db: db, rdb: rdb}
}

// FlexibleDate menerima format "2006-01-02" maupun RFC3339
type FlexibleDate struct{ time.Time }

func (f *FlexibleDate) UnmarshalJSON(data []byte) error {
	s := string(data)
	if len(s) >= 2 {
		s = s[1 : len(s)-1] // hapus tanda kutip
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			f.Time = t
			return nil
		}
	}
	return errors.New("format tanggal tidak valid, gunakan YYYY-MM-DD atau RFC3339")
}

// Struct request ini merepresentasikan data input yang diharapkan dari body request.
type CreateHargaRequest struct {
	KomoditasID string       `json:"komoditas_id" binding:"required,uuid"`
	KecamatanID string       `json:"kecamatan_id" binding:"required,uuid"`
	HargaPerKg  float64      `json:"harga_per_kg" binding:"required,gt=0"`
	Tanggal     FlexibleDate `json:"tanggal" binding:"required"`
}

// Struct request ini merepresentasikan data input yang diharapkan dari body request.
type UpdateHargaRequest struct {
	HargaPerKg float64      `json:"harga_per_kg" binding:"required,gt=0"`
	Tanggal    FlexibleDate `json:"tanggal" binding:"required"`
}

type HargaLatest struct {
	ID            uuid.UUID `json:"id"`
	KomoditasID   uuid.UUID `json:"komoditas_id"`
	KomoditasNama string    `json:"komoditas_nama"`
	KecamatanID   uuid.UUID `json:"kecamatan_id"`
	KecamatanNama string    `json:"kecamatan_nama"`
	HargaPerKg    float64   `json:"harga_per_kg"`
	Tanggal       time.Time `json:"tanggal"`
	PerubahanPct  float64   `json:"perubahan_persen"`
	Trend         string    `json:"trend"` // NAIK, TURUN, STABIL
	GambarURL     string    `json:"gambar_url,omitempty"`
	Kategori      string    `json:"kategori,omitempty"`
}

type TrendData struct {
	Tanggal string  `json:"tanggal"`
	Avg     float64 `json:"avg"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
}

func (h *HargaHandler) invalidateHargaCache(ctx context.Context, komoditasID string) {
	h.deleteCachePattern(ctx, "harga:latest:*")
	if komoditasID == "" {
		return
	}
	h.deleteCachePattern(ctx, fmt.Sprintf("harga:trend:%s:*", komoditasID))
	h.deleteCachePattern(ctx, fmt.Sprintf("forecast:%s:*", komoditasID))
}

func (h *HargaHandler) deleteCachePattern(ctx context.Context, pattern string) {
	iter := h.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	keys := make([]string, 0, 100)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= 100 {
			h.rdb.Del(ctx, keys...)
			keys = keys[:0]
		}
	}
	if len(keys) > 0 {
		h.rdb.Del(ctx, keys...)
	}
}

// GetHarga - GET /api/v1/harga
// Handler ini mengambil data dari backend lalu mengirimkannya sebagai response JSON.
func (h *HargaHandler) GetHarga(c *gin.Context) {
	komoditasID := c.Query("komoditas_id")
	kecamatanID := c.Query("kecamatan_id")
	tanggalMulai := c.Query("tanggal_mulai")
	tanggalAkhir := c.Query("tanggal_akhir")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	query := h.db.Model(&models.HargaPasar{}).
		Preload("Komoditas").
		Preload("Kecamatan")

	if komoditasID != "" {
		query = query.Where("komoditas_id = ?", komoditasID)
	}
	if kecamatanID != "" {
		query = query.Where("kecamatan_id = ?", kecamatanID)
	}
	if tanggalMulai != "" {
		query = query.Where("tanggal >= ?", tanggalMulai)
	}
	if tanggalAkhir != "" {
		query = query.Where("tanggal <= ?", tanggalAkhir)
	}

	var total int64
	query.Count(&total)

	var hasil []models.HargaPasar
	query.Offset((page - 1) * limit).
		Limit(limit).
		Order("tanggal desc").
		Find(&hasil)

	c.JSON(200, gin.H{
		"data":  hasil,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetLatest - GET /api/v1/harga/latest
// Handler ini mengambil data dari backend lalu mengirimkannya sebagai response JSON.
func (h *HargaHandler) GetLatest(c *gin.Context) {
	ctx := context.Background()
	mode := c.DefaultQuery("mode", "agregat")
	cacheKey := fmt.Sprintf("harga:latest:%s", mode)

	// Cek cache
	cached, err := h.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var hasil []HargaLatest
		if json.Unmarshal([]byte(cached), &hasil) == nil {
			c.JSON(200, gin.H{"data": hasil})
			return
		}
	}

	var hasil []HargaLatest

	if mode == "per_kecamatan" {
		// Perilaku lama: satu data per komoditas dari kecamatan pada timestamp terbaru
		var rows []struct {
			ID            uuid.UUID
			KomoditasID   uuid.UUID
			KomoditasNama string
			KecamatanID   uuid.UUID
			KecamatanNama string
			HargaPerKg    float64
			Tanggal       time.Time
			GambarURL     string
			Kategori      string
		}

		if err := h.db.Raw(`
			SELECT
				h.id, h.komoditas_id, k.nama as komoditas_nama,
				h.kecamatan_id, kc.nama as kecamatan_nama,
				h.harga_per_kg, h.tanggal, k.gambar_url, k.kategori
			FROM (
				SELECT hp.*,
					ROW_NUMBER() OVER (
						PARTITION BY hp.komoditas_id
						ORDER BY hp.tanggal DESC, hp.created_at DESC, hp.id DESC
					) AS rn
				FROM harga_pasars hp
			) h
			JOIN komoditas k ON k.id = h.komoditas_id
			JOIN kecamatans kc ON kc.id = h.kecamatan_id
			WHERE h.rn = 1
			ORDER BY k.nama ASC
		`).Scan(&rows).Error; err != nil {
			c.JSON(500, gin.H{"error": "Gagal mengambil harga terbaru"})
			return
		}

		for _, row := range rows {
			var kemarin models.HargaPasar
			h.db.Where("komoditas_id = ? AND tanggal < ?", row.KomoditasID, row.Tanggal).
				Order("tanggal DESC").
				First(&kemarin)

			perubahan := 0.0
			trend := "STABIL"
			if kemarin.ID != uuid.Nil {
				perubahan = (row.HargaPerKg - kemarin.HargaPerKg) / kemarin.HargaPerKg * 100
				if perubahan > 0 {
					trend = "NAIK"
				} else if perubahan < 0 {
					trend = "TURUN"
				}
			}

			hasil = append(hasil, HargaLatest{
				ID:            row.ID,
				KomoditasID:   row.KomoditasID,
				KomoditasNama: row.KomoditasNama,
				KecamatanID:   row.KecamatanID,
				KecamatanNama: row.KecamatanNama,
				HargaPerKg:    row.HargaPerKg,
				Tanggal:       row.Tanggal,
				PerubahanPct:  perubahan,
				Trend:         trend,
				GambarURL:     row.GambarURL,
				Kategori:      row.Kategori,
			})
		}
	} else {
		// Default: agregat rata-rata semua kecamatan per komoditas
		var rows []struct {
			KomoditasID   uuid.UUID
			KomoditasNama string
			HargaPerKg    float64
			Tanggal       time.Time
			PrevHarga     float64
			GambarURL     string
			Kategori      string
		}

		if err := h.db.Raw(`
			WITH latest_date AS (
				SELECT komoditas_id, MAX(DATE(tanggal)) AS tgl
				FROM harga_pasars
				GROUP BY komoditas_id
			), latest_avg AS (
				SELECT h.komoditas_id, k.nama AS komoditas_nama, k.gambar_url, k.kategori, ld.tgl,
					AVG(h.harga_per_kg) AS avg_harga
				FROM harga_pasars h
				JOIN latest_date ld ON ld.komoditas_id = h.komoditas_id AND DATE(h.tanggal) = ld.tgl
				JOIN komoditas k ON k.id = h.komoditas_id
				GROUP BY h.komoditas_id, k.nama, k.gambar_url, k.kategori, ld.tgl
			), prev_avg AS (
				SELECT ld.komoditas_id, AVG(h.harga_per_kg) AS prev_harga
				FROM latest_date ld
				LEFT JOIN harga_pasars h ON h.komoditas_id = ld.komoditas_id AND DATE(h.tanggal) = DATE_SUB(ld.tgl, INTERVAL 1 DAY)
				GROUP BY ld.komoditas_id
			)
			SELECT la.komoditas_id, la.komoditas_nama, la.gambar_url, la.kategori, la.avg_harga AS harga_per_kg,
				la.tgl AS tanggal, COALESCE(pa.prev_harga, 0) AS prev_harga
			FROM latest_avg la
			LEFT JOIN prev_avg pa ON pa.komoditas_id = la.komoditas_id
			ORDER BY la.komoditas_nama ASC
		`).Scan(&rows).Error; err != nil {
			c.JSON(500, gin.H{"error": "Gagal mengambil harga terbaru"})
			return
		}

		for _, row := range rows {
			perubahan := 0.0
			trend := "STABIL"
			if row.PrevHarga > 0 {
				perubahan = (row.HargaPerKg - row.PrevHarga) / row.PrevHarga * 100
				if perubahan > 0 {
					trend = "NAIK"
				} else if perubahan < 0 {
					trend = "TURUN"
				}
			}

			hasil = append(hasil, HargaLatest{
				ID:            uuid.Nil,
				KomoditasID:   row.KomoditasID,
				KomoditasNama: row.KomoditasNama,
				KecamatanID:   uuid.Nil,
				KecamatanNama: "",
				HargaPerKg:    row.HargaPerKg,
				Tanggal:       row.Tanggal,
				PerubahanPct:  perubahan,
				Trend:         trend,
				GambarURL:     row.GambarURL,
				Kategori:      row.Kategori,
			})
		}
	}

	// Cache selama 15 menit
	if data, err := json.Marshal(hasil); err == nil {
		h.rdb.Set(ctx, cacheKey, data, 15*time.Minute)
	}

	c.JSON(200, gin.H{"data": hasil})
}

// GetTrend - GET /api/v1/harga/trend/:komoditas_id
// Handler ini mengambil data dari backend lalu mengirimkannya sebagai response JSON.
func (h *HargaHandler) GetTrend(c *gin.Context) {
	komoditasID := c.Param("komoditas_id")
	kecamatanID := c.Query("kecamatan_id")
	periode := c.DefaultQuery("periode", "7d")

	// Validasi UUID
	if _, err := uuid.Parse(komoditasID); err != nil {
		c.JSON(400, gin.H{"error": "Format komoditas_id tidak valid"})
		return
	}
	if kecamatanID != "" {
		if _, err := uuid.Parse(kecamatanID); err != nil {
			c.JSON(400, gin.H{"error": "Format kecamatan_id tidak valid"})
			return
		}
	}

	// Hitung tanggal mulai
	days := 7
	switch periode {
	case "30d":
		days = 30
	case "90d":
		days = 90
	}
	tanggalMulai := time.Now().AddDate(0, 0, -days)

	ctx := context.Background()
	cacheKey := fmt.Sprintf("harga:trend:%s:%s:%s", komoditasID, kecamatanID, periode)

	// Cek cache
	cached, err := h.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var hasil []TrendData
		if json.Unmarshal([]byte(cached), &hasil) == nil {
			c.JSON(200, gin.H{"data": hasil})
			return
		}
	}

	// Query trend data
	var rows []TrendData
	query := `
		SELECT 
			DATE(tanggal) as tanggal,
			AVG(harga_per_kg) as avg,
			MIN(harga_per_kg) as min,
			MAX(harga_per_kg) as max
		FROM harga_pasars
		WHERE komoditas_id = ? AND tanggal >= ?
	`
	args := []interface{}{komoditasID, tanggalMulai}

	if kecamatanID != "" {
		query += " AND kecamatan_id = ?"
		args = append(args, kecamatanID)
	}

	query += " GROUP BY DATE(tanggal) ORDER BY tanggal ASC"
	if err := h.db.Raw(query, args...).Scan(&rows).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal mengambil tren harga"})
		return
	}

	// Cache selama 30 menit
	if data, err := json.Marshal(rows); err == nil {
		h.rdb.Set(ctx, cacheKey, data, 30*time.Minute)
	}

	c.JSON(200, gin.H{"data": rows})
}

// CreateHarga - POST /api/v1/harga
// Handler ini menerima input dari request lalu membuat data baru di database.
func (h *HargaHandler) CreateHarga(c *gin.Context) {
	var req CreateHargaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Data tidak valid"})
		return
	}

	// Validasi tanggal tidak di masa depan
	if req.Tanggal.Time.After(time.Now()) {
		c.JSON(400, gin.H{"error": "Tanggal tidak boleh di masa depan"})
		return
	}

	// Validasi komoditas dan kecamatan exist
	var komoditas models.Komoditas
	if err := h.db.First(&komoditas, "id = ?", req.KomoditasID).Error; err != nil {
		c.JSON(400, gin.H{"error": "Komoditas tidak ditemukan"})
		return
	}

	var kecamatan models.Kecamatan
	if err := h.db.First(&kecamatan, "id = ?", req.KecamatanID).Error; err != nil {
		c.JSON(400, gin.H{"error": "Kecamatan tidak ditemukan"})
		return
	}

	// Buat record harga
	createdBy, err := uuid.Parse(middleware.GetUserID(c))
	if err != nil {
		c.JSON(401, gin.H{"error": "User tidak valid"})
		return
	}
	komoditasUUID := uuid.MustParse(req.KomoditasID)
	kecamatanUUID := uuid.MustParse(req.KecamatanID)
	harga := models.HargaPasar{
		KomoditasID: komoditasUUID,
		KecamatanID: kecamatanUUID,
		HargaPerKg:  req.HargaPerKg,
		Tanggal:     req.Tanggal.Time,
		CreatedBy:   createdBy,
	}

	if err := h.db.Create(&harga).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal menyimpan data"})
		return
	}

	// Invalidate cache
	ctx := c.Request.Context()
	h.invalidateHargaCache(ctx, req.KomoditasID)

	// Cek anomali harga (async)
	go h.checkAnomalyAndNotify(req.KomoditasID, req.KecamatanID, req.HargaPerKg, komoditas.Nama, kecamatan.Nama)

	// Audit log
	resource := c.FullPath()
	clientIP := c.ClientIP()
	go func() {
		h.db.Create(&models.AuditLog{
			UserID:    createdBy,
			Action:    "CREATE",
			Resource:  resource,
			IPAddress: clientIP,
		})
	}()

	// Load relations untuk response
	h.db.Preload("Komoditas").Preload("Kecamatan").First(&harga, harga.ID)

	c.JSON(201, gin.H{"data": harga})
}

// UpdateHarga - PUT /api/v1/harga/:id
// Handler ini menerima perubahan data dari request lalu memperbaruinya di database.
func (h *HargaHandler) UpdateHarga(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	var req UpdateHargaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Data tidak valid"})
		return
	}

	if req.Tanggal.Time.After(time.Now()) {
		c.JSON(400, gin.H{"error": "Tanggal tidak boleh di masa depan"})
		return
	}

	var harga models.HargaPasar
	if err := h.db.First(&harga, "id = ?", id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Data harga tidak ditemukan"})
		return
	}

	// Petani hanya boleh mengubah data harga miliknya sendiri.
	role := middleware.GetRole(c)
	if role == "petani" {
		uid, err := uuid.Parse(middleware.GetUserID(c))
		if err != nil {
			c.JSON(401, gin.H{"error": "User tidak valid"})
			return
		}
		if harga.CreatedBy != uid {
			c.JSON(403, gin.H{"error": "Akses ditolak"})
			return
		}
	}

	if err := h.db.Model(&harga).Updates(map[string]interface{}{
		"harga_per_kg": req.HargaPerKg,
		"tanggal":      req.Tanggal.Time,
	}).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal memperbarui data harga"})
		return
	}

	ctx := c.Request.Context()
	h.invalidateHargaCache(ctx, harga.KomoditasID.String())

	resource := c.FullPath()
	clientIP := c.ClientIP()
	userID := middleware.GetUserID(c)
	go func() {
		if uid, err := uuid.Parse(userID); err == nil {
			h.db.Create(&models.AuditLog{
				UserID:    uid,
				Action:    "UPDATE",
				Resource:  resource,
				IPAddress: clientIP,
			})
		}
	}()

	h.db.Preload("Komoditas").Preload("Kecamatan").First(&harga, "id = ?", id)
	c.JSON(200, gin.H{"data": harga})
}

// DeleteHarga - DELETE /api/v1/harga/:id
// Handler ini menghapus data tertentu berdasarkan parameter request.
func (h *HargaHandler) DeleteHarga(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	var harga models.HargaPasar
	if err := h.db.First(&harga, "id = ?", id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Data harga tidak ditemukan"})
		return
	}

	if err := h.db.Delete(&models.HargaPasar{}, "id = ?", id).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal menghapus data harga"})
		return
	}

	ctx := c.Request.Context()
	h.invalidateHargaCache(ctx, harga.KomoditasID.String())

	resource := c.FullPath()
	clientIP := c.ClientIP()
	userID := middleware.GetUserID(c)
	go func() {
		if uid, err := uuid.Parse(userID); err == nil {
			h.db.Create(&models.AuditLog{
				UserID:    uid,
				Action:    "DELETE",
				Resource:  resource,
				IPAddress: clientIP,
			})
		}
	}()

	c.JSON(200, gin.H{"message": "Data harga berhasil dihapus"})
}

// GetForecast - GET /api/v1/harga/forecast
// Handler ini mengambil data dari backend lalu mengirimkannya sebagai response JSON.
func (h *HargaHandler) GetForecast(c *gin.Context) {
	komoditasID := c.Query("komoditas_id")
	kecamatanID := c.Query("kecamatan_id")

	if komoditasID == "" {
		c.JSON(400, gin.H{"error": "komoditas_id wajib diisi"})
		return
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("forecast:%s:%s", komoditasID, kecamatanID)

	// Cek cache
	cached, err := h.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var result algorithms.ForecastResult
		if json.Unmarshal([]byte(cached), &result) == nil {
			c.JSON(200, result)
			return
		}
	}

	// Ambil data 90 hari terakhir
	query := h.db.Model(&models.HargaPasar{}).
		Where("komoditas_id = ? AND tanggal >= ?", komoditasID, time.Now().AddDate(0, 0, -90))

	if kecamatanID != "" {
		query = query.Where("kecamatan_id = ?", kecamatanID)
	}

	var hargaList []models.HargaPasar
	query.Order("tanggal asc").Find(&hargaList)

	if len(hargaList) < 7 {
		c.JSON(400, gin.H{"error": "Data historis tidak cukup (minimal 7 hari)"})
		return
	}

	// Ekstrak harga ke slice
	prices := make([]float64, len(hargaList))
	for i, h := range hargaList {
		prices[i] = h.HargaPerKg
	}

	// Jalankan forecast
	result := algorithms.Forecast(prices)

	// Cache selama 6 jam
	if data, err := json.Marshal(result); err == nil {
		h.rdb.Set(ctx, cacheKey, data, 6*time.Hour)
	}

	c.JSON(200, result)
}

// checkAnomalyAndNotify cek apakah harga anomali
// Handler ini menjalankan logika endpoint sesuai kebutuhan fitur pada request yang masuk.
func (h *HargaHandler) checkAnomalyAndNotify(komoditasID, kecamatanID string, hargaBaru float64, namaKomoditas, namaKecamatan string) {
	// Hitung rata-rata 7 hari terakhir
	var avg float64
	h.db.Model(&models.HargaPasar{}).
		Where("komoditas_id = ? AND kecamatan_id = ? AND tanggal >= ?",
			komoditasID, kecamatanID, time.Now().AddDate(0, 0, -7)).
		Select("AVG(harga_per_kg)").
		Scan(&avg)

	if avg == 0 {
		return // Tidak ada data historis
	}

	// Cek jika harga > 120% atau < 80% dari rata-rata
	if hargaBaru > avg*1.2 || hargaBaru < avg*0.8 {
		message := fmt.Sprintf("Harga %s di %s anomali: Rp%.0f (rata-rata 7 hari: Rp%.0f)",
			namaKomoditas, namaKecamatan, hargaBaru, avg)

		var recipients []models.User
		if err := h.db.Where("role IN ? AND is_active = true", []string{"admin", "petugas"}).Find(&recipients).Error; err != nil {
			fmt.Println("ALERT:", message)
			return
		}
		for _, user := range recipients {
			SendNotifikasi(h.db, user.ID, "Anomali harga terdeteksi", message, "warning", "/harga")
		}
	}
}

// RegisterRoutes mendaftarkan semua route harga
// Handler ini menangani proses pendaftaran user baru.
func (h *HargaHandler) RegisterRoutes(r *gin.RouterGroup) {
	harga := r.Group("/harga")
	{
		harga.GET("", h.GetHarga)
		harga.GET("/latest", h.GetLatest)
		harga.GET("/trend/:komoditas_id", h.GetTrend)
		harga.GET("/forecast", h.GetForecast)
		harga.POST("", middleware.JWTAuth(h.rdb), middleware.RequireRole("admin", "petugas", "petani"), h.CreateHarga)
		harga.PUT("/:id", middleware.JWTAuth(h.rdb), middleware.RequireRole("admin", "petugas", "petani"), h.UpdateHarga)
		harga.DELETE("/:id", middleware.JWTAuth(h.rdb), middleware.RequireRole("admin", "petugas"), h.DeleteHarga)
	}
}
