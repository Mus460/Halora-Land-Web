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

type ProjectType string

const (
	ProjectTypeBuilding       ProjectType = "building"
	ProjectTypeInfrastructure ProjectType = "infrastructure"
)

type TeamRole string

const (
	TeamRoleOwner  TeamRole = "owner"
	TeamRoleEditor TeamRole = "editor"
	TeamRoleViewer TeamRole = "viewer"
)

type WorkCategory string

const (
	CategoryPreparation WorkCategory = "preparation"
	CategoryFoundation  WorkCategory = "foundation"
	CategoryConcrete    WorkCategory = "concrete"
	CategoryCanopy      WorkCategory = "canopy"
	CategorySteel       WorkCategory = "steel"
	CategoryStairs      WorkCategory = "stairs"
	CategoryRoof        WorkCategory = "roof"
	CategoryWall        WorkCategory = "wall"
	CategoryPlastering  WorkCategory = "plastering"
	CategoryFinishing   WorkCategory = "finishing"
	CategoryTiles       WorkCategory = "tiles"
	CategoryPaving      WorkCategory = "paving"
	CategoryPainting    WorkCategory = "painting"
	CategoryDoors       WorkCategory = "doors"
	CategoryInterior    WorkCategory = "interior"
	CategoryToilet      WorkCategory = "toilet"
	CategoryMEP         WorkCategory = "mep"
	CategoryCustom      WorkCategory = "custom"
)

type CalculationMethod string

const (
	MethodAHSP        CalculationMethod = "ahsp"
	MethodManual      CalculationMethod = "manual"
	MethodLumpSum     CalculationMethod = "lump_sum"
	MethodManualPrice CalculationMethod = "manual_price"
	MethodCustomPrice CalculationMethod = "custom_price"
)

type ComponentType string

const (
	ComponentMaterial  ComponentType = "material"
	ComponentLabor     ComponentType = "labor"
	ComponentEquipment ComponentType = "equipment"
)

type InvoiceStatus string

const (
	InvoiceDraft InvoiceStatus = "draft"
	InvoiceSent  InvoiceStatus = "sent"
	InvoicePaid  InvoiceStatus = "paid"
)

type FeedbackStatus string

const (
	FeedbackOpen       FeedbackStatus = "open"
	FeedbackInProgress FeedbackStatus = "in_progress"
	FeedbackResolved   FeedbackStatus = "resolved"
	FeedbackClosed     FeedbackStatus = "closed"
)

type TransactionType string

const (
	TransactionExpense TransactionType = "expense"
	TransactionIncome  TransactionType = "income"
)

type TransactionStatus string

const (
	TransactionDraft    TransactionStatus = "draft"
	TransactionApproved TransactionStatus = "approved"
	TransactionReverted TransactionStatus = "reverted"
)

// Tables

type User struct {
	ID          int32     `json:"id"`
	FullName    string    `json:"fullName"`
	Email       string    `json:"email"`
	Role        Role      `json:"role"`
	AccountType string    `json:"accountType"`
	IsDemo      bool      `json:"isDemo"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Project struct {
	ID             int32            `json:"id"`
	UserID         int32            `json:"userId"`
	Name           string           `json:"name"`
	Location       *string          `json:"location"`
	Type           ProjectType      `json:"type"`
	IsPitching     bool             `json:"isPitching"`
	IsDone         bool             `json:"isDone"`
	ContractValue  *decimal.Decimal `json:"contractValue"`
	TimelineMonths int              `json:"timelineMonths"`
	TimelineDays   int              `json:"timelineDays"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`
}

type ProjectTeam struct {
	ID        int32     `json:"id"`
	ProjectID int32     `json:"projectId"`
	UserID    int32     `json:"userId"`
	Role      TeamRole  `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type WorkItem struct {
	ID                int32             `json:"id"`
	ProjectID         int32             `json:"projectId"`
	Category          WorkCategory      `json:"category"`
	Description       string            `json:"description"`
	Volume            decimal.Decimal   `json:"volume"`
	Unit              string            `json:"unit"`
	UnitPrice         decimal.Decimal   `json:"unitPrice"`
	TotalCost         decimal.Decimal   `json:"totalCost"`
	CalculationMethod CalculationMethod `json:"calculationMethod"`
	Level             *string           `json:"level"`
	Type              *string           `json:"type"`
	AnalysisMasterID  *int32            `json:"analysisMasterId"`
	Duration          *decimal.Decimal  `json:"duration"`
	TotalDuration     *decimal.Decimal  `json:"totalDuration"`
	CreatedAt         time.Time         `json:"createdAt"`
	UpdatedAt         time.Time         `json:"updatedAt"`
	ItemDetails       []WorkItemDetail  `json:"itemDetails,omitempty"`
}

type WorkItemProgressLog struct {
	ID         int32     `json:"id"`
	WorkItemID int32     `json:"workItemId"`
	Progress   int       `json:"progress"`
	Note       *string   `json:"note"`
	CreatedAt  time.Time `json:"createdAt"`
}

type WorkItemDetail struct {
	ID               int32           `json:"id"`
	WorkItemID       int32           `json:"workItemId"`
	PriceMasterID    *int32          `json:"priceMasterId"`
	AnalysisMasterID *int32          `json:"analysisMasterId"`
	Name             string          `json:"name"`
	Unit             string          `json:"unit"`
	Coefficient      decimal.Decimal `json:"coefficient"`
	UnitPrice        decimal.Decimal `json:"unitPrice"`
	TotalCost        decimal.Decimal `json:"totalCost"`
	Type             ComponentType   `json:"type"`
	SnapshotAt       time.Time       `json:"snapshotAt"`
	SourceCode       *string         `json:"sourceCode"`
}

type AnalysisMaster struct {
	ID          int32            `json:"id"`
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	Level       int32            `json:"level"`
	ParentID    *int32           `json:"parentId"`
	Unit        *string          `json:"unit"`
	UnitPrice   *decimal.Decimal `json:"unitPrice"`
	Category    *string          `json:"category"`
	IsGlobal    bool             `json:"isGlobal"`
	UserID      *int32           `json:"userId"`
	IsSystem    bool             `json:"isSystem"`
	AHSPCode    *string          `json:"ahspCode"`
	AHSPSheet   *string          `json:"ahspSheet"`
	GeneralCost decimal.Decimal  `json:"generalCost"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
	Children    []AnalysisMaster `json:"children,omitempty"`
}

type AnalysisComponent struct {
	ID               int32            `json:"id"`
	AnalysisMasterID int32            `json:"analysisMasterId"`
	ComponentID      *int32           `json:"componentId"`
	Coefficient      decimal.Decimal  `json:"coefficient"`
	Type             ComponentType    `json:"type"`
	Name             *string          `json:"name"`
	Unit             *string          `json:"unit"`
	UnitPrice        *decimal.Decimal `json:"unitPrice"`
	TotalPrice       *decimal.Decimal `json:"totalPrice"`
	ReferenceCode    *string          `json:"referenceCode"`
	Duration         *decimal.Decimal `json:"duration"`
	Sequence         int32            `json:"sequence"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
}

type PriceMaster struct {
	ID        int32           `json:"id"`
	Name      string          `json:"name"`
	Unit      string          `json:"unit"`
	Price     decimal.Decimal `json:"price"`
	Type      ComponentType   `json:"type"`
	IsGlobal  bool            `json:"isGlobal"`
	UserID    *int32          `json:"userId"`
	AHSPCode  *string         `json:"ahspCode"`
	IsSystem  bool            `json:"isSystem"`
CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

type Client struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	Address   *string   `json:"address"`
	Contact   *string   `json:"contact"`
	UserID    int32     `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Recap struct {
	ID          int32            `json:"id"`
	ProjectID   int32            `json:"projectId"`
	Category    string           `json:"category"`
	Description string           `json:"description"`
	Sequence    int32            `json:"sequence"`
	Margin      *decimal.Decimal `json:"margin"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
}

type Invoice struct {
	ID                   int32           `json:"id"`
	ProjectID            int32           `json:"projectId"`
	Number               string          `json:"number"`
	Date                 time.Time       `json:"date"`
	DueDate              *time.Time      `json:"dueDate"`
	PONumber             *string         `json:"poNumber"`
	BuyerName            *string         `json:"buyerName"`
	BuyerAddress         *string         `json:"buyerAddress"`
	BuyerContact         *string         `json:"buyerContact"`
	Discount             decimal.Decimal `json:"discount"`
	TaxRate              decimal.Decimal `json:"taxRate"`
	PaymentBank          *string         `json:"paymentBank"`
	PaymentAccountNumber *string         `json:"paymentAccountNumber"`
	PaymentAccountName   *string         `json:"paymentAccountName"`
	Notes                *string         `json:"notes"`
	FinanceName          *string         `json:"financeName"`
	Total                decimal.Decimal `json:"total"`
	Status               InvoiceStatus   `json:"status"`
	Items                []InvoiceItem   `json:"items,omitempty"`
	CreatedAt            time.Time       `json:"createdAt"`
	UpdatedAt            time.Time       `json:"updatedAt"`
}

type InvoiceItem struct {
	ID          int32           `json:"id"`
	InvoiceID   int32           `json:"invoiceId"`
	Description string          `json:"description"`
	Qty         decimal.Decimal `json:"qty"`
	Unit        string          `json:"unit"`
	UnitPrice   decimal.Decimal `json:"unitPrice"`
}

type Logistics struct {
	ID           int32           `json:"id"`
	ProjectID    int32           `json:"projectId"`
	MaterialName string          `json:"materialName"`
	Unit         string          `json:"unit"`
	Volume       decimal.Decimal `json:"volume"`
	UnitPrice    decimal.Decimal `json:"unitPrice"`
	TotalCost    decimal.Decimal `json:"totalCost"`
	Date         *time.Time      `json:"date"`
	Description  *string         `json:"description"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}

type Transaction struct {
	ID          int32             `json:"id"`
	ProjectID   int32             `json:"projectId"`
	Date        time.Time         `json:"date"`
	Category    string            `json:"category"`
	Amount      decimal.Decimal   `json:"amount"`
	Description *string           `json:"description"`
	Type        TransactionType   `json:"type"`
	Status      TransactionStatus `json:"status"`
	LogisticsID *int32            `json:"logisticsId"`
	InvoiceID   *int32            `json:"invoiceId"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

type Feedback struct {
	ID        int32           `json:"id"`
	UserID    int32           `json:"userId"`
	Subject   string          `json:"subject"`
	Message   string          `json:"message"`
	Status    FeedbackStatus  `json:"status"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
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
	ProjectID   *int32          `json:"projectId"`
	WorkItemID  *int32          `json:"workItemId"`
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
