package events

import (
	"encoding/json"
	"time"
)

type RoleRemovedFromMember struct {
	MembershipId string    `json:"membership_id"`
	RoleId       string    `json:"role_id"`
	TenantId     string    `json:"tenant_id"`
	Timestamp    time.Time `json:"timestamp"`
}

func NewRoleRemovedFromMember(membershipId, roleId, tenantId string) RoleRemovedFromMember {
	return RoleRemovedFromMember{MembershipId: membershipId, RoleId: roleId, TenantId: tenantId}
}

func (k RoleRemovedFromMember) OccuredOn() time.Time {
	return k.Timestamp
}

func (k RoleRemovedFromMember) JSON() ([]byte, error) {
	return json.Marshal(k)
}
