// Penjelasan file:
// Lokasi: internal/handlers/luas_lahan_handler.go
// Bagian: handler
// File: luas_lahan_handler
// Fungsi utama: File ini menangani request HTTP, membaca input, dan mengirim response API.
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

// Struct handler ini menyimpan dependency yang dibutuhkan untuk melayani endpoint fitur ini.
type LuasLahanHandler struct {
	db *gorm.DB
}

// Struct request ini merepresentasikan data input yang diharapkan dari body request.
type LuasLahanRequest struct {
	KomoditasID string  `json:"komoditas_id" binding:"required,uuid"`
	KecamatanID string  `json:"kecamatan_id" binding:"required,uuid"`
	LuasHa      float64 `json:"luas_ha" binding:"required,gt=0"`
	Tahun       int     `json:"tahun" binding:"omitempty,min=2000,max=2100"`
}

// Constructor ini membuat instance handler baru beserta dependency yang diperlukan.
func NewLuasLahanHandler(db *gorm.DB) *LuasLahanHandler {
	return &LuasLahanHandler{db: db}
}

// Handler ini mengambil data dari backend lalu mengirimkannya sebagai response JSON.
func (h *LuasLahanHandler) GetLuasLahan(c *gin.Context) {
	komoditasID := c.Query("komoditas_id")
	kecamatanID := c.Query("kecamatan_id")
	tahun, _ := strconv.Atoi(c.Query("tahun"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 100
	}

	query := h.db.Model(&models.LuasLahan{}).
		Preload("Komoditas").
		Preload("Kecamatan")

	if komoditasID != "" {
		query = query.Where("komoditas_id = ?", komoditasID)
	}
	if kecamatanID != "" {
		query = query.Where("kecamatan_id = ?", kecamatanID)
	}
	if tahun > 0 {
		query = query.Where("tahun = ?", tahun)
	}

	var total int64
	query.Count(&total)

	var list []models.LuasLahan
	query.
		Order("tahun desc, updated_at desc").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&list)

	c.JSON(200, gin.H{
		"data":  list,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// Handler ini menerima input dari request lalu membuat data baru di database.
func (h *LuasLahanHandler) CreateOrUpdateLuasLahan(c *gin.Context) {
	var req LuasLahanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	if req.Tahun == 0 {
		req.Tahun = time.Now().Year()
	}

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

	var luasLahan models.LuasLahan
	result := h.db.Where(
		"komoditas_id = ? AND kecamatan_id = ? AND tahun = ?",
		req.KomoditasID,
		req.KecamatanID,
		req.Tahun,
	).First(&luasLahan)

	if result.Error != nil {
		luasLahan = models.LuasLahan{
			KomoditasID: uuid.MustParse(req.KomoditasID),
			KecamatanID: uuid.MustParse(req.KecamatanID),
			LuasHa:      req.LuasHa,
			Tahun:       req.Tahun,
			PetugasID:   ptID,
			UpdatedAt:   time.Now(),
		}
		if err := h.db.Create(&luasLahan).Error; err != nil {
			c.JSON(500, gin.H{"error": "Gagal menyimpan data luas lahan"})
			return
		}
	} else {
		if err := h.db.Model(&luasLahan).Updates(map[string]interface{}{
			"luas_ha":    req.LuasHa,
			"petugas_id": ptID,
			"updated_at": time.Now(),
		}).Error; err != nil {
			c.JSON(500, gin.H{"error": "Gagal memperbarui data luas lahan"})
			return
		}
	}

	go func() {
		if uid, err := uuid.Parse(petugasID); err == nil {
			h.db.Create(&models.AuditLog{
				UserID:    uid,
				Action:    "UPDATE",
				Resource:  "luas_lahan",
				IPAddress: c.ClientIP(),
			})
		}
	}()

	h.db.Where("id = ?", luasLahan.ID).
		Preload("Komoditas").
		Preload("Kecamatan").
		First(&luasLahan)

	c.JSON(200, gin.H{"data": luasLahan})
}

// Handler ini menghapus data tertentu berdasarkan parameter request.
func (h *LuasLahanHandler) DeleteLuasLahan(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	var luasLahan models.LuasLahan
	if err := h.db.First(&luasLahan, "id = ?", id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Data luas lahan tidak ditemukan"})
		return
	}

	if err := h.db.Delete(&models.LuasLahan{}, "id = ?", id).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal menghapus data luas lahan"})
		return
	}

	go func() {
		if uid, err := uuid.Parse(middleware.GetUserID(c)); err == nil {
			h.db.Create(&models.AuditLog{
				UserID:    uid,
				Action:    "DELETE",
				Resource:  "luas_lahan",
				IPAddress: c.ClientIP(),
			})
		}
	}()

	c.JSON(200, gin.H{"message": "Data luas lahan berhasil dihapus"})
}

// Handler ini menangani proses pendaftaran user baru.
func (h *LuasLahanHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/luas-lahan", h.GetLuasLahan)
}
