package valueobjects

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type GroupId struct {
	id string
}

func NewGroupId(id string) (GroupId, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return GroupId{}, errors.New("invalid_group_id")
	}

	return GroupId{id}, nil
}

func GenerateGroupId() GroupId {
	return GroupId{uuid.NewString()}
}

func (d GroupId) Value() string {
	return d.id
}

func (d GroupId) Equals(other GroupId) bool {
	return strings.EqualFold(d.id, other.id)
}
