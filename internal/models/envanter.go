package models

import (
	"time"

	"gorm.io/gorm"
)

// Envanter altın ve döviz envanterini temsil eder
type Envanter struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Ürün bilgileri
	Tur    string  `gorm:"not null" json:"tur"`    // "Altın" veya "Döviz"
	Cins   string  `gorm:"not null" json:"cins"`   // "22 Ayar Külçe", "Cumhuriyet Altını", "USD", "EUR" vb.
	Kod    string  `gorm:"not null" json:"kod"`    // API'deki kod: "GAT", "C", "USD", "EUR" vb.
	Miktar float64 `gorm:"not null" json:"miktar"` // Gram veya adet
	Birim  string  `gorm:"not null" json:"birim"`  // "gram", "adet"

	// Alış bilgileri
	AlisTarihi time.Time `gorm:"not null" json:"alis_tarihi"`
	AlisFiyati float64   `gorm:"not null" json:"alis_fiyati"` // Birim başına alış fiyatı (TL)
	ToplamAlis float64   `gorm:"not null" json:"toplam_alis"` // Toplam alış tutarı (TL)

	// Güncel değer bilgileri (her çalıştırmada güncellenir)
	GuncelFiyat float64 `json:"guncel_fiyat"` // Birim başına güncel fiyat (TL)
	GuncelTutar float64 `json:"guncel_tutar"` // Toplam güncel tutar (TL)

	// Kar/Zarar
	KarZarar      float64 `json:"kar_zarar"`       // Güncel tutar - Toplam alış
	KarZararYuzde float64 `json:"kar_zarar_yuzde"` // (Kar/Zarar / Toplam alış) * 100

	// API kaynak bilgisi
	APIKaynak string `json:"api_kaynak,omitempty"` // Hangi API'den fiyat alındığı

	// Notlar
	Notlar string `json:"notlar,omitempty"`
}

// TableName GORM için tablo adını belirtir
func (Envanter) TableName() string {
	return "envanter"
}

// GuncelDegerleriHesapla güncel fiyata göre değerleri hesaplar
func (e *Envanter) GuncelDegerleriHesapla() {
	e.GuncelTutar = e.Miktar * e.GuncelFiyat
	e.KarZarar = e.GuncelTutar - e.ToplamAlis
	if e.ToplamAlis > 0 {
		e.KarZararYuzde = (e.KarZarar / e.ToplamAlis) * 100
	}
}
