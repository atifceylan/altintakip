# 🏦 Altın Takip - Kişisel Altın ve Döviz Envanter Takip Sistemi

Bu uygulama, kişisel altın ve döviz envanterinizi takip etmenizi sağlayan bir TUI (Terminal User Interface) uygulamasıdır. SQLite veritabanı ve GORM ORM kullanarak verilerinizi güvenle saklar ve `http://data.altinkaynak.com/DataService.asmx` servisinden güncel fiyatları çekerek envanterinizin güncel değerini hesaplar.

## ✨ Özellikler

- 🖥️ **İnteraktif TUI**: Ncurses benzeri terminal arayüzü (tview framework)
- 📊 **Envanter Takibi**: Altın ve döviz envanterinizi detaylı şekilde kaydedin
- 💰 **Güncel Fiyatlar**: API'den otomatik güncel fiyat çekme
- 📈 **Kar/Zarar Hesaplama**: Alış fiyatı ile güncel fiyat karşılaştırması
- ⌨️ **Klavye Kısayolları**: F1, F5, Ctrl+L, Ctrl+Q ile hızlı navigasyon
- 🎨 **Renkli Tablo**: Kar/zarar durumuna göre renk kodları
- 📊 **Canlı Özet Panel**: Anlık toplam değerler ve istatistikler
- 🔄 **Otomatik Güncelleme**: Periyodik fiyat güncellemesi
- 🗄️ **SQLite**: Hafif ve taşınabilir veri saklama
- 💾 **Yerel Veri Modu**: API çağrısı yapmadan offline çalışma
- 📁 **Otomatik Klasör**: Kullanıcı dizininde otomatik altintakip klasörü oluşturma

## 🚀 Kurulum

### Gereksinimler

- Go 1.21 veya üstü

### 1. Proje Kurulumu

```bash
# Projeyi indirin
git clone <repository-url>
cd altintakip

# Bağımlılıkları yükleyin
go mod download
```

### 2. Konfigürasyon (İsteğe Bağlı)

Uygulama ilk çalıştırıldığında kullanıcı dizininizde `~/altintakip` klasörünü otomatik olarak oluşturur ve tüm verileri burada saklar.

Varsayılan ayarları değiştirmek için `.altintakip_env` dosyası oluşturabilirsiniz:

```bash
# Örnek konfigürasyon dosyasını kopyalayın
cp .altintakip_env.example .altintakip_env
```

`.altintakip_env` dosyası içeriği:

```env
# Veritabanı dosyası yolu (varsayılan: ~/altintakip/altintakip.db)
DB_PATH=

# Log dosyası yolu (varsayılan: ~/altintakip/altintakip.log)  
LOG_PATH=

# Uygulama veri dizini (varsayılan: ~/altintakip)
APP_DATA_DIR=

# API endpoint (opsiyonel)
API_ENDPOINT=

# Log seviyesi (DEBUG, INFO, WARN, ERROR)
LOG_LEVEL=
```

**Not:** SQLite kullandığımız için harici veritabanı kurulumuna gerek yoktur. Veritabanı dosyası otomatik olarak oluşturulur.

## 🎯 Kullanım

### TUI Uygulamasını Başlatma

```bash
# Normal mod (API'den güncel fiyatları çeker)
go run main.go

# Liste modu (offline - sadece veritabanındaki veriler gösterilir)
go run main.go list
```

### Çalışma Modları

#### **Normal Mod**
- API'den güncel fiyatları otomatik çeker
- 5 dakikada bir otomatik fiyat güncelleme yapar
- F5 ile manuel fiyat güncelleme yapılabilir
- Add/Edit işlemlerinde güncel fiyat girilmezse API'den çekilir

#### **Liste Modu (Offline)**
- Sadece veritabanındaki mevcut veriler gösterilir
- API çağrısı yapılmaz (internet bağlantısı gerekmez)
- Otomatik fiyat güncelleme devre dışıdır
- F5 tuşu devre dışıdır
- Add/Edit işlemlerinde güncel fiyat mutlaka girilmelidir

### Klavye Kısayolları

- **F1**: Yardım penceresini açar/kapatır
- **F5**: Verileri API'den yeniler (sadece normal modda)
- **Ctrl+L**: Yerel veri modunu açar/kapatır (offline)
- **Ctrl+Q**: Uygulamadan çıkar
- **ESC**: Sadece modal pencerelerini kapatır (uygulamayı sonlandırmaz)
- **Tab**: Tablolar arasında geçiş yapar

### TUI Arayüzü

Uygulama şu bölümlerden oluşur:

1. **Ana Tablo**: Envanter verilerini gösterir
2. **Grup Tablosu**: Kod bazlı gruplandırılmış veriler
3. **Durum Çubuğu**: Güncel durum bilgileri
4. **Özet Paneli**: Toplam değerler ve istatistikler
5. **Yardım Modal**: Klavye kısayolları rehberi

### İlk Çalıştırma

Uygulama ilk çalıştırıldığında:
1. `~/altintakip` dizinini otomatik oluşturur
2. Veritabanı dosyasını bu dizinde oluşturur
3. Log dosyasını bu dizinde oluşturur
4. Veritabanı tablolarını otomatik oluşturur
5. TUI arayüzünü başlatır
6. API'den güncel fiyatları çeker
7. İnteraktif tablo olarak sonuçları gösterir

## 📁 Veri Dizini Yapısı

```
~/altintakip/              # Ana uygulama dizini
├── altintakip.db         # SQLite veritabanı
├── altintakip.log        # Log dosyası
└── .altintakip_env       # Konfigürasyon dosyası (opsiyonel)
```

## 📁 Proje Yapısı

```
altintakip/
├── main.go              # Ana giriş noktası
├── cmd/                 # Komut katmanı
│   └── cmd.go          # Uygulama mantığı
├── internal/            # İç paketler
│   ├── models/         # Veri modelleri
│   │   └── envanter.go
│   ├── database/       # Veritabanı işlemleri
│   │   └── database.go
│   ├── services/       # İş mantığı
│   │   ├── altin_kaynak.go
│   │   └── envanter_service.go
│   └── tui/            # TUI arayüzü
│       └── app.go
├── .altintakip_env.example  # Örnek konfigürasyon
├── go.mod              # Go modül dosyası
└── README.md          # Bu dosya
```

## 🔧 API Entegrasyonu

### Desteklenen Varlık Kodları

Uygulama aşağıdaki varlık kodlarını destekler:

**Altın:**
- `B`: 22 Ayar Bilezik
- `B_T`: 22 Ayar Hurda
- `C`: Çeyrek Altın  
- `CH_T`: Külçe Toptan
- `Y`: Yarım Altın
- `T`: Tam Altın (Teklik)
- `GA`: 22 Ayar Gram
- `GAT`: Gram Toptan (Külçe)
- `HH_T`: Has Toptan (24 ayar)

**Döviz:**
- `USD`: Amerikan Doları
- `EUR`: Avrupa Para Birimi
- `GBP`: İngiliz Sterlini

### SOAP API Entegrasyonu

Uygulama `data.altinkaynak.com` servisinden SOAP protokolü ile veri çeker. Her varlık kodu için dinamik olarak API çağrısı yapar ve XML response'unu parse eder.

## 🎨 Özellikler

### Finansal Sayı Formatı

- **Binli Gruplar**: Türk standardında nokta ile ayrım (1.234.567)
- **Ondalık**: Türk standardında virgül ile ayrım (1.234,56)
- **Koşullu Ondalık**: Tam sayılarda .00 gösterilmez (1.234 ₺)
- **Para Birimi**: Başlıklarda parantez içinde (₺)

### Tablo Özellikleri

- **Ana Envanter**: Tüm envanter kalemlerinin detaylı listesi
- **Grup Analizi**: Kod bazlı gruplandırılmış analiz
- **Özet Bilgiler**: Toplam değerler ve genel performans

### Komut Parametreleri

- **Normal Mod**: `go run main.go` - API'den güncel fiyatları çeker
- **Liste Modu**: `go run main.go list` - Sadece veritabanındaki verileri gösterir

## 🔧 Geliştirme

### Yeni Envanter Ekleme

Kodu düzenleyerek `addSampleData` fonksiyonuna yeni ürünler ekleyebilirsiniz:

```go
{
    Tur:        "Altın",
    Cins:       "Yarım Altın",
    Kod:        "Y",
    Miktar:     3.000,
    Birim:      "adet",
    AlisTarihi: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
    AlisFiyati: 5600.00,
    Notlar:     "Yarım altın yatırımı",
}
```

### Veritabanı Şeması

Envanter tablosu şu alanları içerir:
- **id**: Benzersiz kimlik
- **tur**: "Altın" veya "Döviz"
- **cins**: Ürün cinsi (22 Ayar Külçe, USD, EUR vb.)
- **kod**: API kod karşılığı (GA, C, Y, T, vb.)
- **miktar**: Miktar (gram/adet)
- **birim**: Birim (gram/adet)
- **alis_tarihi**: Alış tarihi
- **alis_fiyati**: Birim başına alış fiyatı
- **toplam_alis**: Toplam alış tutarı
- **guncel_fiyat**: Güncel birim fiyatı
- **guncel_tutar**: Güncel toplam tutar
- **kar_zarar**: Kar/zarar miktarı
- **kar_zarar_yuzde**: Kar/zarar yüzdesi

## 🐛 Sorun Giderme

### Uygulama Dizini Oluşturma Hatası
```bash
HATA: Uygulama veri dizini oluşturulamadı
```
- Kullanıcı dizininizde yazma izniniz olduğundan emin olun
- `~/altintakip` klasörünü manuel olarak oluşturmayı deneyin

### Veritabanı Dosyası Hatası
```bash
HATA: SQLite veritabanı bağlantısı kurulamadı
```
- `~/altintakip` dizininde yazma izniniz olduğundan emin olun
- `.altintakip_env` dosyasında belirtilen dizinin mevcut olduğundan emin olun

### API Hatası
```bash
UYARI: fiyatlar alınamadı
```
- İnternet bağlantınızı kontrol edin
- API servisi geçici olarak kapalı olabilir
- Ctrl+L ile yerel veri modunu kullanarak mevcut verilerle çalışabilirsiniz

### Çıkış Sorunu
- Uygulama artık ESC tuşu ile kapanmaz
- **Sadece Ctrl+Q** ile çıkış yapabilirsiniz
- ESC tuşu sadece açık modal pencerelerini kapatır

## 🤝 Katkıda Bulunma

1. Projeyi fork edin
2. Feature branch oluşturun (`git checkout -b feature/amazing-feature`)
3. Değişikliklerinizi commit edin (`git commit -m 'Add amazing feature'`)
4. Branch'inizi push edin (`git push origin feature/amazing-feature`)
5. Pull Request oluşturun

## 📝 Lisans

Bu proje MIT lisansı altında lisanslanmıştır.

## 📞 İletişim

Sorularınız için issue açabilirsiniz.

---

**Not**: Bu uygulama yatırım tavsiyesi vermez. Tüm yatırım kararları kendi sorumluluğunuzdadır.
