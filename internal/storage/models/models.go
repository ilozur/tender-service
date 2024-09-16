package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TenderServiceType string
type TenderStatus string
type OrganizationType string
type BidStatus string
type BidAuthorType string

const (
	Construction TenderServiceType = "Construction"
	Delivery     TenderServiceType = "Delivery"
	Manufacture  TenderServiceType = "Manufacture"
)

const (
	TenderCreated   TenderStatus = "Created"
	TenderPublished TenderStatus = "Published"
	TenderClosed    TenderStatus = "Closed"
)

const (
	IEOrganization  OrganizationType = "IE"
	LLCOrganization OrganizationType = "LLC"
	JSCOrganization OrganizationType = "JSC"
)

const (
	BidCreated   BidStatus = "Created"
	BidPublished BidStatus = "Published"
	BidCanceled  BidStatus = "Canceled"
	BidApproved  BidStatus = "Approved"
	BidRejected  BidStatus = "Rejected"
)

const (
	BidAuthorOrganization BidAuthorType = "Organization"
	BidAuthorUser         BidAuthorType = "User"
)

type Tender struct {
	gorm.Model
	ID               uuid.UUID         `gorm:"type:uuid;default:uuid_generate_v4()"`
	Name             string            `gorm:"type:varchar(100);not null"`
	Description      string            `gorm:"type:varchar(500)"`
	ServiceType      TenderServiceType `gorm:"type:tender_service_type"`
	Status           TenderStatus      `gorm:"type:tender_status;not null"`
	EmployeeUsername string            `gorm:"not null"`
	OrganizationID   uuid.UUID         `gorm:"not null"`
	Organization     Organization
	Version          uint      `gorm:"not null;default:1"`
	CreatedAt        time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type TenderVersion struct {
	gorm.Model
	ID               uuid.UUID         `gorm:"type:uuid;default:uuid_generate_v4()"`
	TenderID         uuid.UUID         `gorm:"type:uuid;"`
	Name             string            `gorm:"type:varchar(100);not null"`
	Description      string            `gorm:"type:varchar(500)"`
	ServiceType      TenderServiceType `gorm:"type:tender_service_type"`
	Status           TenderStatus      `gorm:"type:tender_status;not null"`
	EmployeeUsername string            `gorm:"not null"`
	OrganizationID   uuid.UUID         `gorm:"not null"`
	Organization     Organization
	Version          uint      `gorm:"not null;default:1"`
	CreatedAt        time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type Bid struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Description string    `gorm:"type:varchar(500)"`
	Status      BidStatus `gorm:"type:bid_status;not null"`
	TenderID    uuid.UUID
	Tender      Tender

	EmployeeUsername string    `gorm:"not null"`
	OrganizationID   uuid.UUID `gorm:"not null"`
	Organization     Organization

	AuthorType BidAuthorType `gorm:"type:bid_author_type;not null"`
	Version    uint32        `gorm:"default:1"`
	CreatedAt  time.Time
}

type BidVersion struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Description string    `gorm:"type:varchar(500)"`
	Status      BidStatus `gorm:"type:bid_status;not null"`

	TenderID uuid.UUID
	Tender   Tender

	BidID uuid.UUID
	Bid   Bid

	EmployeeUsername string    `gorm:"not null"`
	OrganizationID   uuid.UUID `gorm:"not null"`
	Organization     Organization

	AuthorType BidAuthorType `gorm:"type:bid_author_type;not null"`
	Version    uint32        `gorm:"default:1"`
	CreatedAt  time.Time
}

type BidFeedback struct {
	gorm.Model
	ID       uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Feedback string    `gorm:"type:varchar(1000);not null"`

	BidID uuid.UUID `gorm:"not null"`
	Bid   Bid

	EmployeeUsername string    `gorm:"not null"`
	OrganizationID   uuid.UUID `gorm:"not null"`
	Organization     Organization

	CreatedAt time.Time
}

type Organization struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Description string
	Type        OrganizationType `gorm:"type:organization_type"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
type Employee struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Username  string    `gorm:"type:varchar(50);unique_index;not null "`
	FirstName string    `gorm:"type:varchar(50)"`
	LastName  string    `gorm:"type:varchar(50)"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type OrganizationResponsible struct {
	ID             uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	OrganizationID uuid.UUID
	Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	EmployeeID     uuid.UUID    `gorm:"column:user_id"`
	Employee       Employee     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Tabler interface {
	TableName() string
}

func (Organization) TableName() string {
	return "organization"
}
func (Employee) TableName() string {
	return "employee"
}

func (OrganizationResponsible) TableName() string {
	return "organization_responsible"
}
