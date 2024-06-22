package events

import (
	"encoding/json"
	"time"
)

type GroupRemovedFromMember struct {
	MembershipId string    `json:"membership_id"`
	GroupId      string    `json:"group_id"`
	TenantId     string    `json:"tenant_id"`
	Timestamp    time.Time `json:"timestamp"`
}

func NewGroupRemovedFromMember(membershipId, groupId, tenantId string) GroupRemovedFromMember {
	return GroupRemovedFromMember{MembershipId: membershipId, GroupId: groupId, TenantId: tenantId}
}

func (k GroupRemovedFromMember) OccuredOn() time.Time {
	return k.Timestamp
}

func (k GroupRemovedFromMember) JSON() ([]byte, error) {
	return json.Marshal(k)
}
