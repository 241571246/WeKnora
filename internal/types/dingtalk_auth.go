package types

import "time"

const DingTalkAuthProvider = "dingtalk"

type DingTalkExternalIdentity struct {
	ID                  string     `json:"id" gorm:"type:varchar(36);primaryKey"`
	Provider            string     `json:"provider" gorm:"type:varchar(32);not null;uniqueIndex:uq_external_auth_identity,priority:1"`
	CorpID              string     `json:"corp_id" gorm:"type:varchar(128);not null;uniqueIndex:uq_external_auth_identity,priority:2"`
	ExternalUserID      string     `json:"external_user_id" gorm:"type:varchar(128);not null;uniqueIndex:uq_external_auth_identity,priority:3"`
	UnionID             string     `json:"union_id,omitempty" gorm:"type:varchar(128);index"`
	UserID              *string    `json:"user_id,omitempty" gorm:"type:varchar(36);index"`
	DisplayName         string     `json:"display_name" gorm:"type:varchar(255);not null"`
	Email               string     `json:"email,omitempty" gorm:"type:varchar(255)"`
	Mobile              string     `json:"mobile,omitempty" gorm:"type:varchar(64)"`
	Status              string     `json:"status" gorm:"type:varchar(32);not null"`
	BindingMethod       string     `json:"binding_method" gorm:"type:varchar(32);not null"`
	RawProfile          JSON       `json:"-" gorm:"type:jsonb;not null;default:'{}'"`
	LastAuthenticatedAt *time.Time `json:"last_authenticated_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (DingTalkExternalIdentity) TableName() string { return "external_auth_identities" }

type DingTalkIdentityProfile struct {
	ExternalUserID string
	UnionID        string
	DisplayName    string
	Email          string
	Mobile         string
	Active         bool
	Raw            map[string]any
}

type DingTalkAuthConfigResponse struct {
	Success     bool   `json:"success"`
	Enabled     bool   `json:"enabled"`
	CorpID      string `json:"corp_id,omitempty"`
	ClientID    string `json:"client_id,omitempty"`
	InsideLogin bool   `json:"inside_login"`
}

type DingTalkAuthURLResponse struct {
	Success          bool   `json:"success"`
	AuthorizationURL string `json:"authorization_url"`
	Nonce            string `json:"-"`
}
