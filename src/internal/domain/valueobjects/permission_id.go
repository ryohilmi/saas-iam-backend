package valueobjects

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type PermissionId struct {
	id string
}

func NewPermissionId(id string) (PermissionId, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return PermissionId{}, errors.New("invalid_permission_id")
	}

	return PermissionId{id}, nil
}

func (d PermissionId) Value() string {
	return d.id
}

func (d PermissionId) Equals(other PermissionId) bool {
	return strings.EqualFold(d.id, other.id)
}
