package types

import (
	"sort"
	"time"

	"gorm.io/gorm"
)

// KBCapability is a single independently grantable knowledge-base operation.
// Keep these values stable: they are persisted in vone_kb_member_capabilities.
type KBCapability string

const (
	KBCapabilityAIQuery          KBCapability = "kb.ai.query"
	KBCapabilityMetadataRead     KBCapability = "kb.metadata.read"
	KBCapabilityDocumentsList    KBCapability = "kb.documents.list"
	KBCapabilityDocumentPreview  KBCapability = "kb.document.preview"
	KBCapabilityChunkPreview     KBCapability = "kb.chunk.preview"
	KBCapabilityDocumentDownload KBCapability = "kb.document.download"
	KBCapabilityDocumentUpload   KBCapability = "kb.document.upload"
	KBCapabilityDocumentEdit     KBCapability = "kb.document.edit"
	KBCapabilityDocumentDelete   KBCapability = "kb.document.delete"
	KBCapabilityDocumentReparse  KBCapability = "kb.document.reparse"
	KBCapabilityChunkEdit        KBCapability = "kb.chunk.edit"
	KBCapabilityChunkDelete      KBCapability = "kb.chunk.delete"
	KBCapabilityFolderManage     KBCapability = "kb.folder.manage"
	KBCapabilitySettingsEdit     KBCapability = "kb.settings.edit"
	KBCapabilityDelete           KBCapability = "kb.delete"
	KBCapabilityMembersManage    KBCapability = "kb.members.manage"
	KBCapabilityOwnersManage     KBCapability = "kb.owners.manage"
)

var allKBCapabilities = []KBCapability{
	KBCapabilityAIQuery, KBCapabilityMetadataRead, KBCapabilityDocumentsList,
	KBCapabilityDocumentPreview, KBCapabilityChunkPreview, KBCapabilityDocumentDownload,
	KBCapabilityDocumentUpload, KBCapabilityDocumentEdit, KBCapabilityDocumentDelete,
	KBCapabilityDocumentReparse, KBCapabilityChunkEdit, KBCapabilityChunkDelete,
	KBCapabilityFolderManage, KBCapabilitySettingsEdit, KBCapabilityDelete,
	KBCapabilityMembersManage, KBCapabilityOwnersManage,
}

func AllKBCapabilities() []KBCapability {
	return append([]KBCapability(nil), allKBCapabilities...)
}

func (c KBCapability) IsValid() bool {
	for _, candidate := range allKBCapabilities {
		if c == candidate {
			return true
		}
	}
	return false
}

type KBMemberRole string

const (
	KBMemberRoleOwner          KBMemberRole = "owner"
	KBMemberRoleEditor         KBMemberRole = "editor"
	KBMemberRoleDocumentViewer KBMemberRole = "document_viewer"
	KBMemberRoleAIUser         KBMemberRole = "ai_user"
	KBMemberRoleCustom         KBMemberRole = "custom"
)

func (r KBMemberRole) IsValid() bool {
	switch r {
	case KBMemberRoleOwner, KBMemberRoleEditor, KBMemberRoleDocumentViewer,
		KBMemberRoleAIUser, KBMemberRoleCustom:
		return true
	default:
		return false
	}
}

// KBRoleCapabilities is the frozen VONE-0.7.2.1 preset contract. Custom roles
// use their explicitly persisted capabilities instead of a preset.
func KBRoleCapabilities(role KBMemberRole) []KBCapability {
	var result []KBCapability
	switch role {
	case KBMemberRoleOwner:
		result = AllKBCapabilities()
	case KBMemberRoleEditor:
		result = []KBCapability{
			KBCapabilityAIQuery, KBCapabilityMetadataRead, KBCapabilityDocumentsList,
			KBCapabilityDocumentPreview, KBCapabilityChunkPreview, KBCapabilityDocumentDownload,
			KBCapabilityDocumentUpload, KBCapabilityDocumentEdit, KBCapabilityDocumentDelete,
			KBCapabilityDocumentReparse, KBCapabilityChunkEdit, KBCapabilityChunkDelete,
			KBCapabilityFolderManage,
		}
	case KBMemberRoleDocumentViewer:
		result = []KBCapability{
			KBCapabilityAIQuery, KBCapabilityMetadataRead, KBCapabilityDocumentsList,
			KBCapabilityDocumentPreview, KBCapabilityChunkPreview, KBCapabilityDocumentDownload,
		}
	case KBMemberRoleAIUser:
		// Metadata read allows the KB name and cited source names to be shown;
		// document/chunk list and preview remain deliberately absent.
		result = []KBCapability{KBCapabilityAIQuery, KBCapabilityMetadataRead}
	default:
		return []KBCapability{}
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

type KBMembership struct {
	ID           uint64               `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID     uint64               `json:"tenant_id" gorm:"not null;index"`
	KBID         string               `json:"kb_id" gorm:"column:kb_id;type:varchar(36);not null;index"`
	UserID       string               `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Role         KBMemberRole         `json:"role" gorm:"type:varchar(32);not null"`
	GrantedBy    string               `json:"granted_by" gorm:"type:varchar(36);not null"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
	DeletedAt    gorm.DeletedAt       `json:"-" gorm:"index"`
	Capabilities []KBMemberCapability `json:"capabilities,omitempty" gorm:"foreignKey:MembershipID"`
}

func (KBMembership) TableName() string { return "vone_kb_memberships" }

type KBMemberCapability struct {
	MembershipID uint64       `json:"membership_id" gorm:"primaryKey"`
	Capability   KBCapability `json:"capability" gorm:"type:varchar(64);primaryKey"`
}

func (KBMemberCapability) TableName() string { return "vone_kb_member_capabilities" }

type KBCollection struct {
	ID        string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	TenantID  uint64         `json:"tenant_id" gorm:"not null;index"`
	ParentID  *string        `json:"parent_id,omitempty" gorm:"type:varchar(36);index"`
	Name      string         `json:"name" gorm:"type:varchar(128);not null"`
	SortOrder int            `json:"sort_order" gorm:"not null;default:0"`
	CreatedBy string         `json:"created_by" gorm:"type:varchar(36);not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (KBCollection) TableName() string { return "vone_kb_collections" }

type KBCollectionBinding struct {
	KBID         string    `json:"kb_id" gorm:"column:kb_id;type:varchar(36);primaryKey"`
	TenantID     uint64    `json:"tenant_id" gorm:"not null;index"`
	CollectionID string    `json:"collection_id" gorm:"type:varchar(36);not null;index"`
	UpdatedBy    string    `json:"updated_by" gorm:"type:varchar(36);not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (KBCollectionBinding) TableName() string { return "vone_kb_collection_bindings" }

type KBPolicySource string

const (
	KBPolicySourceWorkspaceOwner KBPolicySource = "workspace_owner"
	KBPolicySourceMembership     KBPolicySource = "membership"
	KBPolicySourceOrganization   KBPolicySource = "organization_share"
	KBPolicySourceAgent          KBPolicySource = "agent_share"
	KBPolicySourceAPIKey         KBPolicySource = "api_key"
	KBPolicySourceNone           KBPolicySource = "none"
)

type KBPolicyDecision struct {
	Allowed      bool           `json:"allowed"`
	Capability   KBCapability   `json:"capability"`
	Source       KBPolicySource `json:"source"`
	Role         KBMemberRole   `json:"role,omitempty"`
	Capabilities []KBCapability `json:"capabilities,omitempty"`
	Reason       string         `json:"reason,omitempty"`
}
