package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// AltinKaynakService altın kaynak API servisi
type AltinKaynakService struct {
	baseURL string
	client  *http.Client
}

// RestPriceItem REST API'den gelen tek bir ürün fiyat bilgisi
type RestPriceItem struct {
	ID                int      `json:"Id"`
	Kod               string   `json:"Kod"`
	Aciklama          string   `json:"Aciklama"`
	Alis              string   `json:"Alis"`
	Satis             string   `json:"Satis"`
	GuncellenmeZamani string   `json:"GuncellenmeZamani"`
	Durum             *string  `json:"Durum"`
	Main              bool     `json:"Main"`
	DataGroup         int      `json:"DataGroup"`
	Change            *float64 `json:"Change"`
	MobilAciklama     string   `json:"MobilAciklama"`
	WebGroup          *int     `json:"WebGroup"`
	WidgetAciklama    string   `json:"WidgetAciklama"`
}

// AltinFiyatlari REST API'den gelen fiyat bilgilerini tutar
type AltinFiyatlari struct {
	// REST API'den gelen tüm veriler
	GoldItems        []RestPriceItem `json:"gold_items"`
	CurrencyItems    []RestPriceItem `json:"currency_items"`
	GuncellemeTarihi time.Time
}

// NewAltinKaynakService yeni servis instance'ı oluşturur
func NewAltinKaynakService() *AltinKaynakService {
	return &AltinKaynakService{
		baseURL: "https://rest.altinkaynak.com",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetFiyatlar altın ve döviz fiyatlarını REST API'den çeker
func (s *AltinKaynakService) GetFiyatlar() (*AltinFiyatlari, error) {
	fiyatlar := &AltinFiyatlari{
		GuncellemeTarihi: time.Now(),
	}

	// Altın fiyatlarını çek
	goldItems, err := s.getRestData("/Gold.json")
	if err != nil {
		return nil, fmt.Errorf("altın fiyatları çekme hatası: %w", err)
	}
	fiyatlar.GoldItems = goldItems

	// Döviz fiyatlarını çek
	currencyItems, err := s.getRestData("/Currency.json")
	if err != nil {
		return nil, fmt.Errorf("döviz fiyatları çekme hatası: %w", err)
	}
	fiyatlar.CurrencyItems = currencyItems

	return fiyatlar, nil
}

// getRestData REST API'den JSON veri çeker
func (s *AltinKaynakService) getRestData(endpoint string) ([]RestPriceItem, error) {
	url := s.baseURL + endpoint

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("HTTP request oluşturma hatası: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "AltinTakip/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API çağrısı başarısız: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API hatası: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("response okuma hatası: %w", err)
	}

	var items []RestPriceItem
	err = json.Unmarshal(body, &items)
	if err != nil {
		return nil, fmt.Errorf("JSON parse hatası: %w", err)
	}

	return items, nil
}

// GetFiyatByType belirli bir ürün kodu için fiyat döner
func (s *AltinKaynakService) GetFiyatByType(fiyatlar *AltinFiyatlari, kod string) (float64, error) {
	kod = strings.ToUpper(strings.TrimSpace(kod))

	// Altın verilerinde ara
	for _, item := range fiyatlar.GoldItems {
		if strings.ToUpper(item.Kod) == kod {
			return parseFloat(item.Alis), nil
		}
	}

	// Döviz verilerinde ara
	for _, item := range fiyatlar.CurrencyItems {
		if strings.ToUpper(item.Kod) == kod {
			return parseFloat(item.Alis), nil
		}
	}

	return 0, fmt.Errorf("bilinmeyen ürün kodu: %s", kod)
}

// GetAllItems tüm mevcut ürünleri MobilAciklama ile birlikte döner
func (s *AltinKaynakService) GetAllItems(fiyatlar *AltinFiyatlari) []RestPriceItem {
	var allItems []RestPriceItem
	allItems = append(allItems, fiyatlar.GoldItems...)
	allItems = append(allItems, fiyatlar.CurrencyItems...)
	return allItems
}

// GetItemByCode belirli bir kod için ürün bilgisini döner
func (s *AltinKaynakService) GetItemByCode(fiyatlar *AltinFiyatlari, kod string) (*RestPriceItem, error) {
	kod = strings.ToUpper(strings.TrimSpace(kod))

	// Altın verilerinde ara
	for _, item := range fiyatlar.GoldItems {
		if strings.ToUpper(item.Kod) == kod {
			return &item, nil
		}
	}

	// Döviz verilerinde ara
	for _, item := range fiyatlar.CurrencyItems {
		if strings.ToUpper(item.Kod) == kod {
			return &item, nil
		}
	}

	return nil, fmt.Errorf("ürün kodu bulunamadı: %s", kod)
}

// parseFloat string'i float64'e çevirir
func parseFloat(s string) float64 {
	// API'den gelen veriler zaten İngilizce formatında (4342.4900)
	// Sadece direkt parse et
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// Eğer Türkçe format varsa dönüştür
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
		f, err = strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
	}
	return f
}
