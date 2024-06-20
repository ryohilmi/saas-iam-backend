package entities

import (
	vo "iyaem/internal/domain/valueobjects"
)

type User struct {
	id          vo.UserId
	name        string
	email       string
	picture     string
	identities  []vo.Identity
	memberships []Membership
}

func NewUser(
	id vo.UserId,
	name,
	email string,
	picture string,
	identities []vo.Identity,
	memberships []Membership,
) User {
	return User{id, name, email, picture, identities, memberships}
}

func (u *User) Id() vo.UserId {
	return u.id
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Email() string {
	return u.email
}

func (u *User) Picture() string {
	return u.picture
}

func (u *User) Memberships() []Membership {
	return u.memberships
}

func (u *User) Identities() []vo.Identity {
	return u.identities
}

func (u *User) AddMembership(m Membership) {
	u.memberships = append(u.memberships, m)
}
