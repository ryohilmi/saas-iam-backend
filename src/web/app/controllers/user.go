package controllers

import (
	"database/sql"

	"iyaem/platform/authenticator"
)

type UserController struct {
	auth *authenticator.Authenticator
	db   *sql.DB
}

func NewUserController(auth *authenticator.Authenticator, db *sql.DB) *UserController {
	return &UserController{auth, db}
}
