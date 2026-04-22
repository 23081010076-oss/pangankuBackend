// Penjelasan file:
// Lokasi: internal/handlers/telegram_handler.go
// Bagian: handler
// File: telegram_handler
// Fungsi utama: File ini menangani request HTTP, membaca input, dan mengirim response API.
package handlers

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Struct handler ini menyimpan dependency yang dibutuhkan untuk melayani endpoint fitur ini.
type TelegramHandler struct {
	db  *gorm.DB
	rdb *redis.Client
	bot *tgbotapi.BotAPI
}

// Constructor ini membuat instance handler baru beserta dependency yang diperlukan.
func NewTelegramHandler(db *gorm.DB, rdb *redis.Client) *TelegramHandler {
	// Initialize bot using token from environment variables
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	var bot *tgbotapi.BotAPI
	var err error

	if token != "" {
		bot, err = tgbotapi.NewBotAPI(token)
		if err != nil {
			log.Printf("Failed to initialize Telegram Bot: %v\n", err)
		} else {
			log.Printf("Telegram Bot initialized: %s\n", bot.Self.UserName)
		}
	} else {
		log.Println("TELEGRAM_BOT_TOKEN not set, Telegram features disabled")
	}

	return &TelegramHandler{
		db:  db,
		rdb: rdb,
		bot: bot,
	}
}

// RegisterRoutes registers endpoints for custom Telegram-related actions
// Handler ini menangani proses pendaftaran user baru.
func (h *TelegramHandler) RegisterRoutes(r *gin.RouterGroup) {
	tgGroup := r.Group("/telegram")
	{
		tgGroup.POST("/send", h.SendMessage)
	}
}

// SendMessage API endpoint to trigger sending telegram from front-end/test via API
// Handler ini mengirim pesan atau notifikasi ke layanan eksternal.
func (h *TelegramHandler) SendMessage(c *gin.Context) {
	if h.bot == nil {
		c.JSON(503, gin.H{"error": "Telegram Bot is not configured. Missing TELEGRAM_BOT_TOKEN"})
		return
	}

	var req struct {
		ChatID int64  `json:"chat_id" binding:"required"`
		Text   string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	msg := tgbotapi.NewMessage(req.ChatID, req.Text)

	// Set HTML parsing so we can use bold/italic
	msg.ParseMode = "HTML"

	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending telegram: %v", err)
		c.JSON(500, gin.H{"error": "Gagal mengirim pesan Telegram"})
		return
	}

	c.JSON(200, gin.H{"message": "Pesan Telegram berhasil dikirim"})
}

// SendBroadcastMessage is a helper function to send message programmatically
// Handler ini mengirim pesan atau notifikasi ke layanan eksternal.
func (h *TelegramHandler) SendBroadcastMessage(ctx context.Context, chatID int64, textMsg string) error {
	if h.bot == nil {
		return fmt.Errorf("bot not initialized")
	}

	msg := tgbotapi.NewMessage(chatID, textMsg)
	msg.ParseMode = "HTML"

	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send telegram message to %d: %v", chatID, err)
		return err
	}
	return nil
}
