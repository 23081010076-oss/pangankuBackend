package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/panganku/backend/internal/middleware"
	"github.com/panganku/backend/internal/models"
	"gorm.io/gorm"
)

type KecamatanHandler struct {
	db *gorm.DB
}

func NewKecamatanHandler(db *gorm.DB) *KecamatanHandler {
	return &KecamatanHandler{db: db}
}

type KecamatanRequest struct {
	Nama   string  `json:"nama" binding:"required"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	LuasHa float64 `json:"luas_ha"`
}

// GetKecamatan - GET /api/v1/kecamatan
func (h *KecamatanHandler) GetKecamatan(c *gin.Context) {
	var list []models.Kecamatan
	h.db.Order("nama asc").Find(&list)
	c.JSON(200, gin.H{"data": list})
}

// GetKecamatanByID - GET /api/v1/kecamatan/:id
func (h *KecamatanHandler) GetKecamatanByID(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	var kecamatan models.Kecamatan
	if err := h.db.First(&kecamatan, "id = ?", id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Kecamatan tidak ditemukan"})
		return
	}

	c.JSON(200, gin.H{"data": kecamatan})
}

// CreateKecamatan - POST /api/v1/kecamatan
func (h *KecamatanHandler) CreateKecamatan(c *gin.Context) {
	var req KecamatanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Data tidak valid"})
		return
	}

	kecamatan := models.Kecamatan{
		Nama:   req.Nama,
		Lat:    req.Lat,
		Lng:    req.Lng,
		LuasHa: req.LuasHa,
	}

	if err := h.db.Create(&kecamatan).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal menyimpan data"})
		return
	}

	go func() {
		userID := middleware.GetUserID(c)
		if uid, err := uuid.Parse(userID); err == nil {
			h.db.Create(&models.AuditLog{
				UserID:    uid,
				Action:    "CREATE",
				Resource:  "kecamatan",
				IPAddress: c.ClientIP(),
			})
		}
	}()

	c.JSON(201, gin.H{"data": kecamatan})
}

// UpdateKecamatan - PUT /api/v1/kecamatan/:id
func (h *KecamatanHandler) UpdateKecamatan(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	var kecamatan models.Kecamatan
	if err := h.db.First(&kecamatan, "id = ?", id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Kecamatan tidak ditemukan"})
		return
	}

	var req KecamatanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Data tidak valid"})
		return
	}

	h.db.Model(&kecamatan).Updates(map[string]interface{}{
		"nama":    req.Nama,
		"lat":     req.Lat,
		"lng":     req.Lng,
		"luas_ha": req.LuasHa,
	})

	c.JSON(200, gin.H{"data": kecamatan})
}

// DeleteKecamatan - DELETE /api/v1/kecamatan/:id
func (h *KecamatanHandler) DeleteKecamatan(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.db.Delete(&models.Kecamatan{}, "id = ?", id).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal menghapus data"})
		return
	}

	c.JSON(200, gin.H{"message": "Kecamatan berhasil dihapus"})
}

func (h *KecamatanHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/kecamatan", h.GetKecamatan)
	r.GET("/kecamatan/:id", h.GetKecamatanByID)
}
