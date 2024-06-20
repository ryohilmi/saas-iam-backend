package valueobjects

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type TenantId struct {
	id string
}

func NewTenantId(id string) (TenantId, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return TenantId{}, errors.New("invalid_tenant_id")
	}

	return TenantId{id}, nil
}

func (d TenantId) Value() string {
	return d.id
}

func (d TenantId) Equals(other TenantId) bool {
	return strings.EqualFold(d.id, other.id)
}
