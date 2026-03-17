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

type HargaHandler struct {
	db  *gorm.DB
	rdb *redis.Client
}

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

type CreateHargaRequest struct {
	KomoditasID string       `json:"komoditas_id" binding:"required,uuid"`
	KecamatanID string       `json:"kecamatan_id" binding:"required,uuid"`
	HargaPerKg  float64      `json:"harga_per_kg" binding:"required,gt=0"`
	Tanggal     FlexibleDate `json:"tanggal" binding:"required"`
}

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
}

type TrendData struct {
	Tanggal string  `json:"tanggal"`
	Avg     float64 `json:"avg"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
}

// GetHarga - GET /api/v1/harga
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
func (h *HargaHandler) GetLatest(c *gin.Context) {
	ctx := context.Background()
	cacheKey := "harga:latest"

	// Cek cache
	cached, err := h.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var hasil []HargaLatest
		if json.Unmarshal([]byte(cached), &hasil) == nil {
			c.JSON(200, gin.H{"data": hasil})
			return
		}
	}

	// Query latest harga per komoditas
	var rows []struct {
		ID            uuid.UUID
		KomoditasID   uuid.UUID
		KomoditasNama string
		KecamatanID   uuid.UUID
		KecamatanNama string
		HargaPerKg    float64
		Tanggal       time.Time
	}

	h.db.Raw(`
		SELECT DISTINCT ON (h.komoditas_id) 
			h.id, h.komoditas_id, k.nama as komoditas_nama, 
			h.kecamatan_id, kc.nama as kecamatan_nama,
			h.harga_per_kg, h.tanggal
		FROM harga_pasars h
		JOIN komoditas k ON k.id = h.komoditas_id
		JOIN kecamatans kc ON kc.id = h.kecamatan_id
		ORDER BY h.komoditas_id, h.tanggal DESC
	`).Scan(&rows)

	var hasil []HargaLatest
	for _, row := range rows {
		// Hitung perubahan vs hari sebelumnya
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
		})
	}

	// Cache selama 15 menit
	if data, err := json.Marshal(hasil); err == nil {
		h.rdb.Set(ctx, cacheKey, data, 15*time.Minute)
	}

	c.JSON(200, gin.H{"data": hasil})
}

// GetTrend - GET /api/v1/harga/trend/:komoditas_id
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
	cacheKey := fmt.Sprintf("harga:trend:%s:%s", komoditasID, periode)

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
	h.db.Raw(query, args...).Scan(&rows)

	// Cache selama 30 menit
	if data, err := json.Marshal(rows); err == nil {
		h.rdb.Set(ctx, cacheKey, data, 30*time.Minute)
	}

	c.JSON(200, gin.H{"data": rows})
}

// CreateHarga - POST /api/v1/harga
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
	createdBy, _ := uuid.Parse(middleware.GetUserID(c))
	harga := models.HargaPasar{
		KomoditasID: uuid.MustParse(req.KomoditasID),
		KecamatanID: uuid.MustParse(req.KecamatanID),
		HargaPerKg:  req.HargaPerKg,
		Tanggal:     req.Tanggal.Time,
		CreatedBy:   createdBy,
	}

	if err := h.db.Create(&harga).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal menyimpan data"})
		return
	}

	// Invalidate cache
	ctx := context.Background()
	h.rdb.Del(ctx, "harga:latest")
	h.rdb.Del(ctx, fmt.Sprintf("harga:trend:%s:*", req.KomoditasID))

	// Cek anomali harga (async)
	go h.checkAnomalyAndNotify(req.KomoditasID, req.KecamatanID, req.HargaPerKg, komoditas.Nama, kecamatan.Nama)

	// Audit log
	go func() {
		h.db.Create(&models.AuditLog{
			UserID:    createdBy,
			Action:    "CREATE",
			Resource:  c.FullPath(),
			IPAddress: c.ClientIP(),
		})
	}()

	// Load relations untuk response
	h.db.Preload("Komoditas").Preload("Kecamatan").First(&harga, harga.ID)

	c.JSON(201, gin.H{"data": harga})
}

// UpdateHarga - PUT /api/v1/harga/:id
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

	// Pedagang/petani hanya boleh mengubah data harga miliknya sendiri.
	role := middleware.GetRole(c)
	if role == "pedagang" || role == "petani" {
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

	h.db.Model(&harga).Updates(map[string]interface{}{
		"harga_per_kg": req.HargaPerKg,
		"tanggal":      req.Tanggal.Time,
	})

	ctx := context.Background()
	h.rdb.Del(ctx, "harga:latest")
	h.rdb.Del(ctx, fmt.Sprintf("harga:trend:%s:*", harga.KomoditasID.String()))

	go func() {
		if uid, err := uuid.Parse(middleware.GetUserID(c)); err == nil {
			h.db.Create(&models.AuditLog{
				UserID:    uid,
				Action:    "UPDATE",
				Resource:  c.FullPath(),
				IPAddress: c.ClientIP(),
			})
		}
	}()

	h.db.Preload("Komoditas").Preload("Kecamatan").First(&harga, "id = ?", id)
	c.JSON(200, gin.H{"data": harga})
}

// DeleteHarga - DELETE /api/v1/harga/:id
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

	ctx := context.Background()
	h.rdb.Del(ctx, "harga:latest")
	h.rdb.Del(ctx, fmt.Sprintf("harga:trend:%s:*", harga.KomoditasID.String()))

	go func() {
		if uid, err := uuid.Parse(middleware.GetUserID(c)); err == nil {
			h.db.Create(&models.AuditLog{
				UserID:    uid,
				Action:    "DELETE",
				Resource:  c.FullPath(),
				IPAddress: c.ClientIP(),
			})
		}
	}()

	c.JSON(200, gin.H{"message": "Data harga berhasil dihapus"})
}

// GetForecast - GET /api/v1/harga/forecast
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

		// TODO: Simpan ke tabel notifikasi
		// Sementara hanya log
		fmt.Println("ALERT:", message)
	}
}

// RegisterRoutes mendaftarkan semua route harga
func (h *HargaHandler) RegisterRoutes(r *gin.RouterGroup) {
	harga := r.Group("/harga")
	{
		harga.GET("", h.GetHarga)
		harga.GET("/latest", h.GetLatest)
		harga.GET("/trend/:komoditas_id", h.GetTrend)
		harga.GET("/forecast", h.GetForecast)
		harga.POST("", middleware.JWTAuth(h.rdb), middleware.RequireRole("admin", "petugas", "petani", "pedagang"), h.CreateHarga)
		harga.PUT("/:id", middleware.JWTAuth(h.rdb), middleware.RequireRole("admin", "petugas", "petani", "pedagang"), h.UpdateHarga)
		harga.DELETE("/:id", middleware.JWTAuth(h.rdb), middleware.RequireRole("admin", "petugas"), h.DeleteHarga)
	}
}
