package handlers

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/panganku/backend/internal/middleware"
	"github.com/panganku/backend/internal/models"
	"github.com/panganku/backend/internal/security"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

type UpdateProfileRequest struct {
	Name        string  `json:"name"`
	Phone       string  `json:"phone"`
	KecamatanID *string `json:"kecamatan_id"`
}

type ChangePasswordRequest struct {
	PasswordLama string `json:"password_lama" binding:"required"`
	PasswordBaru string `json:"password_baru" binding:"required"`
}

// GetProfile - GET /api/v1/users/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error": "User tidak ditemukan"})
			return
		}
		c.JSON(500, gin.H{"error": "Gagal mengambil data user"})
		return
	}

	c.JSON(200, gin.H{
		"id":           user.ID,
		"name":         user.Name,
		"email":        user.Email,
		"phone":        user.Phone,
		"role":         user.Role,
		"kecamatan_id": user.KecamatanID,
		"is_active":    user.IsActive,
		"created_at":   user.CreatedAt,
		"updated_at":   user.UpdatedAt,
	})
}

// UpdateProfile - PUT /api/v1/users/profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Request tidak valid"})
		return
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(404, gin.H{"error": "User tidak ditemukan"})
		return
	}

	updates := map[string]interface{}{}

	if req.Name != "" {
		updates["name"] = security.SanitizeString(req.Name)
	}
	if req.Phone != "" {
		updates["phone"] = security.SanitizeString(req.Phone)
	}
	if req.KecamatanID != nil {
		if *req.KecamatanID == "" {
			updates["kecamatan_id"] = nil
		} else {
			kid, err := uuid.Parse(*req.KecamatanID)
			if err != nil {
				c.JSON(400, gin.H{"error": "Format kecamatan_id tidak valid"})
				return
			}
			// Validate kecamatan exists
			var kec models.Kecamatan
			if err := h.db.First(&kec, "id = ?", kid).Error; err != nil {
				c.JSON(400, gin.H{"error": "Kecamatan tidak ditemukan"})
				return
			}
			updates["kecamatan_id"] = kid
		}
	}

	if len(updates) == 0 {
		c.JSON(400, gin.H{"error": "Tidak ada data yang diupdate"})
		return
	}

	if err := h.db.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal mengupdate profil"})
		return
	}

	c.JSON(200, gin.H{
		"id":           user.ID,
		"name":         user.Name,
		"email":        user.Email,
		"phone":        user.Phone,
		"role":         user.Role,
		"kecamatan_id": user.KecamatanID,
		"updated_at":   user.UpdatedAt,
	})
}

// ChangePassword - PUT /api/v1/users/change-password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Password lama dan baru wajib diisi"})
		return
	}

	if err := security.ValidatePassword(req.PasswordBaru); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(404, gin.H{"error": "User tidak ditemukan"})
		return
	}

	if !security.VerifyPassword(req.PasswordLama, user.Password) {
		c.JSON(401, gin.H{"error": "Password lama tidak sesuai"})
		return
	}

	hashedNew, err := security.HashPassword(req.PasswordBaru)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal memproses password"})
		return
	}

	if err := h.db.Model(&user).Update("password", hashedNew).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal mengupdate password"})
		return
	}

	c.JSON(200, gin.H{"message": "Password berhasil diubah"})
}

// GetAllUsers - GET /api/v1/users (admin only)
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	page := 1
	limit := 20
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	var users []models.User
	var total int64

	query := h.db.Model(&models.User{})
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)
	query.Offset((page - 1) * limit).Limit(limit).Order("created_at desc").Find(&users)

	c.JSON(200, gin.H{
		"data":  users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// UpdateUserRole - PUT /api/v1/users/:id/role (admin only)
func (h *UserHandler) UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	if !security.ValidateUUID(id) {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Role wajib diisi"})
		return
	}

	validRoles := map[string]bool{"admin": true, "petugas": true, "petani": true, "publik": true}
	if !validRoles[req.Role] {
		c.JSON(400, gin.H{"error": "Role tidak valid"})
		return
	}

	result := h.db.Model(&models.User{}).Where("id = ?", id).Update("role", req.Role)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "Gagal mengupdate role"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(404, gin.H{"error": "User tidak ditemukan"})
		return
	}

	c.JSON(200, gin.H{"message": "Role berhasil diupdate"})
}

func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/users/profile", h.GetProfile)
	r.PUT("/users/profile", h.UpdateProfile)
	r.PUT("/users/change-password", h.ChangePassword)
}

func (h *UserHandler) RegisterAdminRoutes(r *gin.RouterGroup) {
	r.GET("/users", h.GetAllUsers)
	r.PUT("/users/:id/role", h.UpdateUserRole)
}
