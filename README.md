# 🏦 Altın Takip - Kişisel Altın ve Döviz Envanter Takip Sistemi

Bu uygulama, kişisel altın ve döviz envanterinizi takip etmenizi sağlayan bir CLI (komut satırı) uygulamasıdır. PostgreSQL veritabanı ve GORM ORM kullanarak verilerinizi güvenle saklar ve `http://data.altinkaynak.com/DataService.asmx` servisinden güncel fiyatları çekerek envanterinizin güncel değerini hesaplar.

## ✨ Özellikler

- 📊 **Envanter Takibi**: Altın ve döviz envanterinizi detaylı şekilde kaydedin
- 💰 **Güncel Fiyatlar**: API'den otomatik güncel fiyat çekme
- 📈 **Kar/Zarar Hesaplama**: Alış fiyatı ile güncel fiyat karşılaştırması
- 🎨 **CLI Tablosu**: Temiz ve düzenli tablo görünümü
- � **Binli Grup Formatı**: Türk finansal standartlarında sayı gösterimi
- 📊 **Grup Analizi**: Kod bazlı grup analizi ve toplam değerler
- �🔄 **Otomatik Güncelleme**: Her çalıştırmada fiyatları günceller
- � **Liste Modu**: API çağrısı yapmadan mevcut verileri görüntüleme
- �🗄️ **PostgreSQL**: Güvenli veri saklama

## 🚀 Kurulum

### Gereksinimler

- Go 1.21 veya üstü
- PostgreSQL veritabanı

### 1. PostgreSQL Kurulumu

macOS için:
```bash
brew install postgresql
brew services start postgresql
```

Debian/Ubuntu için:
```bash
apt install postgresql-XX #debian 12 icin postgresql-15

# Veritabanı oluşturma
sudo su - postgres
psql
create user altintakip_app password '<your_password_here>';
create database altintakip owner altintakip_app;
```

### 2. Proje Kurulumu

```bash
# Projeyi indirin
git clone <repository-url>
cd altintakip

# Bağımlılıkları yükleyin
go mod download

# Konfigürasyon dosyasını oluşturun
touch .altintakip_env
# .altintakip_env dosyasını düzenleyerek veritabanı bilgilerinizi girin
```

### 3. Konfigürasyon Dosyası (.altintakip_env)

Uygulama konfigürasyon dosyasını önce mevcut dizinde, sonra home dizininde arar. `.altintakip_env` dosyasını oluşturun ve veritabanı bilgilerinizi güncelleyin:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=altintakip_app
DB_PASSWORD=your_password_here
DB_NAME=altintakip
DB_SSLMODE=disable
```

## 🎯 Kullanım

### Normal Mod (API Çağrısı ile)

Uygulamayı çalıştırmak için:

```bash
go run main.go
```

### Liste Modu (Sadece Mevcut Veriler)

API çağrısı yapmadan mevcut verileri görüntülemek için:

```bash
go run main.go list
```

### İlk Çalıştırma

Uygulama ilk çalıştırıldığında:
1. Veritabanı tablolarını otomatik oluşturur
2. Örnek envanter verileri ekler (boş ise)
3. API'den güncel fiyatları çeker
4. Güzel formatlı tablo olarak sonuçları gösterir

### Örnek Çıktı

```
💰 KİŞİSEL ALTIN VE DÖVİZ ENVANTERİ
═══════════════════════════════════════════════════════════════════════════════════════════════════

TÜR      CİNS                 MİKTAR     BİRİM    ALIŞ TARİHİ  ALIŞ FİYATI (₺) TOPLAM ALIŞ (₺) GÜNCEL FİYAT (₺) GÜNCEL TUTAR (₺) KAR/ZARAR (₺)  KAR/ZARAR %
─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
Altın    22 Ayar Külçe       10.00      gram     15.01.2024   2.250           22.500          2.485,26         24.853           +2.353         +10.46%    
Altın    Çeyrek Altın        5.00       adet     20.02.2024   2.800           14.000          3.151,72         15.759           +1.759         +12.56%    

📋 GRUP ANALİZİ
═══════════════════════════════════════════════════════════════════════════════════════════════════

TÜR      CİNS            TOPLAM MİKTAR BİRİM    ORT. ALIŞ FİYATI (₺) TOPLAM ALIŞ (₺) TOPLAM GÜNCEL (₺) TOPLAM KAR/ZARAR (₺) KAR/ZARAR %
────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
Altın    22 Ayar Külçe  10.00         gram     2.250               22.500          24.853            +2.353             +10.46%   
Altın    Çeyrek Altın   5.00          adet     2.800               14.000          15.759            +1.759             +12.56%   

📊 ÖZET BİLGİLER
═══════════════════════════════════════════════════════════════════════════════════════════════════

TOPLAM ALIŞ TUTARI (₺)    TOPLAM GÜNCEL TUTAR (₺)  TOPLAM KAR/ZARAR (₺)    TOPLAM KAR/ZARAR %      
─────────────────────────────────────────────────────────────────────────────────────────
36.500                   40.612                   +4.112                  +11.27%                

🕐 Son güncelleme: 15:04:05 - 27.01.2025
📊 Veriler data.altinkaynak.com'dan alınmaktadır
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
│   └── cli/            # CLI görüntüleme
│       └── table.go
├── .altintakip_env     # Konfigürasyon dosyası
├── .github/
│   └── copilot-instructions.md
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

### Konfigürasyon Dosyası Hatası
```bash
HATA: .altintakip_env dosyası bulunamadı
```
- Mevcut dizinde veya home dizininde `.altintakip_env` dosyası oluşturun
- Dosyada veritabanı bağlantı bilgilerini kontrol edin

### Veritabanı Bağlantı Hatası
```bash
ERRO[0000] veritabanı bağlantısı kurulamadı
```
- PostgreSQL'in çalıştığından emin olun: `brew services start postgresql`
- `.altintakip_env` dosyasındaki veritabanı bilgilerini kontrol edin
- Veritabanının mevcut olduğundan emin olun: `createdb altintakip`

### API Hatası
```bash
UYARI: fiyatlar alınamadı
```
- İnternet bağlantınızı kontrol edin
- API servisi geçici olarak kapalı olabilir
- Liste modunu kullanarak mevcut verilerle çalışabilirsiniz: `go run main.go list`

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
