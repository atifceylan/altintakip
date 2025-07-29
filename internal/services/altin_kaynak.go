package services

import (
	"encoding/xml"
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

// AltinFiyatlari API'den gelen fiyat bilgilerini tutar
type AltinFiyatlari struct {
	CeyrekAltinAlis      float64
	CeyrekAltinSatis     float64
	YarimAltinAlis       float64
	YarimAltinSatis      float64
	TamAltinAlis         float64
	TamAltinSatis        float64
	KulceAltinAlis       float64
	KulceAltinSatis      float64
	Gram22AyarAlis       float64
	Gram22AyarSatis      float64
	Gram22HurdaAyarAlis  float64
	Gram22HurdaAyarSatis float64
	Gram24AyarAlis       float64
	Gram24AyarSatis      float64
	USDKurus             float64
	EURKurus             float64
	GBPKurus             float64
	GuncellemeTarihi     time.Time
}

// XML Response yapıları
type envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    body     `xml:"Body"`
}

type body struct {
	XMLName          xml.Name         `xml:"Body"`
	GoldResponse     goldResponse     `xml:"GetGoldResponse"`
	CurrencyResponse currencyResponse `xml:"GetCurrencyResponse"`
}

type goldResponse struct {
	XMLName xml.Name `xml:"GetGoldResponse"`
	Result  string   `xml:"GetGoldResult"`
}

type currencyResponse struct {
	XMLName xml.Name `xml:"GetCurrencyResponse"`
	Result  string   `xml:"GetCurrencyResult"`
}

// Kurlar API'den gelen kurlar XML yapısı
type Kurlar struct {
	XMLName xml.Name `xml:"Kurlar"`
	Kurlar  []Kur    `xml:"Kur"`
}

type Kur struct {
	Kod               string `xml:"Kod"`
	Aciklama          string `xml:"Aciklama"`
	Alis              string `xml:"Alis"`
	Satis             string `xml:"Satis"`
	GuncellenmeZamani string `xml:"GuncellenmeZamani"`
}

// NewAltinKaynakService yeni servis instance'ı oluşturur
func NewAltinKaynakService() *AltinKaynakService {
	return &AltinKaynakService{
		baseURL: "http://data.altinkaynak.com/DataService.asmx",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetFiyatlar altın ve döviz fiyatlarını çeker
func (s *AltinKaynakService) GetFiyatlar() (*AltinFiyatlari, error) {
	fiyatlar := &AltinFiyatlari{
		GuncellemeTarihi: time.Now(),
	}

	// Altın fiyatlarını çek
	altinData, err := s.callSOAPService("GetGold")
	if err != nil {
		return nil, fmt.Errorf("altın fiyatları çekme hatası: %w", err)
	}

	// Döviz fiyatlarını çek
	dovizData, err := s.callSOAPService("GetCurrency")
	if err != nil {
		return nil, fmt.Errorf("döviz fiyatları çekme hatası: %w", err)
	}

	// Altın fiyatlarını parse et
	err = s.parseAltinData(altinData, fiyatlar)
	if err != nil {
		return nil, fmt.Errorf("altın verisi parse hatası: %w", err)
	}

	// Döviz fiyatlarını parse et
	err = s.parseDovizData(dovizData, fiyatlar)
	if err != nil {
		return nil, fmt.Errorf("döviz verisi parse hatası: %w", err)
	}

	return fiyatlar, nil
}

// callSOAPService belirtilen action ile SOAP servisi çağırır
func (s *AltinKaynakService) callSOAPService(action string) (string, error) {
	// SOAP request body
	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap12:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
  <soap12:Header>
    <AuthHeader xmlns="http://data.altinkaynak.com/">
      <Username>AltinkaynakWebServis</Username>
      <Password>AltinkaynakWebServis</Password>
    </AuthHeader>
  </soap12:Header>
  <soap12:Body>
    <%s xmlns="http://data.altinkaynak.com/" />
  </soap12:Body>
</soap12:Envelope>`, action)

	req, err := http.NewRequest("POST", s.baseURL, strings.NewReader(soapBody))
	if err != nil {
		return "", fmt.Errorf("HTTP request oluşturma hatası: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", fmt.Sprintf("http://data.altinkaynak.com/%s", action))

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API çağrısı başarısız: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API hatası: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("response okuma hatası: %w", err)
	}

	var env envelope
	err = xml.Unmarshal(body, &env)
	if err != nil {
		return "", fmt.Errorf("XML parse hatası: %w", err)
	}

	// Action'a göre doğru response'u döner
	switch action {
	case "GetGold":
		return env.Body.GoldResponse.Result, nil
	case "GetCurrency":
		return env.Body.CurrencyResponse.Result, nil
	default:
		return "", fmt.Errorf("bilinmeyen action: %s", action)
	}
}

// parseAltinData altın fiyat verisini parse eder
func (s *AltinKaynakService) parseAltinData(dataStr string, fiyatlar *AltinFiyatlari) error {
	// HTML encoded XML string'ini decode et
	dataStr = strings.ReplaceAll(dataStr, "&lt;", "<")
	dataStr = strings.ReplaceAll(dataStr, "&gt;", ">")
	dataStr = strings.ReplaceAll(dataStr, "&quot;", "\"")
	dataStr = strings.ReplaceAll(dataStr, "&amp;", "&")

	// XML'i parse et
	var kurlar Kurlar
	err := xml.Unmarshal([]byte(dataStr), &kurlar)
	if err != nil {
		return fmt.Errorf("XML unmarshal hatası: %w", err)
	}

	// Altın kurlarını map et
	for _, kur := range kurlar.Kurlar {
		alis := parseFloat(kur.Alis)
		satis := parseFloat(kur.Satis)

		switch kur.Kod {
		case "B": // 22 Ayar Bilezik (Gram)
			fiyatlar.Gram22AyarAlis = alis
			fiyatlar.Gram22AyarSatis = satis
		case "B_T": // 22 Ayar Hurda
			fiyatlar.Gram22HurdaAyarAlis = alis
			fiyatlar.Gram22HurdaAyarSatis = satis
		case "C": // Çeyrek Altın
			fiyatlar.CeyrekAltinAlis = alis
			fiyatlar.CeyrekAltinSatis = satis
		case "CH_T": // Külçe Toptan
			fiyatlar.KulceAltinAlis = alis
			fiyatlar.KulceAltinSatis = satis
		case "Y": // Yarım Altın
			fiyatlar.YarimAltinAlis = alis
			fiyatlar.YarimAltinSatis = satis
		case "T": // Tam Altın (Teklik)
			fiyatlar.TamAltinAlis = alis
			fiyatlar.TamAltinSatis = satis
		case "GA": // Gram Altın (22 ayar)
			fiyatlar.Gram22AyarAlis = alis
			fiyatlar.Gram22AyarSatis = satis
		case "GAT": // Gram Toptan (Külçe)
			fiyatlar.KulceAltinAlis = alis
			fiyatlar.KulceAltinSatis = satis
		case "HH_T": // Has Toptan (24 ayar)
			fiyatlar.Gram24AyarAlis = alis
			fiyatlar.Gram24AyarSatis = satis
		}
	}

	return nil
}

// parseDovizData döviz kuru verisini parse eder
func (s *AltinKaynakService) parseDovizData(dataStr string, fiyatlar *AltinFiyatlari) error {
	// HTML encoded XML string'ini decode et
	dataStr = strings.ReplaceAll(dataStr, "&lt;", "<")
	dataStr = strings.ReplaceAll(dataStr, "&gt;", ">")
	dataStr = strings.ReplaceAll(dataStr, "&quot;", "\"")
	dataStr = strings.ReplaceAll(dataStr, "&amp;", "&")

	// XML'i parse et
	var kurlar Kurlar
	err := xml.Unmarshal([]byte(dataStr), &kurlar)
	if err != nil {
		return fmt.Errorf("XML unmarshal hatası: %w", err)
	}

	// Döviz kurlarını map et
	for _, kur := range kurlar.Kurlar {
		alis := parseFloat(kur.Alis)

		switch kur.Kod {
		case "USD":
			fiyatlar.USDKurus = alis
		case "EUR":
			fiyatlar.EURKurus = alis
		case "GBP":
			fiyatlar.GBPKurus = alis
		}
	}

	return nil
} // GetFiyatByType belirli bir ürün kodu için fiyat döner
func (s *AltinKaynakService) GetFiyatByType(fiyatlar *AltinFiyatlari, kod string) (float64, error) {
	kod = strings.ToUpper(strings.TrimSpace(kod))

	switch kod {
	// Altın kodları
	case "B": // 22 Ayar Bilezik
		return fiyatlar.Gram22AyarAlis, nil
	case "B_T": // 22 Ayar Hurda
		return fiyatlar.Gram22HurdaAyarAlis, nil
	case "C": // Çeyrek Altın
		return fiyatlar.CeyrekAltinAlis, nil
	case "CH_T": // Külçe Toptan
		return fiyatlar.KulceAltinAlis, nil
	case "Y": // Yarım Altın
		return fiyatlar.YarimAltinAlis, nil
	case "T": // Tam Altın (Teklik)
		return fiyatlar.TamAltinAlis, nil
	case "GA": // Gram Altın
		return fiyatlar.Gram22AyarAlis, nil
	case "GAT": // Gram Toptan (Külçe)
		return fiyatlar.KulceAltinAlis, nil
	case "HH_T": // Has Toptan (24 ayar)
		return fiyatlar.Gram24AyarAlis, nil
	// Döviz kodları
	case "USD": // Amerikan Doları
		return fiyatlar.USDKurus, nil
	case "EUR": // Avrupa Para Birimi
		return fiyatlar.EURKurus, nil
	case "GBP": // İngiliz Sterlini
		return fiyatlar.GBPKurus, nil
	default:
		return 0, fmt.Errorf("bilinmeyen ürün kodu: %s", kod)
	}
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
