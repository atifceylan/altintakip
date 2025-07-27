package database

import (
	"fmt"
	"log"
	"os"

	"altintakip/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB global veritabanı bağlantısı
var DB *gorm.DB

// Config veritabanı yapılandırma yapısı
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Connect veritabanı bağlantısını kurar
func Connect() error {
	config := Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", ""),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "altintakip"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return fmt.Errorf("veritabanı bağlantısı kurulamadı: %w", err)
	}

	log.Println("Veritabanı bağlantısı başarılı")
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
