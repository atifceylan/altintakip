package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"altintakip/internal/database"
	"altintakip/internal/tui"

	"github.com/joho/godotenv"
)

// Execute ana uygulama mantığını çalıştırır
func Execute() {
	// Komut satırı parametrelerini kontrol et
	args := os.Args[1:]
	isListMode := false
	for _, arg := range args {
		if arg == "list" {
			isListMode = true
			break
		}
	}

	// Kullanıcı dizininde altintakip klasörünü oluştur
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Kullanıcı dizini alınamadı: %v", err)
	}

	appDataDir := getEnv("APP_DATA_DIR", filepath.Join(homeDir, "altintakip"))
	if err := os.MkdirAll(appDataDir, 0755); err != nil {
		log.Fatalf("Uygulama veri dizini oluşturulamadı: %v", err)
	}

	// .altintakip_env dosyasını yükle - önce mevcut dizin, sonra home dizini, sonra app data dizini
	envFileName := ".altintakip_env"

	// Önce mevcut dizinde ara
	currentDirEnvPath := envFileName
	if err := godotenv.Load(currentDirEnvPath); err != nil {
		// Mevcut dizinde yoksa home dizininde ara
		if homeDir != "" {
			homeDirEnvPath := filepath.Join(homeDir, envFileName)
			if err := godotenv.Load(homeDirEnvPath); err != nil {
				// Home dizininde de yoksa app data dizininde ara
				appDataEnvPath := filepath.Join(appDataDir, envFileName)
				// App data dizininde de yoksa sessizce devam et (SQLite için env zorunlu değil)
				_ = godotenv.Load(appDataEnvPath)
			}
		}
	}

	// Log dosyası yolunu belirle ve log çıktılarını dosyaya yönlendir
	logPath := getEnv("LOG_PATH", filepath.Join(appDataDir, "altintakip.log"))
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	} else {
		// Log dosyası açılamazsa, log'ları tamamen sustur
		log.SetOutput(io.Discard)
	}

	// Veritabanı bağlantısını kur
	if err := database.Connect(); err != nil {
		log.Fatalf("Veritabanı bağlantısı kurulamadı: %v", err)
	}
	defer database.Close()

	// Veritabanı migrasyonunu çalıştır
	if err := database.Migrate(); err != nil {
		log.Fatalf("Veritabanı migrasyonu başarısız: %v", err)
	}

	// TUI uygulamasını başlat
	app := tui.NewApp()
	if isListMode {
		fmt.Printf("Liste modu: Sadece veritabanındaki veriler gösterilecek (fiyat güncellenmeyecek)...\n")
		app.SetListMode(true)
	} else {
		fmt.Printf("Uygulama başlatılıyor, güncel fiyatlar çekiliyor...\n")
		app.SetListMode(false)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("TUI uygulaması başlatılamadı: %v", err)
	}
}

// getEnv çevre değişkenini alır, yoksa varsayılan değeri döner
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
