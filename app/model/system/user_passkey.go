package system

import (
	"time"

	"gorm.io/gorm"
)

// UserPasskey represents a WebAuthn credential bound to an admin user.
type UserPasskey struct {
	gorm.Model

	UserID              uint       `gorm:"column:user_id" json:"user_id"`
	CredentialID        string     `gorm:"column:credential_id" json:"credential_id"`
	CredentialPublicKey string     `gorm:"column:credential_public_key" json:"-"`
	SignCount           uint32     `gorm:"column:sign_count" json:"sign_count"`
	AAGUID              string     `gorm:"column:aaguid" json:"aaguid"`
	TransportsJSON      string     `gorm:"column:transports_json" json:"-"`
	UserHandle          string     `gorm:"column:user_handle" json:"-"`
	DisplayName         string     `gorm:"column:display_name" json:"display_name"`
	LastUsedAt          *time.Time `gorm:"column:last_used_at" json:"last_used_at,omitempty"`
}

func (u *UserPasskey) TableName() string {
	return "sys_user_passkey"
}
