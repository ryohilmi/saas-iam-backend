package events

import (
	"encoding/json"
	"time"
)

type MemberRemoved struct {
	MembershipId string    `json:"membership_id"`
	Timestamp    time.Time `json:"timestamp"`
}

func NewMemberRemoved(membershipId string) MemberRemoved {
	return MemberRemoved{MembershipId: membershipId}
}

func (k MemberRemoved) OccuredOn() time.Time {
	return k.Timestamp
}

func (k MemberRemoved) JSON() ([]byte, error) {
	return json.Marshal(k)
}
