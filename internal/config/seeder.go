package config

import (
	"log"
	"math/rand"
	"time"

	"github.com/panganku/backend/internal/models"
	"github.com/panganku/backend/internal/security"
	"gorm.io/gorm"
)

func SeedData(db *gorm.DB) {
	log.Println("Seeding initial data...")

	// Cek apakah sudah ada admin
	var adminCount int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount)
	if adminCount > 0 {
		log.Println("Admin already exists, skipping seed")
		return
	}

	// Hash password default
	hashedPassword, err := security.HashPassword("Admin123!")
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return
	}

	// Buat admin user
	admin := models.User{
		Name:     "Administrator",
		Email:    "admin@panganku.id",
		Password: hashedPassword,
		Phone:    "081234567890",
		Role:     "admin",
		IsActive: true,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Printf("Failed to create admin: %v", err)
		return
	}

	log.Println("âœ“ Admin user created successfully")
	log.Println("  Email: admin@panganku.id")
	log.Println("  Password: Admin123!")

	// Seed beberapa komoditas
	komoditas := []models.Komoditas{
		{Nama: "Beras", Satuan: "kg", Kategori: "Padi-padian"},
		{Nama: "Jagung", Satuan: "kg", Kategori: "Padi-padian"},
		{Nama: "Kedelai", Satuan: "kg", Kategori: "Kacang-kacangan"},
		{Nama: "Cabai Merah", Satuan: "kg", Kategori: "Sayuran"},
		{Nama: "Bawang Merah", Satuan: "kg", Kategori: "Sayuran"},
		{Nama: "Gula Pasir", Satuan: "kg", Kategori: "Gula"},
		{Nama: "Minyak Goreng", Satuan: "liter", Kategori: "Minyak"},
		{Nama: "Daging Ayam", Satuan: "kg", Kategori: "Protein"},
		{Nama: "Telur Ayam", Satuan: "kg", Kategori: "Protein"},
	}

	for _, k := range komoditas {
		var count int64
		db.Model(&models.Komoditas{}).Where("nama = ?", k.Nama).Count(&count)
		if count == 0 {
			if err := db.Create(&k).Error; err != nil {
				log.Printf("Failed to create komoditas %s: %v", k.Nama, err)
			} else {
				log.Printf("âœ“ Komoditas created: %s", k.Nama)
			}
		}
	}

	// Seed kecamatan di Kabupaten Lamongan
	kecamatan := []models.Kecamatan{
		{Nama: "Babat", Lat: -7.1340, Lng: 112.1640, LuasHa: 8543},
		{Nama: "Bluluk", Lat: -7.3130, Lng: 112.2370, LuasHa: 6921},
		{Nama: "Brondong", Lat: -6.8790, Lng: 112.2920, LuasHa: 6510},
		{Nama: "Lamongan", Lat: -7.1169, Lng: 112.4131, LuasHa: 4725},
		{Nama: "Deket", Lat: -7.0900, Lng: 112.3000, LuasHa: 5234},
		{Nama: "Glagah", Lat: -7.1280, Lng: 112.3360, LuasHa: 5012},
		{Nama: "Kalitengah", Lat: -7.0600, Lng: 112.3500, LuasHa: 5480},
		{Nama: "Karangbinangun", Lat: -7.0210, Lng: 112.3410, LuasHa: 5195},
		{Nama: "Karanggeneng", Lat: -7.0300, Lng: 112.2800, LuasHa: 5678},
		{Nama: "Kedungpring", Lat: -7.1234, Lng: 112.3456, LuasHa: 6789},
		{Nama: "Kembangbahu", Lat: -7.2000, Lng: 112.3200, LuasHa: 6031},
		{Nama: "Laren", Lat: -6.9530, Lng: 112.4310, LuasHa: 5920},
		{Nama: "Maduran", Lat: -7.0050, Lng: 112.3740, LuasHa: 4880},
		{Nama: "Mantup", Lat: -7.2900, Lng: 112.3650, LuasHa: 7420},
		{Nama: "Modo", Lat: -7.2680, Lng: 112.2120, LuasHa: 7124},
		{Nama: "Ngimbang", Lat: -7.2360, Lng: 112.1960, LuasHa: 7540},
		{Nama: "Paciran", Lat: -6.8800, Lng: 112.3700, LuasHa: 5432},
		{Nama: "Pucuk", Lat: -7.1500, Lng: 112.5000, LuasHa: 4567},
		{Nama: "Sambeng", Lat: -7.2400, Lng: 112.2770, LuasHa: 7350},
		{Nama: "Sarirejo", Lat: -7.0500, Lng: 112.3500, LuasHa: 6345},
		{Nama: "Sekaran", Lat: -7.1800, Lng: 112.2100, LuasHa: 6842},
		{Nama: "Solokuro", Lat: -6.9620, Lng: 112.3300, LuasHa: 5892},
		{Nama: "Sugio", Lat: -7.2070, Lng: 112.4200, LuasHa: 6285},
		{Nama: "Sukodadi", Lat: -7.0650, Lng: 112.4100, LuasHa: 5168},
		{Nama: "Sukorame", Lat: -7.3030, Lng: 112.3070, LuasHa: 6980},
		{Nama: "Tikung", Lat: -7.2000, Lng: 112.4500, LuasHa: 7123},
		{Nama: "Turi", Lat: -7.1900, Lng: 112.3800, LuasHa: 5640},
	}

	for _, kec := range kecamatan {
		var count int64
		db.Model(&models.Kecamatan{}).Where("nama = ?", kec.Nama).Count(&count)
		if count == 0 {
			if err := db.Create(&kec).Error; err != nil {
				log.Printf("Failed to create kecamatan %s: %v", kec.Nama, err)
			} else {
				log.Printf("âœ“ Kecamatan created: %s", kec.Nama)
			}
		}
	}

	log.Println("âœ“ Seeding completed successfully")
}

// SeedDummyData mengisi data harga pasar dan stok pangan untuk demo
func SeedDummyData(db *gorm.DB) {
	// Skip jika sudah ada data harga
	var hargaCount int64
	db.Model(&models.HargaPasar{}).Count(&hargaCount)
	if hargaCount > 0 {
		return
	}

	log.Println("Seeding dummy harga & stok pangan data...")

	var admin models.User
	db.Where("role = ?", "admin").First(&admin)

	var komoditasList []models.Komoditas
	db.Find(&komoditasList)
	var kecamatanList []models.Kecamatan
	db.Find(&kecamatanList)

	if len(komoditasList) == 0 || len(kecamatanList) == 0 {
		log.Println("Komoditas/kecamatan belum ada, skip dummy seed")
		return
	}

	// Harga dasar (Rp/kg) per komoditas
	basePrices := map[string]float64{
		"Beras":         12500,
		"Jagung":        5500,
		"Kedelai":       14000,
		"Cabai Merah":   45000,
		"Bawang Merah":  28000,
		"Gula Pasir":    16000,
		"Minyak Goreng": 15500,
		"Daging Ayam":   35000,
		"Telur Ayam":    27000,
	}

	// Variasi harga harian per komoditas (supaya grafik terlihat dinamis)
	dailyDelta := map[string][]float64{
		"Beras":         {0.0, +200, -100, +150, -50, +300, -200},
		"Jagung":        {0.0, -100, +200, -150, +100, -200, +250},
		"Kedelai":       {0.0, +300, -200, +100, -300, +200, -100},
		"Cabai Merah":   {0.0, +2000, -1000, +3000, -2000, +1500, -500},
		"Bawang Merah":  {0.0, -500, +1000, -800, +1200, -600, +400},
		"Gula Pasir":    {0.0, +100, -50, +150, -100, +200, -150},
		"Minyak Goreng": {0.0, -100, +200, -150, +100, -200, +300},
		"Daging Ayam":   {0.0, +500, -300, +800, -500, +300, -200},
		"Telur Ayam":    {0.0, +300, -200, +400, -300, +500, -200},
	}

	// Seed 30 hari harga pasar
	now := time.Now()
	rng := rand.New(rand.NewSource(42))
	for _, kec := range kecamatanList {
		for _, kom := range komoditasList {
			base := basePrices[kom.Nama]
			if base == 0 {
				base = 10000
			}
			deltas := dailyDelta[kom.Nama]
			running := base
			for day := 29; day >= 0; day-- {
				tanggal := now.AddDate(0, 0, -day)
				if day < len(deltas) {
					running += deltas[day]
				}
				// Tambah noise kecil Â±1% per kecamatan
				noise := (rng.Float64()*0.02 - 0.01) * base
				harga := running + noise
				if harga < base*0.7 {
					harga = base * 0.7
				}
				db.Create(&models.HargaPasar{
					KomoditasID: kom.ID,
					KecamatanID: kec.ID,
					HargaPerKg:  harga,
					Tanggal:     tanggal,
					CreatedBy:   admin.ID,
				})
			}
		}
	}
	log.Printf("âœ“ Harga pasar: 30 hari Ã— %d komoditas Ã— %d kecamatan", len(komoditasList), len(kecamatanList))

	// Stok level per kecamatan (supaya terlihat kondisi beragam)
	stokLevel := []float64{0.85, 0.22, 0.60, 0.78, 0.35, 0.72, 0.18, 0.67, 0.45, 0.28}
	for i, kec := range kecamatanList {
		level := stokLevel[i%len(stokLevel)]
		for _, kom := range komoditasList {
			kapasitas := 40000.0 + float64(kec.LuasHa)*0.4
			stok := kapasitas * level
			db.Create(&models.StokPangan{
				KomoditasID: kom.ID,
				KecamatanID: kec.ID,
				StokKg:      stok,
				KapasitasKg: kapasitas,
				PetugasID:   admin.ID,
			})
		}
	}
	log.Printf("âœ“ Stok pangan: %d komoditas Ã— %d kecamatan", len(komoditasList), len(kecamatanList))

	// Beberapa laporan darurat
	var kec0, kec1 models.Kecamatan
	db.Where("nama = ?", "Babat").First(&kec0)
	db.Where("nama = ?", "Tikung").First(&kec1)
	laporanList := []models.LaporanDarurat{
		{PelaporID: admin.ID, KecamatanID: kec0.ID, JenisMasalah: "Kelangkaan Beras", Deskripsi: "Stok beras habis di pasar Babat", Status: "baru", Prioritas: 1},
		{PelaporID: admin.ID, KecamatanID: kec1.ID, JenisMasalah: "Kenaikan Harga Cabai", Deskripsi: "Harga cabai naik drastis melebihi 50%", Status: "proses", Prioritas: 2},
		{PelaporID: admin.ID, KecamatanID: kec0.ID, JenisMasalah: "Distribusi Terlambat", Deskripsi: "Distribusi jagung ke gudang terlambat 3 hari", Status: "selesai", Prioritas: 3},
	}
	for _, lap := range laporanList {
		db.Create(&lap)
	}
	log.Println("âœ“ Laporan darurat dummy berhasil dibuat")
	log.Println("âœ“ Dummy data seeding selesai")
}
