package events

import (
	"encoding/json"
	"time"
)

type GroupAddedToMember struct {
	MembershipId string    `json:"membership_id"`
	GroupId      string    `json:"group_id"`
	TenantId     string    `json:"tenant_id"`
	Timestamp    time.Time `json:"timestamp"`
}

func NewGroupAddedToMember(membershipId, groupId, tenantId string) GroupAddedToMember {
	return GroupAddedToMember{MembershipId: membershipId, GroupId: groupId, TenantId: tenantId}
}

func (k GroupAddedToMember) OccuredOn() time.Time {
	return k.Timestamp
}

func (k GroupAddedToMember) JSON() ([]byte, error) {
	return json.Marshal(k)
}
