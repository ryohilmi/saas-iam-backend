package events

import (
	"encoding/json"
	"time"
)

type MemberDemoted struct {
	MembershipId string    `json:"membership_id"`
	Timestamp    time.Time `json:"timestamp"`
}

func NewMemberDemoted(membershipId string) MemberDemoted {
	return MemberDemoted{MembershipId: membershipId}
}

func (k MemberDemoted) OccuredOn() time.Time {
	return k.Timestamp
}

func (k MemberDemoted) JSON() ([]byte, error) {
	return json.Marshal(k)
}
