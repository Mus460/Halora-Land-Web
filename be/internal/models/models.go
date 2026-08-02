package models

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// Enums as typed string constants (ARCHITECTURE.md §4 porting note).

type Role string

const (
	RoleAdmin Role = "ADMIN"
	RoleUser  Role = "USER"
	RoleDemo  Role = "DEMO"
)

type TipeProyek string

const (
	TipeProyekGedung TipeProyek = "gedung"
	TipeProyekInfra  TipeProyek = "infra"
)

type RoleTimProyek string

const (
	TimRoleOwner  RoleTimProyek = "owner"
	TimRoleEditor RoleTimProyek = "editor"
	TimRoleViewer RoleTimProyek = "viewer"
)

type KategoriPekerjaan string

const (
	KategoriPersiapan  KategoriPekerjaan = "persiapan"
	KategoriPondasi    KategoriPekerjaan = "pondasi"
	KategoriBeton      KategoriPekerjaan = "beton"
	KategoriKanopi     KategoriPekerjaan = "kanopi"
	KategoriBaja       KategoriPekerjaan = "baja"
	KategoriTangga     KategoriPekerjaan = "tangga"
	KategoriAtap       KategoriPekerjaan = "atap"
	KategoriDinding    KategoriPekerjaan = "dinding"
	KategoriPlesteran  KategoriPekerjaan = "plesteran"
	KategoriAcian      KategoriPekerjaan = "acian"
	KategoriKeramik    KategoriPekerjaan = "keramik"
	KategoriPaving     KategoriPekerjaan = "paving"
	KategoriPengecatan KategoriPekerjaan = "pengecatan"
	KategoriPintu      KategoriPekerjaan = "pintu"
	KategoriInterior   KategoriPekerjaan = "interior"
	KategoriToilet     KategoriPekerjaan = "toilet"
	KategoriMEP        KategoriPekerjaan = "mep"
	KategoriCustom     KategoriPekerjaan = "custom"
)

type MetodeHitung string

const (
	MetodeAHSP         MetodeHitung = "ahsp"
	MetodeManual       MetodeHitung = "manual"
	MetodeHargaBorong  MetodeHitung = "harga_borong"
	MetodeHargaManual  MetodeHitung = "harga_manual"
	MetodeHargaCustom  MetodeHitung = "harga_custom"
)

type TipeKomponen string

const (
	KomponenMaterial TipeKomponen = "material"
	KomponenUpah     TipeKomponen = "upah"
	KomponenAlat     TipeKomponen = "alat"
)

type StatusInvoice string

const (
	InvoiceDraft StatusInvoice = "draft"
	InvoiceSent  StatusInvoice = "sent"
	InvoicePaid  StatusInvoice = "paid"
)

type StatusFeedback string

const (
	FeedbackOpen       StatusFeedback = "open"
	FeedbackInProgress StatusFeedback = "in_progress"
	FeedbackResolved   StatusFeedback = "resolved"
	FeedbackClosed     StatusFeedback = "closed"
)

// Tables

type User struct {
	ID          int32     `json:"id"`
	NamaLengkap string    `json:"namaLengkap"`
	Email       string    `json:"email"`
	Role        Role      `json:"role"`
	AccountType string    `json:"accountType"`
	IsDemo      bool      `json:"isDemo"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Proyek struct {
	ID          int32            `json:"id"`
	UserID      int32            `json:"userId"`
	NamaProyek  string           `json:"namaProyek"`
	Lokasi      *string          `json:"lokasi"`
	Tipe        TipeProyek       `json:"tipe"`
	NilaiKontrak *decimal.Decimal `json:"nilaiKontrak"`
	Timeline    *string          `json:"timeline"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
}

type TimProyek struct {
	ID        int32          `json:"id"`
	ProyekID  int32          `json:"proyekId"`
	UserID    int32          `json:"userId"`
	Role      RoleTimProyek  `json:"role"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

type Pekerjaan struct {
	ID              int32            `json:"id"`
	ProyekID        int32            `json:"proyekId"`
	Kategori        KategoriPekerjaan `json:"kategori"`
	UraianPekerjaan string           `json:"uraianPekerjaan"`
	Volume          decimal.Decimal  `json:"volume"`
	Satuan          string           `json:"satuan"`
	HargaSatuan     decimal.Decimal  `json:"hargaSatuan"`
	TotalBiaya      decimal.Decimal  `json:"totalBiaya"`
	MetodeHitung    MetodeHitung     `json:"metodeHitung"`
	LevelPekerjaan  *string          `json:"levelPekerjaan"`
	TipePekerjaan   *string          `json:"tipePekerjaan"`
	MasterAnalisaID *int32           `json:"masterAnalisaId"`
	Waktu           *decimal.Decimal `json:"waktu"`
	TotalWaktu      *decimal.Decimal `json:"totalWaktu"`
	CreatedAt       time.Time        `json:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt"`
	DetailAnalisa   []DetailAnalisa  `json:"detailAnalisa,omitempty"`
}

type DetailAnalisa struct {
	ID             int32            `json:"id"`
	PekerjaanID    int32            `json:"pekerjaanId"`
	MasterHargaID  *int32           `json:"masterHargaId"`
	MasterAnalisaID *int32          `json:"masterAnalisaId"`
	Nama           string           `json:"nama"`
	Satuan         string           `json:"satuan"`
	Koef           decimal.Decimal  `json:"koef"`
	HargaSatuan    decimal.Decimal  `json:"hargaSatuan"`
	TotalBiaya     decimal.Decimal  `json:"totalBiaya"`
	Tipe           TipeKomponen     `json:"tipe"`
	SnapshotAt     time.Time        `json:"snapshotAt"`
	SourceKode     *string          `json:"sourceKode"`
}

type MasterAnalisa struct {
	ID          int32             `json:"id"`
	Kode        string            `json:"kode"`
	Nama        string            `json:"nama"`
	Level       int32             `json:"level"`
	ParentID    *int32            `json:"parentId"`
	Satuan      *string           `json:"satuan"`
	HargaSatuan *decimal.Decimal  `json:"hargaSatuan"`
	Kategori    *string           `json:"kategori"`
	IsGlobal    bool              `json:"isGlobal"`
	UserID      *int32            `json:"userId"`
	IsSystem    bool              `json:"isSystem"`
	AHSPKode    *string           `json:"ahspKode"`
	AHSPSheet   *string           `json:"ahspSheet"`
	BiayaUmum   decimal.Decimal   `json:"biayaUmum"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	Children    []MasterAnalisa   `json:"children,omitempty"`
}

type RincianAnalisa struct {
	ID             int32            `json:"id"`
	MasterAnalisaID int32           `json:"masterAnalisaId"`
	KomponenID     *int32           `json:"komponenId"`
	Koef           decimal.Decimal  `json:"koef"`
	Tipe           TipeKomponen     `json:"tipe"`
	Nama           *string          `json:"nama"`
	Satuan         *string          `json:"satuan"`
	HargaSatuan    *decimal.Decimal `json:"hargaSatuan"`
	JumlahHarga    *decimal.Decimal `json:"jumlahHarga"`
	KodeReferensi  *string          `json:"kodeReferensi"`
	Waktu          *decimal.Decimal `json:"waktu"`
	Urutan         int32            `json:"urutan"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`
}

type MasterHarga struct {
	ID        int32           `json:"id"`
	Nama      string          `json:"nama"`
	Satuan    string          `json:"satuan"`
	Harga     decimal.Decimal `json:"harga"`
	Kategori  TipeKomponen    `json:"kategori"`
	IsGlobal  bool            `json:"isGlobal"`
	UserID    *int32          `json:"userId"`
	KodeAHSP  *string         `json:"kodeAHSP"`
	IsSystem  bool            `json:"isSystem"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

type Rekap struct {
	ID        int32            `json:"id"`
	ProyekID  int32            `json:"proyekId"`
	Kategori  string           `json:"kategori"`
	Uraian    string           `json:"uraian"`
	Urutan    int32            `json:"urutan"`
	Margin    *decimal.Decimal `json:"margin"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

type Invoice struct {
	ID        int32           `json:"id"`
	ProyekID  int32           `json:"proyekId"`
	Nomor     string          `json:"nomor"`
	Tanggal   time.Time       `json:"tanggal"`
	Total     decimal.Decimal `json:"total"`
	Status    StatusInvoice   `json:"status"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

type Logistik struct {
	ID           int32            `json:"id"`
	ProyekID     int32            `json:"proyekId"`
	NamaMaterial string           `json:"namaMaterial"`
	Satuan       string           `json:"satuan"`
	Volume       decimal.Decimal  `json:"volume"`
	HargaSatuan  decimal.Decimal  `json:"hargaSatuan"`
	TotalBiaya   decimal.Decimal  `json:"totalBiaya"`
	Tanggal      *time.Time       `json:"tanggal"`
	Keterangan   *string          `json:"keterangan"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
}

type Realisasi struct {
	ID         int32            `json:"id"`
	ProyekID   int32            `json:"proyekId"`
	Tanggal    time.Time        `json:"tanggal"`
	Kategori   string           `json:"kategori"`
	Jumlah     decimal.Decimal  `json:"jumlah"`
	Keterangan *string          `json:"keterangan"`
	CreatedAt  time.Time        `json:"createdAt"`
	UpdatedAt  time.Time        `json:"updatedAt"`
}

type Feedback struct {
	ID        int32          `json:"id"`
	UserID    int32          `json:"userId"`
	Subject   string         `json:"subject"`
	Message   string         `json:"message"`
	Status    StatusFeedback `json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	Replies   []FeedbackReply `json:"replies,omitempty"`
}

type FeedbackReply struct {
	ID         int32     `json:"id"`
	FeedbackID int32     `json:"feedbackId"`
	UserID     int32     `json:"userId"`
	Message    string    `json:"message"`
	IsAdmin    bool      `json:"isAdmin"`
	CreatedAt  time.Time `json:"createdAt"`
}

type News struct {
	ID        int32     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AuditLog struct {
	ID          int32           `json:"id"`
	ProyekID    *int32          `json:"proyekId"`
	PekerjaanID *int32          `json:"pekerjaanId"`
	UserID      int32           `json:"userId"`
	Action      string          `json:"action"`
	EntityType  string          `json:"entityType"`
	EntityID    *int32          `json:"entityId"`
	OldValue    json.RawMessage `json:"oldValue,omitempty"`
	NewValue    json.RawMessage `json:"newValue,omitempty"`
	Description *string         `json:"description"`
	IPAddress   *string         `json:"ipAddress"`
	UserAgent   *string         `json:"userAgent"`
	CreatedAt   time.Time       `json:"createdAt"`
}
