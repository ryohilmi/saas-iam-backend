package valueobjects

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type MembershipId struct {
	id string
}

func NewMembershipId(id string) (UserId, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return UserId{}, errors.New("invalid_user_id")
	}

	return UserId{id}, nil
}

func GenerateMembershipId() OrganizationId {
	return OrganizationId{uuid.NewString()}
}

func (d MembershipId) Value() string {
	return d.id
}

func (d MembershipId) Equals(other UserId) bool {
	return strings.EqualFold(d.id, other.id)
}
