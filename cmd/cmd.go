package cmd

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"altintakip/internal/cli"
	"altintakip/internal/database"
	"altintakip/internal/models"
	"altintakip/internal/services"

	"github.com/joho/godotenv"
)

// Execute ana uygulama mantığını çalıştırır
func Execute() {
	// Command line parametrelerini kontrol et
	listMode := false
	if len(os.Args) > 1 && os.Args[1] == "list" {
		listMode = true
	}

	// .altintakip_env dosyasını yükle - önce mevcut dizin, sonra home dizini
	envFileName := ".altintakip_env"

	// Önce mevcut dizinde ara
	currentDirEnvPath := envFileName
	if err := godotenv.Load(currentDirEnvPath); err != nil {
		// Mevcut dizinde yoksa home dizininde ara
		homeDir, homeErr := os.UserHomeDir()
		if homeErr != nil {
			log.Fatalf("HATA: Home directory alınamadı ve mevcut dizinde %s dosyası bulunamadı: %v", envFileName, homeErr)
		}

		homeDirEnvPath := filepath.Join(homeDir, envFileName)
		if err := godotenv.Load(homeDirEnvPath); err != nil {
			log.Fatalf("HATA: %s dosyası bulunamadı. Arama yerleri:\n- Mevcut dizin: %s\n- Home dizin: %s", envFileName, currentDirEnvPath, homeDirEnvPath)
		}
	}

	// CLI renderer'ı başlat
	renderer := cli.NewTableRenderer()
	renderer.PrintWelcome()

	// Veritabanı bağlantısını kur
	if err := database.Connect(); err != nil {
		renderer.PrintError(err)
		os.Exit(1)
	}
	defer database.Close()

	// Veritabanı migrasyonunu çalıştır
	if err := database.Migrate(); err != nil {
		renderer.PrintError(err)
		os.Exit(1)
	}

	// Servisleri başlat
	envanterService := services.NewEnvanterService()

	// Envanter verisini kontrol et, yoksa örnek veri ekle
	envanterler, err := envanterService.GetAllEnvanter()
	if err != nil {
		renderer.PrintError(err)
		os.Exit(1)
	}

	// Veri yoksa örnek veri ekle
	if len(envanterler) == 0 {
		renderer.PrintInfo("Envanter boş, örnek veriler ekleniyor...")
		if err := addSampleData(envanterService); err != nil {
			renderer.PrintError(err)
			os.Exit(1)
		}
		// Verileri tekrar çek
		envanterler, err = envanterService.GetAllEnvanter()
		if err != nil {
			renderer.PrintError(err)
			os.Exit(1)
		}
	}

	// Liste modu kontrolü
	if listMode {
		renderer.PrintInfo("List modu: Mevcut veriler kullanılıyor (API çağrısı yapılmıyor)")
	} else {
		// API'den fiyatları çek ve envanterleri güncelle
		altinKaynakService := services.NewAltinKaynakService()
		fiyatlar, err := altinKaynakService.GetFiyatlar()
		if err != nil {
			log.Printf("UYARI: fiyatlar alınamadı: %v", err)
			renderer.PrintInfo("API'den fiyat alınamadı, mevcut verilerle devam ediliyor...")
		} else {
			// Envanter fiyatlarını güncelle
			for i := range envanterler {
				guncelFiyat, err := altinKaynakService.GetFiyatByType(fiyatlar, envanterler[i].Kod)
				if err != nil {
					log.Printf("UYARI: %s için fiyat bulunamadı: %v", envanterler[i].Kod, err)
					continue
				}

				// Fiyat bilgilerini güncelle
				envanterler[i].GuncelFiyat = guncelFiyat
				envanterler[i].ToplamAlis = envanterler[i].AlisFiyati * envanterler[i].Miktar
				envanterler[i].GuncelTutar = guncelFiyat * envanterler[i].Miktar
				envanterler[i].KarZarar = envanterler[i].GuncelTutar - envanterler[i].ToplamAlis

				if envanterler[i].ToplamAlis > 0 {
					envanterler[i].KarZararYuzde = (envanterler[i].KarZarar / envanterler[i].ToplamAlis) * 100
				}

				// Veritabanını güncelle
				if err := envanterService.UpdateEnvanter(&envanterler[i]); err != nil {
					log.Printf("UYARI: envanter güncellenirken hata: %v", err)
				}
			}
		}
	}

	// Tabloları görüntüle
	renderer.RenderEnvanterTable(envanterler, nil)

	// Grup analizini görüntüle
	gruplar, err := envanterService.GetKodBazliGruplar()
	if err != nil {
		renderer.PrintError(err)
		os.Exit(1)
	}
	renderer.RenderKodBazliGrupTable(gruplar)

	// Toplam hesaplamaları
	var toplamAlis, toplamGuncel, toplamKar float64
	for _, envanter := range envanterler {
		toplamAlis += envanter.ToplamAlis
		toplamGuncel += envanter.GuncelTutar
		toplamKar += envanter.KarZarar
	}

	var toplamKarYuzde float64
	if toplamAlis > 0 {
		toplamKarYuzde = (toplamKar / toplamAlis) * 100
	}

	toplamlar := map[string]float64{
		"toplam_alis":      toplamAlis,
		"toplam_guncel":    toplamGuncel,
		"toplam_kar":       toplamKar,
		"toplam_kar_yuzde": toplamKarYuzde,
	}

	// Özet tablosunu görüntüle
	renderer.RenderSummary(toplamlar)
}

// addSampleData örnek envanter verilerini ekler
func addSampleData(service *services.EnvanterService) error {
	sampleData := []models.Envanter{
		// Altın Ürünleri
		{
			Tur:        "Altın",
			Cins:       "22 Ayar Bilezik",
			Kod:        "B", // 22 Ayar Bilezik
			Miktar:     15.500,
			Birim:      "gram",
			AlisTarihi: time.Date(2025, 1, 15, 0, 0, 0, 0, time.Local),
			AlisFiyati: 2400.00,
			Notlar:     "22 ayar bilezik takısı",
		},
		{
			Tur:        "Altın",
			Cins:       "22 Ayar Hurda",
			Kod:        "B_T", // 22 Ayar Hurda
			Miktar:     8.250,
			Birim:      "gram",
			AlisTarihi: time.Date(2025, 2, 10, 0, 0, 0, 0, time.Local),
			AlisFiyati: 2200.00,
			Notlar:     "Hurda altın alımı",
		},
		{
			Tur:        "Altın",
			Cins:       "Çeyrek Altın",
			Kod:        "C", // Çeyrek Altın
			Miktar:     8.000,
			Birim:      "adet",
			AlisTarihi: time.Date(2025, 2, 20, 0, 0, 0, 0, time.Local),
			AlisFiyati: 2800.00,
			Notlar:     "Çeyrek altın yatırımı",
		},
		{
			Tur:        "Altın",
			Cins:       "Külçe Toptan",
			Kod:        "CH_T", // Külçe Toptan
			Miktar:     50.000,
			Birim:      "gram",
			AlisTarihi: time.Date(2025, 3, 5, 0, 0, 0, 0, time.Local),
			AlisFiyati: 2350.00,
			Notlar:     "Külçe altın toptan alım",
		},
		{
			Tur:        "Altın",
			Cins:       "Yarım Altın",
			Kod:        "Y", // Yarım Altın
			Miktar:     4.000,
			Birim:      "adet",
			AlisTarihi: time.Date(2025, 3, 15, 0, 0, 0, 0, time.Local),
			AlisFiyati: 5600.00,
			Notlar:     "Yarım altın yatırımı",
		},
		{
			Tur:        "Altın",
			Cins:       "Tam Altın",
			Kod:        "T", // Tam Altın (Teklik)
			Miktar:     3.000,
			Birim:      "adet",
			AlisTarihi: time.Date(2025, 3, 25, 0, 0, 0, 0, time.Local),
			AlisFiyati: 11200.00,
			Notlar:     "Cumhuriyet altını",
		},
		{
			Tur:        "Altın",
			Cins:       "22 Ayar Gram",
			Kod:        "GA", // Gram Altın
			Miktar:     25.750,
			Birim:      "gram",
			AlisTarihi: time.Date(2025, 4, 1, 0, 0, 0, 0, time.Local),
			AlisFiyati: 2300.00,
			Notlar:     "22 ayar gram altın",
		},
		{
			Tur:        "Altın",
			Cins:       "22 Ayar Külçe",
			Kod:        "GAT", // Gram Toptan (Külçe)
			Miktar:     100.000,
			Birim:      "gram",
			AlisTarihi: time.Date(2024, 4, 10, 0, 0, 0, 0, time.Local),
			AlisFiyati: 2280.00,
			Notlar:     "Külçe altın toptan",
		},
		{
			Tur:        "Altın",
			Cins:       "24 Ayar Gram",
			Kod:        "HH_T", // Has Toptan (24 ayar)
			Miktar:     35.500,
			Birim:      "gram",
			AlisTarihi: time.Date(2024, 4, 20, 0, 0, 0, 0, time.Local),
			AlisFiyati: 2450.00,
			Notlar:     "24 ayar has altın",
		},
		// Döviz Ürünleri
		{
			Tur:        "Döviz",
			Cins:       "USD",
			Kod:        "USD", // Amerikan Doları
			Miktar:     1500.000,
			Birim:      "adet",
			AlisTarihi: time.Date(2025, 1, 5, 0, 0, 0, 0, time.Local),
			AlisFiyati: 30.50,
			Notlar:     "Dolar rezervi",
		},
		{
			Tur:        "Döviz",
			Cins:       "EUR",
			Kod:        "EUR", // Avrupa Para Birimi
			Miktar:     800.000,
			Birim:      "adet",
			AlisTarihi: time.Date(2025, 2, 1, 0, 0, 0, 0, time.Local),
			AlisFiyati: 33.20,
			Notlar:     "Euro yatırımı",
		},
		{
			Tur:        "Döviz",
			Cins:       "GBP",
			Kod:        "GBP", // İngiliz Sterlini
			Miktar:     300.000,
			Birim:      "adet",
			AlisTarihi: time.Date(2025, 2, 15, 0, 0, 0, 0, time.Local),
			AlisFiyati: 38.75,
			Notlar:     "Sterlin yatırımı",
		},
	}

	for _, item := range sampleData {
		if err := service.AddEnvanter(&item); err != nil {
			return err
		}
	}

	return nil
}
