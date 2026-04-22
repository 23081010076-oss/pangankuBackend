// Penjelasan file:
// Lokasi: internal/handlers/komoditas_handler.go
// Bagian: handler
// File: komoditas_handler
// Fungsi utama: File ini menangani request HTTP, membaca input, dan mengirim response API.
package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/panganku/backend/internal/middleware"
	"github.com/panganku/backend/internal/models"
	"gorm.io/gorm"
)

// Struct handler ini menyimpan dependency yang dibutuhkan untuk melayani endpoint fitur ini.
type KomoditasHandler struct {
	db *gorm.DB
}

// Constructor ini membuat instance handler baru beserta dependency yang diperlukan.
func NewKomoditasHandler(db *gorm.DB) *KomoditasHandler {
	return &KomoditasHandler{db: db}
}

// Struct request ini merepresentasikan data input yang diharapkan dari body request.
type KomoditasRequest struct {
	Nama     string `json:"nama" binding:"required"`
	Satuan   string `json:"satuan"`
	Kategori string `json:"kategori"`
}

// GetKomoditas - GET /api/v1/komoditas
// Handler ini mengambil data dari backend lalu mengirimkannya sebagai response JSON.
func (h *KomoditasHandler) GetKomoditas(c *gin.Context) {
	kategori := c.Query("kategori")

	query := h.db.Model(&models.Komoditas{})
	if kategori != "" {
		query = query.Where("kategori = ?", kategori)
	}

	var list []models.Komoditas
	query.Order("nama asc").Find(&list)

	c.JSON(200, gin.H{"data": list})
}

// GetKomoditasByID - GET /api/v1/komoditas/:id
// Handler ini mengambil data dari backend lalu mengirimkannya sebagai response JSON.
func (h *KomoditasHandler) GetKomoditasByID(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	var komoditas models.Komoditas
	if err := h.db.First(&komoditas, "id = ?", id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Komoditas tidak ditemukan"})
		return
	}

	c.JSON(200, gin.H{"data": komoditas})
}

// CreateKomoditas - POST /api/v1/komoditas
// Handler ini menerima input dari request lalu membuat data baru di database.
func (h *KomoditasHandler) CreateKomoditas(c *gin.Context) {
	var req KomoditasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Data tidak valid"})
		return
	}

	satuan := req.Satuan
	if satuan == "" {
		satuan = "kg"
	}

	komoditas := models.Komoditas{
		Nama:     req.Nama,
		Satuan:   satuan,
		Kategori: req.Kategori,
	}

	if err := h.db.Create(&komoditas).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal menyimpan data"})
		return
	}

	go func() {
		userID := middleware.GetUserID(c)
		if uid, err := uuid.Parse(userID); err == nil {
			h.db.Create(&models.AuditLog{
				UserID:    uid,
				Action:    "CREATE",
				Resource:  "komoditas",
				IPAddress: c.ClientIP(),
			})
		}
	}()

	c.JSON(201, gin.H{"data": komoditas})
}

// UpdateKomoditas - PUT /api/v1/komoditas/:id
// Handler ini menerima perubahan data dari request lalu memperbaruinya di database.
func (h *KomoditasHandler) UpdateKomoditas(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	var komoditas models.Komoditas
	if err := h.db.First(&komoditas, "id = ?", id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Komoditas tidak ditemukan"})
		return
	}

	var req KomoditasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Data tidak valid"})
		return
	}

	h.db.Model(&komoditas).Updates(map[string]interface{}{
		"nama":     req.Nama,
		"satuan":   req.Satuan,
		"kategori": req.Kategori,
	})

	c.JSON(200, gin.H{"data": komoditas})
}

// DeleteKomoditas - DELETE /api/v1/komoditas/:id
// Handler ini menghapus data tertentu berdasarkan parameter request.
func (h *KomoditasHandler) DeleteKomoditas(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.db.Delete(&models.Komoditas{}, "id = ?", id).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal menghapus data"})
		return
	}

	c.JSON(200, gin.H{"message": "Komoditas berhasil dihapus"})
}

// Handler ini menangani proses pendaftaran user baru.
func (h *KomoditasHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/komoditas", h.GetKomoditas)
	r.GET("/komoditas/:id", h.GetKomoditasByID)
}
