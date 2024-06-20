package events

import (
	"encoding/json"
	"time"
)

type RoleAddedToMember struct {
	MembershipId string    `json:"membership_id"`
	RoleId       string    `json:"role_id"`
	TenantId     string    `json:"tenant_id"`
	Timestamp    time.Time `json:"timestamp"`
}

func NewRoleAddedToMember(membershipId, roleId, tenantId string) RoleAddedToMember {
	return RoleAddedToMember{MembershipId: membershipId, RoleId: roleId, TenantId: tenantId}
}

func (k RoleAddedToMember) OccuredOn() time.Time {
	return k.Timestamp
}

func (k RoleAddedToMember) JSON() ([]byte, error) {
	return json.Marshal(k)
}
