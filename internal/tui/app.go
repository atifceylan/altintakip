package tui

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"altintakip/internal/models"
	"altintakip/internal/services"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/shopspring/decimal"
)

// CinsItem her cins için hem görünen isim hem de API kodu tutar
type CinsItem struct {
	Name string // Görünen isim
	Code string // API kodu
}

// Sabit değerler
var (
	turOptions = []string{"Altın", "Döviz"}

	birimOptions = []string{"gram", "adet", "kilogram", "ons"}

	// Cins seçenekleri ve kod eşleştirmeleri birleşik yapı
	cinsMapping = map[string][]CinsItem{
		"Altın": {
			{"24 Ayar Kesme", "CH_T"},
			{"24 Ayar Gram", "GA"},
			{"Külçe Altın", "GAT"},
			{"22 Ayar Bilezik", "B"},
			{"22 Ayar Gram", "B_T"},
			{"Çeyrek Altın", "C"},
			{"Yarım Altın", "Y"},
			{"Tam Altın", "T"},
		},
		"Döviz": {
			{"USD", "USD"},
			{"EUR", "EUR"},
			{"GBP", "GBP"},
		},
	}
)

// App TUI uygulaması yapısı
type App struct {
	app       *tview.Application
	table     *tview.Table
	grupTable *tview.Table
	ozetTable *tview.Table
	pages     *tview.Pages
	mainFlex  *tview.Flex

	// Scroll indicator'lar
	envanterScrollIndicator *tview.TextView
	grupScrollIndicator     *tview.TextView

	// Fiyat güncelleme servisi
	envanterService *services.EnvanterService
	updateTicker    *time.Ticker
	stopChan        chan bool

	// Liste modu (offline mod)
	isListMode bool
}

// NewApp yeni TUI uygulaması oluşturur
func NewApp() *App {
	return &App{
		app:             tview.NewApplication(),
		pages:           tview.NewPages(),
		envanterService: services.NewEnvanterService(),
		stopChan:        make(chan bool),
		isListMode:      false,
	}
}

// SetListMode liste modunu ayarlar
func (a *App) SetListMode(listMode bool) {
	a.isListMode = listMode
}

// Run TUI uygulamasını başlatır
func (a *App) Run() error {
	// Liste modu değilse güncel fiyatları çek
	if !a.isListMode {
		log.Printf("Uygulama başlatılıyor, güncel fiyatlar getiriliyor...")
		err := a.envanterService.UpdateGuncelFiyatlar()
		if err != nil {
			log.Printf("UYARI: Güncel fiyatlar alınamadı: %v", err)
		} else {
			log.Printf("Güncel fiyatlar başarıyla güncellendi")
		}

		// Otomatik güncelleme başlat (5 dakika) - sadece liste modu değilse
		a.startAutoUpdate()
	} else {
		log.Printf("Liste modu: Sadece veritabanındaki veriler gösterilecek")
	}

	// Terminal ortamını kontrol et
	//log.Printf("TERM: %s", os.Getenv("TERM"))
	//log.Printf("COLORTERM: %s", os.Getenv("COLORTERM"))

	// Basit tablo oluştur
	a.table = tview.NewTable()
	a.table.SetBorders(true)
	a.table.SetSelectable(true, false)
	a.table.SetFixed(1, 0)                                                                                 // İlk satırı (header) sabit tut
	a.table.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)) // Seçili satır stilini ayarla
	a.table.SetTitle(" 📊 ENVANTER ")
	a.table.SetBorderColor(tcell.ColorBlue)
	// Envanter tablosu focus aldığında grup tablosunun seçimini kaldır
	a.table.SetFocusFunc(func() {
		a.grupTable.SetSelectable(false, false)
		a.table.SetSelectable(true, false)
		a.updateScrollIndicators()
	})

	// Envanter tablosunda seçim değiştiğinde scroll indicator'ı güncelle
	a.table.SetSelectionChangedFunc(func(row, column int) {
		a.updateScrollIndicators()
	})

	// Envanter scroll indicator'ı oluştur
	a.envanterScrollIndicator = tview.NewTextView()
	a.envanterScrollIndicator.SetTextAlign(tview.AlignCenter)
	a.envanterScrollIndicator.SetTextColor(tcell.ColorBlue)
	a.envanterScrollIndicator.SetText("▲\n│\n│\n▼")

	// Grup tablosu oluştur
	a.grupTable = tview.NewTable()
	a.grupTable.SetBorders(true)
	a.grupTable.SetSelectable(false, false) // Başlangıçta seçilemez
	a.grupTable.SetFixed(1, 0)              // İlk satırı (header) sabit tut
	a.grupTable.SetTitle(" 📋 GRUP ANALİZİ ")
	a.grupTable.SetBorderColor(tcell.ColorGreen)
	// Grup tablosu focus aldığında kendisini seçilebilir yap ve envanter tablosunu seçilemez yap
	a.grupTable.SetFocusFunc(func() {
		a.table.SetSelectable(false, false)
		a.grupTable.SetSelectable(true, false)
		a.grupTable.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack))
		a.updateScrollIndicators()
	})

	// Grup tablosunda seçim değiştiğinde scroll indicator'ı güncelle
	a.grupTable.SetSelectionChangedFunc(func(row, column int) {
		a.updateScrollIndicators()
	})

	// Grup scroll indicator'ı oluştur
	a.grupScrollIndicator = tview.NewTextView()
	a.grupScrollIndicator.SetTextAlign(tview.AlignCenter)
	a.grupScrollIndicator.SetTextColor(tcell.ColorGreen)
	a.grupScrollIndicator.SetText("▲\n│\n▼")

	// Özet tablosu oluştur
	a.ozetTable = tview.NewTable()
	a.ozetTable.SetBorders(true)
	a.ozetTable.SetSelectable(false, false)
	a.ozetTable.SetTitle(" 💰 ÖZET BİLGİLER ")
	a.ozetTable.SetBorderColor(tcell.ColorYellow)

	// Başlıklar
	headers := []string{
		"TÜR", "CİNS", "MİKTAR", "ALIŞ TARİHİ", "ALIŞ FİYATI ₺", "TOPLAM ALIŞ ₺", "GÜNCEL FİYAT ₺", "GÜNCEL TUTAR ₺", "KAR/ZARAR ₺",
	}

	for col, header := range headers {
		a.table.SetCell(0, col, tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false))
	}

	// Verileri yükle
	log.Printf("Veri yükleme işlemi başlatılıyor...")
	a.loadData()
	a.loadGrupData() // Grup analizini yükle
	a.loadOzetData() // Özet verilerini yükle
	log.Printf("Veri yükleme işlemi tamamlandı, TUI başlatılıyor...")

	// İlk başta envanter tablosuna focus ayarla
	a.app.SetFocus(a.table)

	// Klavye kısayolları
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Eğer modal açıksa ve Escape tuşuna basılmışsa, sadece modal'ı kapat
		if a.pages.HasPage("add-form") || a.pages.HasPage("edit-form") || a.pages.HasPage("delete-confirm") || a.pages.HasPage("message") {
			if event.Key() == tcell.KeyEscape {
				// Hangi modal açıksa onu kapat
				if a.pages.HasPage("add-form") {
					a.pages.RemovePage("add-form")
					a.clearAllTables()
					a.app.ForceDraw()
					a.loadData()
					a.loadGrupData()
					a.loadOzetData()
					a.app.SetFocus(a.table)
				} else if a.pages.HasPage("edit-form") {
					a.pages.RemovePage("edit-form")
					a.clearAllTables()
					a.app.ForceDraw()
					a.loadData()
					a.loadGrupData()
					a.loadOzetData()
					a.app.SetFocus(a.table)
				} else if a.pages.HasPage("delete-confirm") {
					a.pages.RemovePage("delete-confirm")
					a.app.ForceDraw()
					a.app.SetFocus(a.table)
				} else if a.pages.HasPage("message") {
					a.pages.RemovePage("message")
					a.app.ForceDraw()
					a.app.SetFocus(a.table)
				}
				return nil
			}
			// Modal açıkken diğer kısayolları blokla
			if event.Key() == tcell.KeyCtrlQ {
				a.app.Stop()
				return nil
			}
			return event
		}

		// Ana ekrandayken normal kısayollar
		switch event.Key() {
		case tcell.KeyCtrlQ:
			a.app.Stop()
			return nil
		case tcell.KeyTab:
			// Tab ile sadece envanter ve grup tabloları arasında geçiş
			current := a.app.GetFocus()
			switch current {
			case a.table:
				a.app.SetFocus(a.grupTable)
			case a.grupTable:
				a.app.SetFocus(a.table)
			default:
				a.app.SetFocus(a.table)
			}
			return nil
		case tcell.KeyF5:
			// Manuel fiyat güncelleme - sadece liste modu değilse
			if a.isListMode {
				a.showMessage("Liste modunda fiyat güncelleme yapılmaz!")
				return nil
			}
			go func() {
				log.Printf("Manuel fiyat güncelleme başlatılıyor...")
				err := a.envanterService.UpdateGuncelFiyatlar()
				if err != nil {
					log.Printf("Manuel fiyat güncelleme başarısız: %v", err)
					a.app.QueueUpdateDraw(func() {
						a.showMessage(fmt.Sprintf("Fiyat güncelleme başarısız: %v", err))
					})
				} else {
					log.Printf("Manuel fiyat güncelleme başarılı")
					a.app.QueueUpdateDraw(func() {
						a.loadData()
						a.loadGrupData()
						a.loadOzetData()
						a.showMessage("Fiyatlar başarıyla güncellendi!")
					})
				}
			}()
			return nil
		}
		switch event.Rune() {
		case 'e', 'E': // Ekleme
			a.showAddForm()
			return nil
		case 'd', 'D': // Düzenleme
			a.showEditForm()
			return nil
		case 's', 'S': // Silme
			a.showDeleteConfirm()
			return nil
		}
		return event
	})

	// Layout oluştur - 3 tablo dikey olarak + alt boşluk
	headerText := "🏦 ALTIN TAKİP - TUI (F5: Yenile, Tab: Tablolar Arası Geçiş, E: Ekle, D: Düzenle, S: Sil, Ctrl+Q: Çıkış)"
	if a.isListMode {
		headerText = "🏦 ALTIN TAKİP - LİSTE MODU (OFFLINE - Tab: Tablolar Arası Geçiş, E: Ekle, D: Düzenle, S: Sil, Ctrl+Q: Çıkış)"
	}

	a.mainFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().
			SetText(headerText).
			SetTextAlign(tview.AlignCenter).
			SetTextColor(tcell.ColorYellow), 1, 0, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(a.table, 0, 1, true).
				AddItem(a.envanterScrollIndicator, 1, 0, false), 0, 3, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(a.grupTable, 0, 1, false).
				AddItem(a.grupScrollIndicator, 1, 0, false), 0, 2, false).
		AddItem(a.ozetTable, 5, 0, false).
		AddItem(tview.NewTextView(), 1, 0, false) // Alt boşluk

	// Pages ile modal yönetimi
	a.pages.AddPage("main", a.mainFlex, true, true)

	log.Printf("TUI başlatılıyor...")
	a.app.SetRoot(a.pages, true)

	// Root ayarlandıktan sonra focus ve seçimi ayarla
	a.app.SetFocus(a.table)
	// İlk veri satırını seç (header değil)
	if a.table.GetRowCount() > 1 {
		a.table.Select(1, 0)
	}

	// Focus'u garanti altına almak için goroutine ile gecikmeli focus
	go func() {
		time.Sleep(100 * time.Millisecond)
		a.app.QueueUpdateDraw(func() {
			a.app.SetFocus(a.table)
			if a.table.GetRowCount() > 1 {
				a.table.Select(1, 0)
			}
			a.table.SetSelectable(true, false)
		})
	}()

	result := a.app.Run()

	// Uygulama kapatılırken otomatik güncellemeyi durdur
	a.stopAutoUpdate()

	log.Printf("TUI sonlandı, hata: %v", result)
	return result
}

// loadData verileri yükler
func (a *App) loadData() {
	// Tabloyu temizle
	a.table.Clear()

	// Başlıkları yeniden ekle
	headers := []string{
		"TÜR", "CİNS", "MİKTAR", "ALIŞ TARİHİ", "ALIŞ FİYATI ₺", "TOPLAM ALIŞ ₺", "GÜNCEL FİYAT ₺", "GÜNCEL TUTAR ₺", "KAR/ZARAR ₺",
	}

	for col, header := range headers {
		a.table.SetCell(0, col, tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false))
	}

	envanterService := services.NewEnvanterService()
	envanterler, err := envanterService.GetAllEnvanterFromDB()
	if err != nil {
		log.Printf("Veri yüklenemedi: %v", err)

		// Hata mesajı göster
		a.table.SetCell(1, 0, tview.NewTableCell(fmt.Sprintf("HATA: %v", err)).
			SetTextColor(tcell.ColorRed).
			SetAlign(tview.AlignCenter))
		return
	}

	log.Printf("Yüklenen envanter sayısı: %d", len(envanterler))

	// Veri yoksa bilgi göster
	if len(envanterler) == 0 {
		a.table.SetCell(1, 0, tview.NewTableCell("Envanter boş").
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter))
		return
	}

	// Verileri tabloya ekle
	for row, envanter := range envanterler {
		karZarar := envanter.GuncelTutar - envanter.ToplamAlis
		karZararColor := tcell.ColorGreen
		karZararPrefix := "+"
		if karZarar < 0 {
			karZararColor = tcell.ColorRed
			karZararPrefix = ""
		}

		a.table.SetCell(row+1, 0, tview.NewTableCell(envanter.Tur))
		a.table.SetCell(row+1, 1, tview.NewTableCell(envanter.Cins))
		a.table.SetCell(row+1, 2, tview.NewTableCell(fmt.Sprintf("%s %s", formatQuantity(envanter.Miktar, envanter.Birim), envanter.Birim)))
		a.table.SetCell(row+1, 3, tview.NewTableCell(envanter.AlisTarihi.Format("02.01.2006")))
		a.table.SetCell(row+1, 4, tview.NewTableCell(formatMoney(envanter.AlisFiyati)))
		a.table.SetCell(row+1, 5, tview.NewTableCell(formatMoney(envanter.ToplamAlis)))
		a.table.SetCell(row+1, 6, tview.NewTableCell(formatMoney(envanter.GuncelFiyat)))
		a.table.SetCell(row+1, 7, tview.NewTableCell(formatMoney(envanter.GuncelTutar)))
		a.table.SetCell(row+1, 8, tview.NewTableCell(fmt.Sprintf("%s%s", karZararPrefix, formatMoney(karZarar))).
			SetTextColor(karZararColor))
	}

	log.Printf("Tablo verileri hazırlandı")
	a.updateScrollIndicators() // Scroll indicator'ları güncelle

	// Envanter tablosunun seçilebilir olduğundan emin ol ve focus ayarla
	a.table.SetSelectable(true, false)
	if a.app != nil {
		a.app.SetFocus(a.table)
	}
}

// loadGrupData grup analizini yükler
func (a *App) loadGrupData() {
	envanterService := services.NewEnvanterService()
	gruplar, err := envanterService.GetKodBazliGruplar()
	if err != nil {
		log.Printf("Grup verileri yüklenemedi: %v", err)
		a.grupTable.SetCell(1, 0, tview.NewTableCell(fmt.Sprintf("HATA: %v", err)).
			SetTextColor(tcell.ColorRed).
			SetAlign(tview.AlignCenter))
		return
	}

	a.grupTable.Clear()

	// Grup tablosu başlıkları
	headers := []string{
		"TÜR", "CİNS", "TOPLAM MİKTAR", "BİRİM", "ORT. ALIŞ FİYATI ₺",
		"TOPLAM ALIŞ ₺", "TOPLAM GÜNCEL ₺", "TOPLAM KAR/ZARAR ₺", "KAR/ZARAR %",
	}

	for col, header := range headers {
		a.grupTable.SetCell(0, col, tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false))
	}

	// Grup verilerini slice'a çevir ve sırala
	type GrupVeri struct {
		Tur     string
		Cins    string
		Veriler map[string]interface{}
	}

	var grupSlice []GrupVeri
	for _, veri := range gruplar {
		grupSlice = append(grupSlice, GrupVeri{
			Tur:     veri["tur"].(string),
			Cins:    veri["cins"].(string),
			Veriler: veri,
		})
	}

	// Tur'a göre, sonra Cins'e göre sırala
	sort.Slice(grupSlice, func(i, j int) bool {
		if grupSlice[i].Tur != grupSlice[j].Tur {
			return grupSlice[i].Tur < grupSlice[j].Tur
		}
		return grupSlice[i].Cins < grupSlice[j].Cins
	})

	// Grup verilerini sırala ve göster
	row := 1
	for _, grupItem := range grupSlice {
		veri := grupItem.Veriler
		karZarar := veri["toplam_kar_zarar"].(float64)
		karZararYuzde := veri["kar_zarar_yuzde"].(float64)
		karZararColor := tcell.ColorGreen
		karZararPrefix := "+"
		if karZarar < 0 {
			karZararColor = tcell.ColorRed
			karZararPrefix = ""
		}

		a.grupTable.SetCell(row, 0, tview.NewTableCell(veri["tur"].(string)))
		a.grupTable.SetCell(row, 1, tview.NewTableCell(veri["cins"].(string)))
		a.grupTable.SetCell(row, 2, tview.NewTableCell(formatQuantity(veri["toplam_miktar"].(float64), veri["birim"].(string))))
		a.grupTable.SetCell(row, 3, tview.NewTableCell(veri["birim"].(string)))
		a.grupTable.SetCell(row, 4, tview.NewTableCell(formatMoney(veri["ortalama_alis_fiyati"].(float64))))
		a.grupTable.SetCell(row, 5, tview.NewTableCell(formatMoney(veri["toplam_alis_tutar"].(float64))))
		a.grupTable.SetCell(row, 6, tview.NewTableCell(formatMoney(veri["toplam_guncel_tutar"].(float64))))
		a.grupTable.SetCell(row, 7, tview.NewTableCell(fmt.Sprintf("%s%s", karZararPrefix, formatMoney(karZarar))).
			SetTextColor(karZararColor))
		a.grupTable.SetCell(row, 8, tview.NewTableCell(fmt.Sprintf("%s%.2f%%", karZararPrefix, karZararYuzde)).
			SetTextColor(karZararColor))
		row++
	}

	log.Printf("Grup tablosu hazırlandı")
	a.updateScrollIndicators() // Scroll indicator'ları güncelle
}

// loadOzetData özet verilerini yükler
func (a *App) loadOzetData() {
	envanterService := services.NewEnvanterService()
	envanterler, err := envanterService.GetAllEnvanterFromDB()
	if err != nil {
		log.Printf("Özet verileri yüklenemedi: %v", err)
		return
	}

	a.ozetTable.Clear()

	// Özet tablosu başlıkları
	headers := []string{
		"TOPLAM ALIŞ TUTARI ₺", "TOPLAM GÜNCEL TUTAR ₺", "TOPLAM KAR/ZARAR ₺", "TOPLAM KAR/ZARAR %",
	}

	for col, header := range headers {
		a.ozetTable.SetCell(0, col, tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false))
	}

	// Toplam hesaplamaları
	var toplamAlis, toplamGuncel, toplamKar float64
	for _, envanter := range envanterler {
		toplamAlis += envanter.ToplamAlis
		toplamGuncel += envanter.GuncelTutar
		toplamKar += (envanter.GuncelTutar - envanter.ToplamAlis)
	}

	var toplamKarYuzde float64
	if toplamAlis > 0 {
		toplamKarYuzde = (toplamKar / toplamAlis) * 100
	}

	// Renk belirleme
	karColor := tcell.ColorGreen
	karPrefix := "+"
	if toplamKar < 0 {
		karColor = tcell.ColorRed
		karPrefix = ""
	}

	// Özet satırını ekle
	a.ozetTable.SetCell(1, 0, tview.NewTableCell(formatMoney(toplamAlis)))
	a.ozetTable.SetCell(1, 1, tview.NewTableCell(formatMoney(toplamGuncel)))
	a.ozetTable.SetCell(1, 2, tview.NewTableCell(fmt.Sprintf("%s%s", karPrefix, formatMoney(toplamKar))).
		SetTextColor(karColor))
	a.ozetTable.SetCell(1, 3, tview.NewTableCell(fmt.Sprintf("%s%.2f%%", karPrefix, toplamKarYuzde)).
		SetTextColor(karColor))

	log.Printf("Özet tablosu hazırlandı")
}

// getCinsOptions belirtilen tür için cins seçeneklerini string slice olarak döner
func getCinsOptions(tur string) []string {
	items, exists := cinsMapping[tur]
	if !exists {
		return []string{}
	}

	options := make([]string, len(items))
	for i, item := range items {
		options[i] = item.Name
	}
	return options
}

// getCinsCode belirtilen cins ismi için API kodunu döner
func getCinsCode(cinsName string) string {
	for _, items := range cinsMapping {
		for _, item := range items {
			if item.Name == cinsName {
				return item.Code
			}
		}
	}
	return ""
}

// showAddForm yeni envanter ekleme formunu gösterir
func (a *App) showAddForm() {
	// Tür dropdown
	turDropdown := tview.NewDropDown().
		SetLabel("Tür").
		SetOptions(turOptions, nil)

	// Cins dropdown - başlangıçta boş
	cinsDropdown := tview.NewDropDown().
		SetLabel("Cins")

	// Birim dropdown
	birimDropdown := tview.NewDropDown().
		SetLabel("Birim").
		SetOptions(birimOptions, nil)

	// Kod alanı (gizli)
	var selectedKod string

	// Tür değiştiğinde Cins seçeneklerini güncelle
	turDropdown.SetSelectedFunc(func(text string, index int) {
		cinsOptions := getCinsOptions(text)
		cinsDropdown.SetOptions(cinsOptions, func(cinsText string, cinsIndex int) {
			// Cins seçildiğinde kodu belirle
			selectedKod = getCinsCode(cinsText)
		})
	})

	// Ana form - tüm alanlar alt alta
	form := tview.NewForm()
	form.AddFormItem(turDropdown)
	form.AddFormItem(cinsDropdown)
	form.AddInputField("Miktar", "", 20, nil, nil)
	form.AddFormItem(birimDropdown)
	form.AddInputField("Alış Tarihi", "", 20, nil, nil)
	form.AddInputField("Alış Fiyatı", "", 20, nil, nil)
	form.AddInputField("Güncel Fiyat (opsiyonel)", "", 20, nil, nil)

	// Butonlar
	form.AddButton("Kaydet", func() {
		a.saveNewEnvanterSingle(form, turDropdown, cinsDropdown, birimDropdown, selectedKod)
	})
	form.AddButton("İptal", func() {
		a.pages.RemovePage("add-form")
		a.clearAllTables()
		a.app.ForceDraw()
		a.loadData()
		a.loadGrupData()
		a.loadOzetData()
		a.app.SetFocus(a.table)
	})

	// Ana form container
	mainForm := form

	mainForm.SetTitle(" ➕ YENİ ENVANTER EKLE (Tarih: DD.MM.YYYY, Sayılar: 1234,56 veya 1234.56, Güncel Fiyat boş bırakılabilir) ").SetBorder(true)
	mainForm.SetBackgroundColor(tcell.ColorBlack)

	// Modal olarak göster - daha büyük boyut
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(mainForm, 16, 1, true).
			AddItem(nil, 0, 1, false), 80, 1, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("add-form", modal, true, true)
	a.app.SetFocus(form)
}

// saveNewEnvanterSingle tek form ile yeni envanter kaydeder
func (a *App) saveNewEnvanterSingle(form *tview.Form, turDropdown, cinsDropdown, birimDropdown *tview.DropDown, selectedKod string) {
	// Form değerlerini al
	turIndex, turText := turDropdown.GetCurrentOption()
	cinsIndex, cinsText := cinsDropdown.GetCurrentOption()
	miktar := form.GetFormItem(2).(*tview.InputField).GetText()
	birimIndex, birimText := birimDropdown.GetCurrentOption()
	alisTarihi := form.GetFormItem(4).(*tview.InputField).GetText()
	alisFiyatiStr := form.GetFormItem(5).(*tview.InputField).GetText()
	guncelFiyatStr := form.GetFormItem(6).(*tview.InputField).GetText()

	// Ortak validation ve parsing
	miktarVal, alisFiyati, guncelFiyat, err := a.validateAndParseFormData(turIndex, cinsIndex, birimIndex, miktar, alisFiyatiStr, guncelFiyatStr)
	if err != nil {
		a.showMessageWithReturn(err.Error(), form)
		return
	}

	// Alış tarihini parse et
	alisTarihiTime, err := a.parseAlisTarihi(alisTarihi)
	if err != nil {
		a.showMessageWithReturn("Alış tarihi geçerli formatta olmalı! (DD.MM.YYYY)", form)
		return
	}

	// Envanter kaydı oluştur
	envanter := a.createEnvanter(selectedKod, turText, cinsText, miktarVal, alisFiyati, guncelFiyat, birimText, alisTarihiTime)

	// Veritabanına kaydet
	if err := a.saveEnvanterToDatabase(&envanter); err != nil {
		a.showMessageWithReturn(fmt.Sprintf("Kaydetme hatası: %v", err), form)
		return
	}

	// Form'u kapat ve tabloyu güncelle
	a.refreshDataAndCloseForm("add-form", "Envanter başarıyla eklendi!")
}

// updateEnvanterSingle tek form ile envanter günceller
func (a *App) updateEnvanterSingle(form *tview.Form, turDropdown, cinsDropdown, birimDropdown *tview.DropDown, selectedKod string, row int) {
	// Form değerlerini al
	turIndex, turText := turDropdown.GetCurrentOption()
	cinsIndex, cinsText := cinsDropdown.GetCurrentOption()
	miktar := form.GetFormItem(2).(*tview.InputField).GetText()
	birimIndex, birimText := birimDropdown.GetCurrentOption()
	alisTarihi := form.GetFormItem(4).(*tview.InputField).GetText()
	alisFiyatiStr := form.GetFormItem(5).(*tview.InputField).GetText()
	guncelFiyatStr := form.GetFormItem(6).(*tview.InputField).GetText()

	// Ortak validation ve parsing
	miktarVal, alisFiyati, guncelFiyat, err := a.validateAndParseFormData(turIndex, cinsIndex, birimIndex, miktar, alisFiyatiStr, guncelFiyatStr)
	if err != nil {
		a.showMessageWithReturn(err.Error(), form)
		return
	}

	// Alış tarihini parse et
	alisTarihiTime, err := a.parseAlisTarihi(alisTarihi)
	if err != nil {
		a.showMessageWithReturn("Alış tarihi geçerli formatta olmalı! (DD.MM.YYYY)", form)
		return
	}

	// Mevcut kaydın ID'sini bul
	envanterService := services.NewEnvanterService()
	envanterler, err := envanterService.GetAllEnvanterFromDB()
	if err != nil || row-1 >= len(envanterler) {
		a.showMessageWithReturn("Kayıt bulunamadı!", form)
		return
	}

	// Güncellenen envanter kaydı
	envanter := envanterler[row-1]
	envanter.Kod = selectedKod
	envanter.Tur = turText
	envanter.Cins = cinsText
	envanter.Miktar = miktarVal
	envanter.Birim = birimText
	envanter.AlisFiyati = alisFiyati

	// Güncel fiyat güncellemesi - eğer girilmişse güncelle, girilmemişse (0 ise) API'den çekilecek
	if guncelFiyat > 0 {
		envanter.GuncelFiyat = guncelFiyat
	} else {
		// Güncel fiyat girilmemişse 0 olarak ayarla, servis API'den çekecek
		envanter.GuncelFiyat = 0
	}

	if alisTarihi != "" {
		envanter.AlisTarihi = alisTarihiTime // Sadece girilmişse güncelle
	}

	// Veritabanında güncelle
	if err := a.updateEnvanterInDatabase(&envanter); err != nil {
		a.showMessageWithReturn(fmt.Sprintf("Güncelleme başarısız: %v", err), form)
		return
	}

	// Form'u kapat ve tabloyu güncelle
	a.refreshDataAndCloseForm("edit-form", "Envanter başarıyla güncellendi!")
}

// showEditForm seçili kaydı düzenleme formunu gösterir
func (a *App) showEditForm() {
	row, _ := a.table.GetSelection()
	if row <= 0 {
		a.showMessage("Lütfen düzenlemek için bir kayıt seçin!")
		return
	}

	// Mevcut kaydın verilerini veritabanından al
	envanterService := services.NewEnvanterService()
	envanterler, err := envanterService.GetAllEnvanterFromDB()
	if err != nil || row-1 >= len(envanterler) {
		a.showMessage("Kayıt bulunamadı!")
		return
	}

	envanter := envanterler[row-1]

	// Veritabanındaki orijinal değerleri kullan (precision kaybı olmadan)
	tur := envanter.Tur
	cins := envanter.Cins
	birim := envanter.Birim

	// Miktarı decimal kullanarak hassas şekilde formatla
	miktarDecimal := decimal.NewFromFloat(envanter.Miktar)
	miktarStr := miktarDecimal.StringFixed(2)
	// Sondaki sıfırları temizle
	miktarStr = strings.TrimSuffix(miktarStr, ".00")
	miktarStr = strings.TrimSuffix(miktarStr, ".0")

	// Fiyat değerlerini veritabanından decimal kullanarak hassas şekilde formatla
	alisFiyatiDecimal := decimal.NewFromFloat(envanter.AlisFiyati)
	guncelFiyatDecimal := decimal.NewFromFloat(envanter.GuncelFiyat)
	alisFiyati := alisFiyatiDecimal.StringFixed(2)
	guncelFiyat := guncelFiyatDecimal.StringFixed(2)

	// Alış tarihini string'e çevir
	alisTarihiStr := envanter.AlisTarihi.Format("02.01.2006")

	// Tür dropdown - mevcut değeri seç
	turDropdown := tview.NewDropDown().
		SetLabel("Tür").
		SetOptions(turOptions, nil)
	turIndex := findIndex(turOptions, tur)
	if turIndex >= 0 {
		turDropdown.SetCurrentOption(turIndex)
	}

	// Cins dropdown - mevcut değeri seç
	cinsDropdown := tview.NewDropDown().
		SetLabel("Cins")
	cinsOpts := getCinsOptions(tur)
	if len(cinsOpts) > 0 {
		cinsDropdown.SetOptions(cinsOpts, nil)
		cinsIndex := findIndex(cinsOpts, cins)
		if cinsIndex >= 0 {
			cinsDropdown.SetCurrentOption(cinsIndex)
		}
	}

	// Birim dropdown - mevcut değeri seç
	birimDropdown := tview.NewDropDown().
		SetLabel("Birim").
		SetOptions(birimOptions, nil)
	birimIndex := findIndex(birimOptions, birim)
	if birimIndex >= 0 {
		birimDropdown.SetCurrentOption(birimIndex)
	}

	// Kod alanı (gizli) - mevcut kodu belirle
	var selectedKod string
	selectedKod = getCinsCode(cins)

	// Tür değiştiğinde Cins seçeneklerini güncelle
	turDropdown.SetSelectedFunc(func(text string, index int) {
		cinsOptions := getCinsOptions(text)
		cinsDropdown.SetOptions(cinsOptions, func(cinsText string, cinsIndex int) {
			// Cins seçildiğinde kodu belirle
			selectedKod = getCinsCode(cinsText)
		})
	})

	// Ana form - tüm alanlar alt alta
	form := tview.NewForm()
	form.AddFormItem(turDropdown)
	form.AddFormItem(cinsDropdown)
	form.AddInputField("Miktar", miktarStr, 20, nil, nil)
	form.AddFormItem(birimDropdown)
	form.AddInputField("Alış Tarihi", alisTarihiStr, 20, nil, nil)
	form.AddInputField("Alış Fiyatı", alisFiyati, 20, nil, nil)
	form.AddInputField("Güncel Fiyat (opsiyonel)", guncelFiyat, 20, nil, nil)

	// Butonlar
	form.AddButton("Güncelle", func() {
		a.updateEnvanterSingle(form, turDropdown, cinsDropdown, birimDropdown, selectedKod, row)
	})
	form.AddButton("İptal", func() {
		a.pages.RemovePage("edit-form")
		a.clearAllTables()
		a.app.ForceDraw()
		a.loadData()
		a.loadGrupData()
		a.loadOzetData()
		a.app.SetFocus(a.table)
	})

	// Ana form container
	mainForm := form

	mainForm.SetTitle(" ✏️ ENVANTER DÜZENLE (Tarih: DD.MM.YYYY, Sayılar: 1234,56 veya 1234.56, Güncel Fiyat boş bırakılabilir) ").SetBorder(true)
	mainForm.SetBackgroundColor(tcell.ColorBlack)

	// Modal olarak göster - daha büyük boyut
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(mainForm, 16, 1, true).
			AddItem(nil, 0, 1, false), 80, 1, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("edit-form", modal, true, true)
	a.app.SetFocus(form)
}

// showDeleteConfirm silme onayı gösterir
func (a *App) showDeleteConfirm() {
	row, _ := a.table.GetSelection()
	if row <= 0 {
		a.showMessage("Lütfen silmek için bir kayıt seçin!")
		return
	}

	tur := a.table.GetCell(row, 0).Text
	cins := a.table.GetCell(row, 1).Text

	modal := tview.NewModal().
		SetText(fmt.Sprintf("'%s - %s' kaydını silmek istediğinizden emin misiniz?", tur, cins)).
		AddButtons([]string{"Sil", "İptal"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Sil" {
				a.deleteEnvanter(row)
			}
			a.pages.RemovePage("delete-confirm")
			a.app.ForceDraw()
			a.app.SetFocus(a.table)
		})

	a.pages.AddPage("delete-confirm", modal, true, true)
}

// showMessage bilgi mesajı gösterir
func (a *App) showMessage(message string) {
	a.showMessageWithReturn(message, nil)
}

// showMessageWithReturn mesaj gösterir ve kapandığında belirtilen widget'a focus yapar
func (a *App) showMessageWithReturn(message string, returnWidget tview.Primitive) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Tamam"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("message")
			a.app.ForceDraw()
			if returnWidget != nil {
				a.app.SetFocus(returnWidget)
			} else {
				a.app.SetFocus(a.table)
			}
		})

	a.pages.AddPage("message", modal, true, true)
	a.app.SetFocus(modal)
}

// parsePrice fiyat stringini decimal'a çevirir (Türkçe format desteği)
func parsePrice(price string) decimal.Decimal {
	// Para birimi sembollerini ve boşlukları temizle
	price = strings.ReplaceAll(price, " ₺", "")
	price = strings.ReplaceAll(price, "₺", "")
	price = strings.TrimSpace(price)

	// Boş string kontrolü
	if price == "" {
		return decimal.Zero
	}

	// Türkçe format: virgül ondalık ayırıcı, nokta binlik ayırıcı
	// 1.234,56 -> 1234.56
	// Basit yaklaşım: sadece virgülleri noktaya çevir
	price = strings.ReplaceAll(price, ",", ".")

	// Birden fazla nokta varsa (binlik ayırıcı durumu), son nokta hariç diğerlerini kaldır
	dotCount := strings.Count(price, ".")
	if dotCount > 1 {
		// En son noktanın pozisyonunu bul
		lastDotIndex := strings.LastIndex(price, ".")
		beforeLastDot := price[:lastDotIndex]
		afterLastDot := price[lastDotIndex:]

		// Son nokta öncesindeki noktaları kaldır (binlik ayırıcı)
		beforeLastDot = strings.ReplaceAll(beforeLastDot, ".", "")
		price = beforeLastDot + afterLastDot
	}

	// Decimal ile parse et
	result, err := decimal.NewFromString(price)
	if err != nil {
		return decimal.Zero
	}
	return result
} // validateAndParseFormData form verilerini validate eder ve parse eder
func (a *App) validateAndParseFormData(turIndex, cinsIndex, birimIndex int, miktar, alisFiyatiStr, guncelFiyatStr string) (float64, float64, float64, error) {
	// Validasyon
	if turIndex < 0 || cinsIndex < 0 || miktar == "" || birimIndex < 0 {
		return 0, 0, 0, fmt.Errorf("lütfen tüm alanları doldurun")
	}

	// Miktar için decimal kullan
	miktarVal := parsePrice(miktar)
	if miktarVal.IsZero() && miktar != "0" && miktar != "0,0" && miktar != "0.0" {
		return 0, 0, 0, fmt.Errorf("miktar geçerli bir sayı olmalı")
	}

	// Fiyatları decimal ile parse et
	alisFiyati := parsePrice(alisFiyatiStr)
	if alisFiyati.IsZero() && alisFiyatiStr != "0" && alisFiyatiStr != "0,0" && alisFiyatiStr != "0.0" {
		return 0, 0, 0, fmt.Errorf("alış fiyatı geçerli bir sayı olmalı")
	}

	// Güncel fiyat opsiyonel - boş bırakılabilir
	guncelFiyat := parsePrice(guncelFiyatStr)
	if guncelFiyatStr != "" && guncelFiyat.IsZero() && guncelFiyatStr != "0" && guncelFiyatStr != "0,0" && guncelFiyatStr != "0.0" {
		return 0, 0, 0, fmt.Errorf("güncel fiyat geçerli bir sayı olmalı")
	}

	// Float64'e çevir (veritabanı için)
	return miktarVal.InexactFloat64(), alisFiyati.InexactFloat64(), guncelFiyat.InexactFloat64(), nil
}

// parseAlisTarihi alış tarihini parse eder
func (a *App) parseAlisTarihi(alisTarihi string) (time.Time, error) {
	if alisTarihi != "" {
		return time.Parse("02.01.2006", alisTarihi)
	}
	return time.Now(), nil
}

// createEnvanter envanter objesi oluşturur
func (a *App) createEnvanter(selectedKod, turText, cinsText string, miktarVal, alisFiyati, guncelFiyat float64, birimText string, alisTarihiTime time.Time) models.Envanter {
	envanter := models.Envanter{
		Kod:         selectedKod,
		Tur:         turText,
		Cins:        cinsText,
		Miktar:      miktarVal,
		Birim:       birimText,
		AlisFiyati:  alisFiyati,
		GuncelFiyat: guncelFiyat,
		AlisTarihi:  alisTarihiTime,
	}

	// Toplam alış tutarını hesapla
	envanter.ToplamAlis = miktarVal * alisFiyati

	// Eğer güncel fiyat girilmişse hesapla, girilmemişse API'den çekilecek
	if guncelFiyat > 0 {
		envanter.GuncelDegerleriHesapla()
	}

	return envanter
}

// saveEnvanterToDatabase envanter'ı veritabanına kaydeder
func (a *App) saveEnvanterToDatabase(envanter *models.Envanter) error {
	envanterService := services.NewEnvanterService()
	return envanterService.AddEnvanterWithMode(envanter, a.isListMode)
}

// updateEnvanterInDatabase envanter'ı veritabanında günceller
func (a *App) updateEnvanterInDatabase(envanter *models.Envanter) error {
	envanterService := services.NewEnvanterService()
	return envanterService.UpdateEnvanterWithMode(envanter, a.isListMode)
}

// refreshDataAndCloseForm form'u kapatır ve verileri yeniler
func (a *App) refreshDataAndCloseForm(formName, successMessage string) {
	a.pages.RemovePage(formName)

	// Tabloları tamamen temizle
	a.clearAllTables()

	// Ekranı temizle ve yeniden çiz
	a.app.ForceDraw()

	// Verileri yeniden yükle
	a.loadData()
	a.loadGrupData()
	a.loadOzetData()

	// Focus'u geri getir
	a.app.SetFocus(a.table)

	// Ekranı zorla yenile ve mesaj göster
	go func() {
		time.Sleep(100 * time.Millisecond)
		a.app.QueueUpdateDraw(func() {
			a.app.ForceDraw()
			a.showMessage(successMessage)
		})
	}()
}

// clearAllTables tüm tabloları temizler
func (a *App) clearAllTables() {
	a.table.Clear()
	a.grupTable.Clear()
	a.ozetTable.Clear()
} // findIndex string slice'ında belirtilen değerin indeksini bulur
func findIndex(slice []string, value string) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

// deleteEnvanter seçili envanter kaydını siler
func (a *App) deleteEnvanter(row int) {
	// Mevcut kaydın ID'sini bul
	envanterService := services.NewEnvanterService()
	envanterler, err := envanterService.GetAllEnvanterFromDB()
	if err != nil || row-1 >= len(envanterler) {
		a.showMessage("Kayıt bulunamadı!")
		return
	}

	// Silinecek envanter
	envanter := envanterler[row-1]

	// Veritabanından sil
	err = envanterService.DeleteEnvanter(envanter.ID)
	if err != nil {
		a.showMessage(fmt.Sprintf("Silme başarısız: %v", err))
		return
	}

	// Başarılı, verileri yenile
	a.loadData()
	a.loadGrupData()
	a.loadOzetData()

	// Başarı mesajını göster
	go func() {
		a.showMessage("Envanter başarıyla silindi!")
	}()
}

// startAutoUpdate otomatik fiyat güncellemeyi başlatır (5 dakikada bir)
func (a *App) startAutoUpdate() {
	a.updateTicker = time.NewTicker(5 * time.Minute)

	go func() {
		for {
			select {
			case <-a.updateTicker.C:
				log.Printf("Otomatik fiyat güncelleme başlatılıyor...")
				err := a.envanterService.UpdateGuncelFiyatlar()
				if err != nil {
					log.Printf("UYARI: Otomatik fiyat güncelleme başarısız: %v", err)
				} else {
					log.Printf("Otomatik fiyat güncelleme başarılı")
					// UI'yi güncelle
					a.app.QueueUpdateDraw(func() {
						a.loadData()
						a.loadGrupData()
						a.loadOzetData()
					})
				}
			case <-a.stopChan:
				log.Printf("Otomatik fiyat güncelleme durduruldu")
				return
			}
		}
	}()

	log.Printf("Otomatik fiyat güncelleme başlatıldı (5 dakikada bir)")
}

// stopAutoUpdate otomatik fiyat güncellemeyi durdurur
func (a *App) stopAutoUpdate() {
	if a.updateTicker != nil {
		a.updateTicker.Stop()
	}

	// stopChan'e sinyal gönder
	select {
	case a.stopChan <- true:
	default:
		// Kanal dolu ise, zaten durdurulmuş demektir
	}
}

// formatMoney para birimini formatlar (decimal kullanarak hassas)
func formatMoney(amount float64) string {
	// Float64'ü decimal'a çevir
	d := decimal.NewFromFloat(amount)

	// 2 ondalık basamakla formatla
	formatted := d.StringFixed(2)

	// Türkçe format'a çevir: 1234.56 -> 1.234,56
	return formatTurkishNumber(formatted)
}

// formatTurkishNumber sayıyı Türkçe formata çevirir (binlik: nokta, ondalık: virgül)
func formatTurkishNumber(numberStr string) string {
	// Ondalık kısmı ayır
	parts := strings.Split(numberStr, ".")
	intPart := parts[0]
	decimalPart := ""

	if len(parts) > 1 {
		decimalPart = parts[1]
		// Sondaki sıfırları kaldır
		decimalPart = strings.TrimRight(decimalPart, "0")
	}

	// Binlik ayırıcı ekle (nokta)
	result := addThousandSeparator(intPart)

	// Ondalık kısım varsa virgül ile ekle
	if decimalPart != "" {
		result = result + "," + decimalPart
	}

	return result
}

// formatQuantity miktarı formatlar (decimal kullanarak hassas)
func formatQuantity(amount float64, unit string) string {
	// Float64'ü decimal'a çevir
	d := decimal.NewFromFloat(amount)

	if unit == "adet" {
		// Adet ise tam sayı olarak göster
		return d.Truncate(0).String()
	}

	// Diğer birimler için ondalıklı göster ama gereksiz sıfırları temizle
	if d.Equal(d.Truncate(0)) {
		// Tam sayı ise ondalık gösterme
		return formatTurkishNumber(d.Truncate(0).String())
	}

	// 2 ondalık basamakla formatla
	formatted := d.StringFixed(2)
	return formatTurkishNumber(formatted)
}

// addThousandSeparator sayıya binlik ayırıcı ekler
func addThousandSeparator(s string) string {
	// Negatif sayıları kontrol et
	negative := false
	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}

	n := len(s)
	if n <= 3 {
		if negative {
			return "-" + s
		}
		return s
	}

	var result strings.Builder
	for i, digit := range s {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteString(".")
		}
		result.WriteRune(digit)
	}

	if negative {
		return "-" + result.String()
	}
	return result.String()
}

// updateScrollIndicators scroll indicator'ları günceller
func (a *App) updateScrollIndicators() {
	// Envanter tablosu için scroll indicator güncelle
	envanterRowCount := a.table.GetRowCount()
	envanterRow, _ := a.table.GetSelection()

	if envanterRowCount > 0 {
		// Scroll pozisyonunu hesapla (0-100 arası)
		scrollPercent := float64(envanterRow) / float64(envanterRowCount-1) * 100

		var indicator string
		if envanterRowCount <= 1 {
			indicator = "│"
		} else if scrollPercent <= 10 {
			indicator = "▲\n│\n│\n│"
		} else if scrollPercent >= 90 {
			indicator = "│\n│\n│\n▼"
		} else {
			indicator = "│\n■\n│\n│"
		}

		a.envanterScrollIndicator.SetText(indicator)
	}

	// Grup tablosu için scroll indicator güncelle
	grupRowCount := a.grupTable.GetRowCount()
	grupRow, _ := a.grupTable.GetSelection()

	if grupRowCount > 0 {
		// Scroll pozisyonunu hesapla (0-100 arası)
		scrollPercent := float64(grupRow) / float64(grupRowCount-1) * 100

		var indicator string
		if grupRowCount <= 1 {
			indicator = "│"
		} else if scrollPercent <= 10 {
			indicator = "▲\n│\n│"
		} else if scrollPercent >= 90 {
			indicator = "│\n│\n▼"
		} else {
			indicator = "│\n■\n│"
		}

		a.grupScrollIndicator.SetText(indicator)
	}
}
