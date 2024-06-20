package entities

import (
	"fmt"
	vo "iyaem/internal/domain/valueobjects"
)

type Application struct {
	id          vo.ApplicationId
	name        string
	permissions []Permission
}

func NewApplication(id vo.ApplicationId, name string, permissions []Permission) Application {
	return Application{id, name, permissions}
}

func (u Application) String() string {
	return fmt.Sprint(u.id.Value(), " ", u.name, "\n", "Permissions: ", u.permissions, "\n")
}

func (u *Application) Id() vo.ApplicationId {
	return u.id
}

func (u *Application) Name() string {
	return u.name
}

func (u *Application) Permissions() []Permission {
	return u.permissions
}

func (u *Application) AddPermission(p Permission) {
	u.permissions = append(u.permissions, p)
}
