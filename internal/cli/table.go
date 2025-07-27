package cli

import (
	"fmt"
	"time"

	"altintakip/internal/models"
)

// TableRenderer CLI tablo görüntüleme işlemlerini yönetir
type TableRenderer struct{}

// NewTableRenderer yeni tablo renderer oluşturur
func NewTableRenderer() *TableRenderer {
	return &TableRenderer{}
}

// RenderEnvanterTable envanter tablosunu görüntüler
func (tr *TableRenderer) RenderEnvanterTable(envanterler []models.Envanter, toplamlar map[string]float64) {
	if len(envanterler) == 0 {
		fmt.Println("📦 Envanterde kayıtlı ürün bulunmuyor.")
		fmt.Println("💡 Yeni ürün eklemek için uygulama kodunu güncelleyin.")
		return
	}

	fmt.Println("💰 KİŞİSEL ALTIN VE DÖVİZ ENVANTERİ")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════════════════════")
	fmt.Println()

	// Header
	fmt.Printf("%-8s %-20s %-10s %-8s %-12s %-12s %-14s %-14s %-16s %-14s %-12s\n",
		"TÜR", "CİNS", "MİKTAR", "BİRİM", "ALIŞ TARİHİ", "ALIŞ FİYATI (₺)", "TOPLAM ALIŞ (₺)", "GÜNCEL FİYAT (₺)", "GÜNCEL TUTAR (₺)", "KAR/ZARAR (₺)", "KAR/ZARAR %")

	fmt.Println("───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")

	// Satırları yazdır
	for _, envanter := range envanterler {
		karZararStr := tr.formatMoneyWithThousands(envanter.KarZarar)
		karZararYuzdeStr := tr.formatPercentage(envanter.KarZararYuzde)

		// Kar/zarar durumuna göre işaret ekleme
		if envanter.KarZarar >= 0 {
			karZararStr = "+" + karZararStr
			karZararYuzdeStr = "+" + karZararYuzdeStr
		}

		fmt.Printf("%-8s %-20s %-10s %-8s %-12s %-12s %-14s %-14s %-16s %-14s %-12s\n",
			envanter.Tur,
			envanter.Cins,
			tr.formatMiktar(envanter.Miktar, envanter.Birim),
			envanter.Birim,
			envanter.AlisTarihi.Format("02.01.2006"),
			tr.formatMoneyWithThousands(envanter.AlisFiyati),
			tr.formatMoneyWithThousands(envanter.ToplamAlis),
			tr.formatMoneyWithThousands(envanter.GuncelFiyat),
			tr.formatMoneyWithThousands(envanter.GuncelTutar),
			karZararStr,
			karZararYuzdeStr,
		)
	}

	fmt.Println("───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
}

// RenderSummary özet bilgilerini görüntüler
func (tr *TableRenderer) RenderSummary(toplamlar map[string]float64) {
	fmt.Println()
	fmt.Println("📊 ÖZET BİLGİLER")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════════════════════")
	fmt.Println()

	// Özet hesaplama
	toplamKarZarar := toplamlar["toplam_kar"]
	toplamKarZararYuzde := toplamlar["toplam_kar_yuzde"]

	karZararStr := tr.formatMoneyWithThousands(toplamKarZarar)
	karZararYuzdeStr := tr.formatPercentage(toplamKarZararYuzde)

	if toplamKarZarar >= 0 {
		karZararStr = "+" + karZararStr
		karZararYuzdeStr = "+" + karZararYuzdeStr
	}

	// Özet tablosu
	fmt.Printf("%-25s %-25s %-20s %-20s\n",
		"TOPLAM ALIŞ TUTARI (₺)", "TOPLAM GÜNCEL TUTAR (₺)", "TOPLAM KAR/ZARAR (₺)", "TOPLAM KAR/ZARAR %")
	fmt.Println("─────────────────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("%-25s %-25s %-20s %-20s\n",
		tr.formatMoneyWithThousands(toplamlar["toplam_alis"]),
		tr.formatMoneyWithThousands(toplamlar["toplam_guncel"]),
		karZararStr,
		karZararYuzdeStr,
	)

	fmt.Println()
	fmt.Printf("🕐 Son güncelleme: %s\n", time.Now().Format("15:04:05 - 02.01.2006"))
	fmt.Println("📊 Veriler data.altinkaynak.com'dan alınmaktadır")
}

// RenderKodBazliGrupTable kod bazlı grup tablosunu görüntüler
func (tr *TableRenderer) RenderKodBazliGrupTable(gruplar map[string]map[string]interface{}) {
	fmt.Println()
	fmt.Println("📋 GRUP ANALİZİ")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════════════════════")
	fmt.Println()

	// Header
	fmt.Printf("%-8s %-15s %-12s %-8s %-15s %-15s %-15s %-15s %-12s\n",
		"TÜR", "CİNS", "TOPLAM MİKTAR", "BİRİM", "ORT. ALIŞ FİYATI (₺)", "TOPLAM ALIŞ (₺)", "TOPLAM GÜNCEL (₺)", "TOPLAM KAR/ZARAR (₺)", "KAR/ZARAR %")

	fmt.Println("─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")

	// Kod bazlı sıralı gösterim için slice oluştur
	type KodGrup struct {
		Kod  string
		Veri map[string]interface{}
	}

	var kodGruplar []KodGrup
	for kod, veri := range gruplar {
		kodGruplar = append(kodGruplar, KodGrup{Kod: kod, Veri: veri})
	}

	// Tür ve birim'e göre sırala (tür asc, birim asc)
	for i := 0; i < len(kodGruplar)-1; i++ {
		for j := i + 1; j < len(kodGruplar); j++ {
			// Önce türe göre karşılaştır
			turI := kodGruplar[i].Veri["tur"].(string)
			turJ := kodGruplar[j].Veri["tur"].(string)

			if turI > turJ {
				kodGruplar[i], kodGruplar[j] = kodGruplar[j], kodGruplar[i]
			} else if turI == turJ {
				// Tür aynıysa birime göre karşılaştır
				birimI := kodGruplar[i].Veri["birim"].(string)
				birimJ := kodGruplar[j].Veri["birim"].(string)

				if birimI > birimJ {
					kodGruplar[i], kodGruplar[j] = kodGruplar[j], kodGruplar[i]
				}
			}
		}
	}

	// Satırları yazdır
	for _, kodGrup := range kodGruplar {
		veri := kodGrup.Veri

		karZararStr := tr.formatMoneyWithThousands(veri["toplam_kar_zarar"].(float64))
		karZararYuzdeStr := tr.formatPercentage(veri["kar_zarar_yuzde"].(float64))

		// Kar/zarar durumuna göre işaret ekleme
		if veri["toplam_kar_zarar"].(float64) >= 0 {
			karZararStr = "+" + karZararStr
			karZararYuzdeStr = "+" + karZararYuzdeStr
		}

		fmt.Printf("%-8s %-15s %-12s %-8s %-15s %-15s %-15s %-15s %-12s\n",
			veri["tur"].(string),
			veri["cins"].(string),
			tr.formatMiktar(veri["toplam_miktar"].(float64), veri["birim"].(string)),
			veri["birim"].(string),
			tr.formatMoneyWithThousands(veri["ortalama_alis_fiyati"].(float64)),
			tr.formatMoneyWithThousands(veri["toplam_alis_tutar"].(float64)),
			tr.formatMoneyWithThousands(veri["toplam_guncel_tutar"].(float64)),
			karZararStr,
			karZararYuzdeStr,
		)
	}

	fmt.Println("─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
}

// formatMoney para formatını düzenler
func (tr *TableRenderer) formatMoney(amount float64) string {
	return fmt.Sprintf("%.2f ₺", amount)
}

// formatMoneyWithThousands para formatını binli gruplarla düzenler
func (tr *TableRenderer) formatMoneyWithThousands(amount float64) string {
	// Sayıyı string'e çevir
	str := fmt.Sprintf("%.2f", amount)

	// Ondalık kısmı ayır
	dotIndex := -1
	for i, char := range str {
		if char == '.' {
			dotIndex = i
			break
		}
	}

	var intPart, decPart string
	if dotIndex == -1 {
		intPart = str
		decPart = "00"
	} else {
		intPart = str[:dotIndex]
		decPart = str[dotIndex+1:]
	}

	// Tam kısmı binli gruplara ayır
	var result string
	for i, char := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result += "."
		}
		result += string(char)
	}

	// Ondalık kısım 00 ise ondalık gösterme
	if decPart == "00" {
		return result
	} else {
		return fmt.Sprintf("%s,%s", result, decPart)
	}
}

// formatFloat float sayıyı formatlar
func (tr *TableRenderer) formatFloat(num float64) string {
	return fmt.Sprintf("%.3f", num)
}

// formatMiktar birim türüne göre miktar formatlar
func (tr *TableRenderer) formatMiktar(miktar float64, birim string) string {
	if birim == "gram" {
		return fmt.Sprintf("%.2f", miktar)
	} else {
		// adet için tam sayı
		return fmt.Sprintf("%.0f", miktar)
	}
}

// formatPercentage yüzde formatını düzenler
func (tr *TableRenderer) formatPercentage(percent float64) string {
	return fmt.Sprintf("%.2f%%", percent)
}

// PrintError hata mesajını görüntüler
func (tr *TableRenderer) PrintError(err error) {
	fmt.Printf("❌ HATA: %v\n", err)
}

// PrintSuccess başarı mesajını görüntüler
func (tr *TableRenderer) PrintSuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}

// PrintInfo bilgi mesajını görüntüler
func (tr *TableRenderer) PrintInfo(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// PrintWelcome karşılama mesajını görüntüler
func (tr *TableRenderer) PrintWelcome() {
	fmt.Println("🏦 ALTIN TAKİP - Kişisel Altın ve Döviz Envanter Takip Sistemi")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════════════════════")
	fmt.Println("📈 Güncel fiyatlar alınıyor ve envanter güncelleniyor...")
	fmt.Println()
}
