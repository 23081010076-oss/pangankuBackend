// Penjelasan file:
// Lokasi: internal/handlers/distribusi_handler.go
// Bagian: handler
// File: distribusi_handler
// Fungsi utama: File ini menangani request HTTP, membaca input, dan mengirim response API.
package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/panganku/backend/internal/algorithms"
	"github.com/panganku/backend/internal/middleware"
	"github.com/panganku/backend/internal/models"
	"gorm.io/gorm"
)

// Struct handler ini menyimpan dependency yang dibutuhkan untuk melayani endpoint fitur ini.
type DistribusiHandler struct {
	db *gorm.DB
}

// Constructor ini membuat instance handler baru beserta dependency yang diperlukan.
func NewDistribusiHandler(db *gorm.DB) *DistribusiHandler {
	return &DistribusiHandler{db: db}
}

// Struct request ini merepresentasikan data input yang diharapkan dari body request.
type CreateDistribusiRequest struct {
	DariKecamatanID string    `json:"dari_kecamatan_id" binding:"required,uuid"`
	KeKecamatanID   string    `json:"ke_kecamatan_id" binding:"required,uuid"`
	KomoditasID     string    `json:"komoditas_id" binding:"required,uuid"`
	JumlahKg        float64   `json:"jumlah_kg" binding:"required,gt=0"`
	NamaDriver      string    `json:"nama_driver"`
	NamaKendaraan   string    `json:"nama_kendaraan"`
	JadwalBerangkat time.Time `json:"jadwal_berangkat" binding:"required"`
}

// Struct request ini merepresentasikan data input yang diharapkan dari body request.
type UpdateDistribusiStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=terjadwal dijadwalkan proses selesai batal dibatalkan"`
}

// GetDistribusi - GET /api/v1/distribusi
// Handler ini mengambil data dari backend lalu mengirimkannya sebagai response JSON.
func (h *DistribusiHandler) GetDistribusi(c *gin.Context) {
	status := c.Query("status")
	komoditasID := c.Query("komoditas_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	query := h.db.Model(&models.Distribusi{}).
		Preload("DariKecamatan").
		Preload("KeKecamatan").
		Preload("Komoditas")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if komoditasID != "" {
		query = query.Where("komoditas_id = ?", komoditasID)
	}

	var total int64
	query.Count(&total)

	var list []models.Distribusi
	query.Offset((page - 1) * limit).Limit(limit).Order("created_at desc").Find(&list)

	c.JSON(200, gin.H{
		"data":  list,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetDistribusiByID - GET /api/v1/distribusi/:id
// Handler ini mengambil data dari backend lalu mengirimkannya sebagai response JSON.
func (h *DistribusiHandler) GetDistribusiByID(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	var distribusi models.Distribusi
	if err := h.db.Where("id = ?", id).
		Preload("DariKecamatan").
		Preload("KeKecamatan").
		Preload("Komoditas").
		First(&distribusi).Error; err != nil {
		c.JSON(404, gin.H{"error": "Data distribusi tidak ditemukan"})
		return
	}

	c.JSON(200, gin.H{"data": distribusi})
}

// CreateDistribusi - POST /api/v1/distribusi
// Handler ini menerima input dari request lalu membuat data baru di database.
func (h *DistribusiHandler) CreateDistribusi(c *gin.Context) {
	var req CreateDistribusiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	if req.DariKecamatanID == req.KeKecamatanID {
		c.JSON(400, gin.H{"error": "Kecamatan asal dan tujuan tidak boleh sama"})
		return
	}

	// Validasi kecamatan & komoditas ada
	var dari, ke models.Kecamatan
	var komoditas models.Komoditas
	if h.db.First(&dari, "id = ?", req.DariKecamatanID).Error != nil {
		c.JSON(400, gin.H{"error": "Kecamatan asal tidak ditemukan"})
		return
	}
	if h.db.First(&ke, "id = ?", req.KeKecamatanID).Error != nil {
		c.JSON(400, gin.H{"error": "Kecamatan tujuan tidak ditemukan"})
		return
	}
	if h.db.First(&komoditas, "id = ?", req.KomoditasID).Error != nil {
		c.JSON(400, gin.H{"error": "Komoditas tidak ditemukan"})
		return
	}

	createdBy, _ := uuid.Parse(middleware.GetUserID(c))
	distribusi := models.Distribusi{
		DariKecamatanID: uuid.MustParse(req.DariKecamatanID),
		KeKecamatanID:   uuid.MustParse(req.KeKecamatanID),
		KomoditasID:     uuid.MustParse(req.KomoditasID),
		JumlahKg:        req.JumlahKg,
		Status:          "terjadwal",
		NamaDriver:      req.NamaDriver,
		NamaKendaraan:   req.NamaKendaraan,
		JadwalBerangkat: req.JadwalBerangkat,
		CreatedBy:       createdBy,
	}

	if err := h.db.Create(&distribusi).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal menyimpan data"})
		return
	}

	go func() {
		h.db.Create(&models.AuditLog{
			UserID:    createdBy,
			Action:    "CREATE",
			Resource:  "distribusi",
			IPAddress: c.ClientIP(),
		})
	}()

	c.JSON(201, gin.H{"data": distribusi})
}

// UpdateDistribusiStatus - PUT /api/v1/distribusi/:id/status
// Handler ini menerima perubahan data dari request lalu memperbaruinya di database.
func (h *DistribusiHandler) UpdateDistribusiStatus(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	var req UpdateDistribusiStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Status tidak valid"})
		return
	}

	var distribusi models.Distribusi
	if err := h.db.First(&distribusi, "id = ?", id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Data distribusi tidak ditemukan"})
		return
	}

	h.db.Model(&distribusi).Update("status", req.Status)

	c.JSON(200, gin.H{"data": distribusi, "message": "Status berhasil diperbarui"})
}

// GetRute - GET /api/v1/distribusi/:id/rute
// Handler ini mengambil data dari backend lalu mengirimkannya sebagai response JSON.
func (h *DistribusiHandler) GetRute(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	var distribusi models.Distribusi
	if err := h.db.Where("id = ?", id).
		Preload("DariKecamatan").
		Preload("KeKecamatan").
		First(&distribusi).Error; err != nil {
		c.JSON(404, gin.H{"error": "Data distribusi tidak ditemukan"})
		return
	}

	// Ambil semua kecamatan untuk graph Dijkstra
	var kecamatanList []models.Kecamatan
	h.db.Find(&kecamatanList)

	nodes := make([]algorithms.KecamatanNode, len(kecamatanList))
	for i, k := range kecamatanList {
		nodes[i] = algorithms.KecamatanNode{
			ID:  k.ID.String(),
			Lat: k.Lat,
			Lng: k.Lng,
		}
	}

	rute, jarak := algorithms.Dijkstra(
		nodes,
		distribusi.DariKecamatanID.String(),
		distribusi.KeKecamatanID.String(),
	)

	// Enrich rute dengan nama kecamatan
	kecamatanMap := make(map[string]string)
	for _, k := range kecamatanList {
		kecamatanMap[k.ID.String()] = k.Nama
	}

	type RuteStep struct {
		KecamatanID   string `json:"kecamatan_id"`
		KecamatanNama string `json:"kecamatan_nama"`
	}
	var ruteDetail []RuteStep
	for _, kid := range rute {
		ruteDetail = append(ruteDetail, RuteStep{
			KecamatanID:   kid,
			KecamatanNama: kecamatanMap[kid],
		})
	}

	c.JSON(200, gin.H{
		"rute":     ruteDetail,
		"jarak_km": jarak,
	})
}

// DeleteDistribusi - DELETE /api/v1/distribusi/:id
// Handler ini menghapus data tertentu berdasarkan parameter request.
func (h *DistribusiHandler) DeleteDistribusi(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.db.Delete(&models.Distribusi{}, "id = ?", id).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal menghapus data"})
		return
	}

	c.JSON(200, gin.H{"message": "Data distribusi berhasil dihapus"})
}

// Handler ini menangani proses pendaftaran user baru.
func (h *DistribusiHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/distribusi", h.GetDistribusi)
	r.GET("/distribusi/:id", h.GetDistribusiByID)
	r.GET("/distribusi/:id/rute", h.GetRute)
}
