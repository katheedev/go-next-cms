package models

import "time"

type UserRole string

const (
	RoleSubmitter UserRole = "submitter"
	RoleAdmin     UserRole = "admin"
)

type DealStatus string

const (
	DealDraft     DealStatus = "draft"
	DealPending   DealStatus = "pending"
	DealApproved  DealStatus = "approved"
	DealPublished DealStatus = "published"
	DealRejected  DealStatus = "rejected"
	DealExpired   DealStatus = "expired"
)

type Country struct {
	ID              int64
	Code            string
	Name            string
	DefaultLanguage string
}

type City struct {
	ID        int64
	CountryID int64
	Name      string
	Slug      string
}

type Category struct {
	ID   int64
	Name string
	Slug string
}

type Merchant struct {
	ID       int64
	Name     string
	Slug     string
	LogoURL  string
	Contact  string
	Verified bool
}

type DealType struct {
	ID   int64
	Code string
	Name string
}

type Deal struct {
	ID              int64
	Title           string
	Slug            string
	Description     string
	CountryID       int64
	CountryCode     string
	CityID          int64
	CityName        string
	CategoryID      int64
	CategoryName    string
	CategorySlug    string
	MerchantID      *int64
	MerchantName    *string
	DealTypeID      int64
	DealTypeName    string
	StartAt         time.Time
	EndAt           time.Time
	Featured        bool
	ImageURL        string
	Status          DealStatus
	CreatedByUserID int64
	RejectionReason *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type DealTranslation struct {
	DealID      int64
	Lang        string
	Title       string
	Description string
}

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Name         string
	Role         UserRole
	CreatedAt    time.Time
}

type AdminConfig struct {
	Key   string
	Value []byte
}

type DealFilter struct {
	CountryCode  string
	CitySlug     string
	CategorySlug string
	DealTypeID   int64
	MerchantID   int64
	Search       string
	EndingSoon   bool
	Page         int
	PageSize     int
}
