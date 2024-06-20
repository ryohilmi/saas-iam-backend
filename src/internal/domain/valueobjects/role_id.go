package valueobjects

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type RoleId struct {
	id string
}

func NewRoleId(id string) (RoleId, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return RoleId{}, errors.New("invalid_Role_id")
	}

	return RoleId{id}, nil
}

func (d RoleId) Value() string {
	return d.id
}

func (d RoleId) Equals(other RoleId) bool {
	return strings.EqualFold(d.id, other.id)
}
