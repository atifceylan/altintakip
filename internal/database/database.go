package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"altintakip/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB global veritabanı bağlantısı
var DB *gorm.DB

// Connect veritabanı bağlantısını kurar
func Connect() error {
	// Kullanıcı dizininde altintakip klasörünü varsayılan yol olarak kullan
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("kullanıcı dizini alınamadı: %w", err)
	}

	appDataDir := getEnv("APP_DATA_DIR", filepath.Join(homeDir, "altintakip"))
	defaultDBPath := filepath.Join(appDataDir, "altintakip.db")

	// SQLite veritabanı dosyasının yolu
	dbPath := getEnv("DB_PATH", defaultDBPath)

	// Veritabanı dosyasının dizinini oluştur
	dbDir := filepath.Dir(dbPath)
	if dbDir != "." && dbDir != "" {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return fmt.Errorf("veritabanı dizini oluşturulamadı: %w", err)
		}
	}

	var dbErr error
	DB, dbErr = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if dbErr != nil {
		return fmt.Errorf("SQLite veritabanı bağlantısı kurulamadı: %w", dbErr)
	}

	log.Printf("SQLite veritabanı bağlantısı başarılı: %s", dbPath)
	return nil
}

// Migrate veritabanı tablolarını oluşturur/günceller
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("veritabanı bağlantısı kurulmamış")
	}

	err := DB.AutoMigrate(&models.Envanter{})
	if err != nil {
		return fmt.Errorf("veritabanı migrasyonu başarısız: %w", err)
	}

	log.Println("Veritabanı migrasyonu tamamlandı")
	return nil
}

// Close veritabanı bağlantısını kapatır
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// getEnv çevre değişkenini alır, yoksa varsayılan değeri döner
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDB veritabanı bağlantısını döner
func GetDB() *gorm.DB {
	return DB
}
