package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/panganku/backend/internal/middleware"
	"github.com/panganku/backend/internal/models"
	"gorm.io/gorm"
)

type StokHandler struct {
	db *gorm.DB
}

func NewStokHandler(db *gorm.DB) *StokHandler {
	return &StokHandler{db: db}
}

type StokRequest struct {
	KomoditasID string  `json:"komoditas_id" binding:"required,uuid"`
	KecamatanID string  `json:"kecamatan_id" binding:"required,uuid"`
	StokKg      float64 `json:"stok_kg" binding:"required,gte=0"`
	KapasitasKg float64 `json:"kapasitas_kg" binding:"required,gt=0"`
}

func hitungStatusStok(stokKg, kapasitasKg float64) string {
	if kapasitasKg == 0 {
		return "kritis"
	}
	pct := stokKg / kapasitasKg * 100
	switch {
	case pct >= 70:
		return "aman"
	case pct >= 30:
		return "waspada"
	default:
		return "kritis"
	}
}

// GetStok - GET /api/v1/stok
func (h *StokHandler) GetStok(c *gin.Context) {
	komoditasID := c.Query("komoditas_id")
	kecamatanID := c.Query("kecamatan_id")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	query := h.db.Model(&models.StokPangan{}).
		Preload("Komoditas").
		Preload("Kecamatan")

	if komoditasID != "" {
		query = query.Where("komoditas_id = ?", komoditasID)
	}
	if kecamatanID != "" {
		query = query.Where("kecamatan_id = ?", kecamatanID)
	}

	var allStok []models.StokPangan
	query.Order("updated_at desc").Find(&allStok)

	// Filter by status (aman/waspada/kritis) setelah query karena status dihitung
	type StokResponse struct {
		models.StokPangan
		StatusStok string  `json:"status_stok"`
		StokPersen float64 `json:"stok_persen"`
	}

	var hasil []StokResponse
	for _, s := range allStok {
		stokStatus := hitungStatusStok(s.StokKg, s.KapasitasKg)
		if status != "" && stokStatus != status {
			continue
		}
		persen := 0.0
		if s.KapasitasKg > 0 {
			persen = s.StokKg / s.KapasitasKg * 100
		}
		hasil = append(hasil, StokResponse{
			StokPangan: s,
			StatusStok: stokStatus,
			StokPersen: persen,
		})
	}

	// Pagination manual
	total := len(hasil)
	start := (page - 1) * limit
	end := start + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	paginated := hasil[start:end]

	c.JSON(200, gin.H{
		"data":  paginated,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetStokByKecamatan - GET /api/v1/stok/kecamatan/:id
func (h *StokHandler) GetStokByKecamatan(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	var stokList []models.StokPangan
	h.db.Where("kecamatan_id = ?", id).
		Preload("Komoditas").
		Preload("Kecamatan").
		Find(&stokList)

	type StokResp struct {
		models.StokPangan
		StatusStok string  `json:"status_stok"`
		StokPersen float64 `json:"stok_persen"`
	}

	var hasil []StokResp
	for _, s := range stokList {
		persen := 0.0
		if s.KapasitasKg > 0 {
			persen = s.StokKg / s.KapasitasKg * 100
		}
		hasil = append(hasil, StokResp{
			StokPangan: s,
			StatusStok: hitungStatusStok(s.StokKg, s.KapasitasKg),
			StokPersen: persen,
		})
	}

	c.JSON(200, gin.H{"data": hasil})
}

// CreateOrUpdateStok - POST /api/v1/stok
func (h *StokHandler) CreateOrUpdateStok(c *gin.Context) {
	var req StokRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	// Cek komoditas & kecamatan ada
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

	petugasID := middleware.GetUserID(c)
	ptID, _ := uuid.Parse(petugasID)

	// Upsert: cek apakah sudah ada kombinasi komoditas+kecamatan
	var stok models.StokPangan
	result := h.db.Where("komoditas_id = ? AND kecamatan_id = ?", req.KomoditasID, req.KecamatanID).First(&stok)

	if result.Error != nil {
		// Create baru
		stok = models.StokPangan{
			KomoditasID: uuid.MustParse(req.KomoditasID),
			KecamatanID: uuid.MustParse(req.KecamatanID),
			StokKg:      req.StokKg,
			KapasitasKg: req.KapasitasKg,
			PetugasID:   ptID,
			UpdatedAt:   time.Now(),
		}
		h.db.Create(&stok)
	} else {
		// Update existing
		h.db.Model(&stok).Updates(map[string]interface{}{
			"stok_kg":      req.StokKg,
			"kapasitas_kg": req.KapasitasKg,
			"petugas_id":   ptID,
			"updated_at":   time.Now(),
		})
	}

	// Audit log
	go func() {
		if uid, err := uuid.Parse(petugasID); err == nil {
			h.db.Create(&models.AuditLog{
				UserID:    uid,
				Action:    "UPDATE",
				Resource:  "stok_pangan",
				IPAddress: c.ClientIP(),
			})
		}
	}()

	// Reload dengan relasi
	h.db.Where("id = ?", stok.ID).Preload("Komoditas").Preload("Kecamatan").First(&stok)

	persen := 0.0
	if stok.KapasitasKg > 0 {
		persen = stok.StokKg / stok.KapasitasKg * 100
	}

	c.JSON(200, gin.H{
		"data":        stok,
		"status_stok": hitungStatusStok(stok.StokKg, stok.KapasitasKg),
		"stok_persen": persen,
	})
}

func (h *StokHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/stok", h.GetStok)
	r.GET("/stok/kecamatan/:id", h.GetStokByKecamatan)
}

// DeleteStok - DELETE /api/v1/stok/:id
func (h *StokHandler) DeleteStok(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	var stok models.StokPangan
	if err := h.db.First(&stok, "id = ?", id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Data stok tidak ditemukan"})
		return
	}

	if err := h.db.Delete(&models.StokPangan{}, "id = ?", id).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal menghapus data stok"})
		return
	}

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

	c.JSON(200, gin.H{"message": "Data stok berhasil dihapus"})
}
