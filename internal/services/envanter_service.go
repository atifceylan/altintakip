package services

import (
	"fmt"
	"log"

	"altintakip/internal/database"
	"altintakip/internal/models"
)

// EnvanterService envanter işlemlerini yönetir
type EnvanterService struct {
	altinService *AltinKaynakService
}

// NewEnvanterService yeni envanter servisi oluşturur
func NewEnvanterService() *EnvanterService {
	return &EnvanterService{
		altinService: NewAltinKaynakService(),
	}
}

// GetAllEnvanter tüm envanter kayıtlarını getirir
func (s *EnvanterService) GetAllEnvanter() ([]models.Envanter, error) {
	var envanter []models.Envanter

	err := database.GetDB().Order("tur asc, alis_tarihi asc").Find(&envanter).Error
	if err != nil {
		return nil, fmt.Errorf("envanter kayıtları getirilemedi: %w", err)
	}

	return envanter, nil
}

// UpdateGuncelFiyatlar tüm envanter için güncel fiyatları günceller
func (s *EnvanterService) UpdateGuncelFiyatlar() error {
	log.Println("Güncel fiyatlar alınıyor...")

	// API'den güncel fiyatları al
	fiyatlar, err := s.altinService.GetFiyatlar()
	if err != nil {
		return fmt.Errorf("fiyatlar alınamadı: %w", err)
	}

	log.Printf("Fiyatlar başarıyla alındı. Güncelleme tarihi: %s", fiyatlar.GuncellemeTarihi.Format("2006-01-02 15:04:05"))

	// Tüm envanter kayıtlarını al
	var envanterler []models.Envanter
	err = database.GetDB().Find(&envanterler).Error
	if err != nil {
		return fmt.Errorf("envanter kayıtları getirilemedi: %w", err)
	}

	// Her envanter için güncel fiyatı güncelle
	for i := range envanterler {
		guncelFiyat, err := s.altinService.GetFiyatByType(fiyatlar, envanterler[i].Kod)
		if err != nil {
			log.Printf("UYARI: Kod %s için fiyat bulunamadı: %v", envanterler[i].Kod, err)
			continue
		}

		envanterler[i].GuncelFiyat = guncelFiyat
		//envanterler[i].APIKaynak = "data.altinkaynak.com"
		envanterler[i].GuncelDegerleriHesapla()

		// Veritabanında güncelle
		err = database.GetDB().Save(&envanterler[i]).Error
		if err != nil {
			log.Printf("HATA: Envanter ID %d güncellenemedi: %v", envanterler[i].ID, err)
		}
	}

	log.Printf("Toplam %d envanter kaydı güncellendi", len(envanterler))
	return nil
}

// AddEnvanter yeni envanter kaydı ekler
func (s *EnvanterService) AddEnvanter(envanter *models.Envanter) error {
	return s.AddEnvanterWithMode(envanter, false)
}

// AddEnvanterWithMode yeni envanter kaydı ekler (liste modu parametreli)
func (s *EnvanterService) AddEnvanterWithMode(envanter *models.Envanter, isListMode bool) error {
	// Toplam alış tutarını hesapla
	envanter.ToplamAlis = envanter.Miktar * envanter.AlisFiyati

	// Liste modunda değilse ve güncel fiyat girilmemişse (0 ise) API'den çek
	if !isListMode && envanter.GuncelFiyat == 0 {
		fiyatlar, err := s.altinService.GetFiyatlar()
		if err != nil {
			log.Printf("UYARI: Güncel fiyat alınamadı, sadece alış bilgileri kaydediliyor: %v", err)
		} else {
			guncelFiyat, err := s.altinService.GetFiyatByType(fiyatlar, envanter.Kod)
			if err == nil {
				envanter.GuncelFiyat = guncelFiyat
				log.Printf("Güncel fiyat API'den çekildi: %s = %.2f", envanter.Kod, guncelFiyat)
			} else {
				log.Printf("UYARI: Kod %s için güncel fiyat bulunamadı: %v", envanter.Kod, err)
			}
		}
	} else if isListMode {
		log.Printf("Liste modu: Güncel fiyat API'den çekilmeyecek")
	}

	// Güncel değerleri hesapla
	envanter.GuncelDegerleriHesapla()

	err := database.GetDB().Create(envanter).Error
	if err != nil {
		return fmt.Errorf("envanter kaydedilemedi: %w", err)
	}

	log.Printf("Yeni envanter kaydı eklendi: %s %s (%.2f %s)", envanter.Tur, envanter.Cins, envanter.Miktar, envanter.Birim)
	return nil
}

// DeleteEnvanter envanter kaydını siler
func (s *EnvanterService) DeleteEnvanter(id uint) error {
	err := database.GetDB().Delete(&models.Envanter{}, id).Error
	if err != nil {
		return fmt.Errorf("envanter silinemedi: %w", err)
	}

	log.Printf("Envanter kaydı silindi: ID %d", id)
	return nil
}

// UpdateEnvanter envanter kaydını günceller
func (s *EnvanterService) UpdateEnvanter(envanter *models.Envanter) error {
	return s.UpdateEnvanterWithMode(envanter, false)
}

// UpdateEnvanterWithMode envanter kaydını günceller (liste modu parametreli)
func (s *EnvanterService) UpdateEnvanterWithMode(envanter *models.Envanter, isListMode bool) error {
	// Toplam alış tutarını yeniden hesapla
	envanter.ToplamAlis = envanter.Miktar * envanter.AlisFiyati

	// Liste modunda değilse ve güncel fiyat 0 ise API'den çek
	if !isListMode && envanter.GuncelFiyat == 0 {
		fiyatlar, err := s.altinService.GetFiyatlar()
		if err != nil {
			log.Printf("UYARI: Güncel fiyat alınamadı: %v", err)
		} else {
			guncelFiyat, err := s.altinService.GetFiyatByType(fiyatlar, envanter.Kod)
			if err == nil {
				envanter.GuncelFiyat = guncelFiyat
				log.Printf("Güncel fiyat API'den çekildi: %s = %.2f", envanter.Kod, guncelFiyat)
			} else {
				log.Printf("UYARI: Kod %s için güncel fiyat bulunamadı: %v", envanter.Kod, err)
			}
		}
	} else if isListMode {
		log.Printf("Liste modu: Güncel fiyat API'den çekilmeyecek")
	}

	// Güncel değerleri hesapla
	envanter.GuncelDegerleriHesapla()

	err := database.GetDB().Save(envanter).Error
	if err != nil {
		return fmt.Errorf("envanter güncellenemedi: %w", err)
	}

	log.Printf("Envanter kaydı güncellendi: %s %s (%.2f %s)", envanter.Tur, envanter.Cins, envanter.Miktar, envanter.Birim)
	return nil
}

// GetToplamDegerler toplam değerleri hesaplar
func (s *EnvanterService) GetToplamDegerler() (map[string]float64, error) {
	var envanterler []models.Envanter
	err := database.GetDB().Find(&envanterler).Error
	if err != nil {
		return nil, fmt.Errorf("envanter kayıtları getirilemedi: %w", err)
	}

	toplamlar := map[string]float64{
		"toplam_alis":   0,
		"toplam_guncel": 0,
		"toplam_kar":    0,
	}

	for _, envanter := range envanterler {
		toplamlar["toplam_alis"] += envanter.ToplamAlis
		toplamlar["toplam_guncel"] += envanter.GuncelTutar
		toplamlar["toplam_kar"] += envanter.KarZarar
	}

	// Toplam kar/zarar yüzdesi
	if toplamlar["toplam_alis"] > 0 {
		toplamlar["toplam_kar_yuzde"] = (toplamlar["toplam_kar"] / toplamlar["toplam_alis"]) * 100
	}

	return toplamlar, nil
}

// GetKodBazliGruplar kod bazlı gruplu verileri hesaplar
func (s *EnvanterService) GetKodBazliGruplar() (map[string]map[string]interface{}, error) {
	var envanterler []models.Envanter
	err := database.GetDB().Order("tur asc, cins asc").Find(&envanterler).Error
	if err != nil {
		return nil, fmt.Errorf("envanter kayıtları getirilemedi: %w", err)
	}

	gruplar := make(map[string]map[string]interface{})

	// Kod bazlı gruplama
	for _, envanter := range envanterler {
		kod := envanter.Kod
		if kod == "" {
			kod = "TANIMSIZ"
		}

		if _, exists := gruplar[kod]; !exists {
			gruplar[kod] = map[string]interface{}{
				"tur":                 envanter.Tur,
				"cins":                envanter.Cins,
				"birim":               envanter.Birim,
				"toplam_miktar":       0.0,
				"toplam_alis_tutar":   0.0,
				"toplam_guncel_tutar": 0.0,
				"toplam_kar_zarar":    0.0,
				"adet":                0,
			}
		}

		grup := gruplar[kod]
		grup["toplam_miktar"] = grup["toplam_miktar"].(float64) + envanter.Miktar
		grup["toplam_alis_tutar"] = grup["toplam_alis_tutar"].(float64) + envanter.ToplamAlis
		grup["toplam_guncel_tutar"] = grup["toplam_guncel_tutar"].(float64) + envanter.GuncelTutar
		grup["toplam_kar_zarar"] = grup["toplam_kar_zarar"].(float64) + envanter.KarZarar
		grup["adet"] = grup["adet"].(int) + 1
	}

	// Ortalama değerleri hesapla
	for kod, grup := range gruplar {
		adet := grup["adet"].(int)
		if adet > 0 {
			grup["ortalama_alis_fiyati"] = grup["toplam_alis_tutar"].(float64) / grup["toplam_miktar"].(float64)
			grup["ortalama_kar_zarar"] = grup["toplam_kar_zarar"].(float64) / float64(adet)

			// Kar/zarar yüzdesi
			if grup["toplam_alis_tutar"].(float64) > 0 {
				grup["kar_zarar_yuzde"] = (grup["toplam_kar_zarar"].(float64) / grup["toplam_alis_tutar"].(float64)) * 100
			} else {
				grup["kar_zarar_yuzde"] = 0.0
			}
		}
		gruplar[kod] = grup
	}

	return gruplar, nil
}

// GetAllEnvanterFromDB sadece veritabanından envanter kayıtlarını getirir (API çağrısı yapmaz)
func (s *EnvanterService) GetAllEnvanterFromDB() ([]models.Envanter, error) {
	var envanter []models.Envanter

	err := database.GetDB().Order("tur asc, alis_tarihi asc").Find(&envanter).Error
	if err != nil {
		return nil, fmt.Errorf("envanter kayıtları getirilemedi: %w", err)
	}

	log.Printf("Veritabanından %d envanter kaydı alındı", len(envanter))
	return envanter, nil
}

// GetEnvanterByID ID'ye göre envanter kaydını getirir
func (s *EnvanterService) GetEnvanterByID(id uint) (*models.Envanter, error) {
	var envanter models.Envanter
	err := database.GetDB().First(&envanter, id).Error
	if err != nil {
		return nil, fmt.Errorf("envanter kaydı bulunamadı: %w", err)
	}

	return &envanter, nil
}
