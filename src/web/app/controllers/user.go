package controllers

import (
	"database/sql"
	"log"
	"net/http"

	"iyaem/platform/authenticator"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	auth *authenticator.Authenticator
	db   *sql.DB
}

func NewUserController(auth *authenticator.Authenticator, db *sql.DB) *UserController {
	return &UserController{auth, db}
}

func (c *UserController) DoesUserExist(ctx *gin.Context) {
	authorizationHeader := ctx.Request.Header.Get("Authorization")

	if authorizationHeader == "" {
		ctx.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	type Params struct {
		Email string `form:"email" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindQuery(&params)

	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Email is required",
		})
		return
	}

	token := authorizationHeader[len("Bearer "):]
	_, err = DecodeJWT(token)
	if err != nil {
		log.Print(err)
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	var userExists bool
	row := c.db.QueryRow(`
		SELECT EXISTS(
		SELECT 1 FROM public.user 
		WHERE email=$1);`, params.Email)

	row.Scan(&userExists)

	if !userExists {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "User does not exist",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User exists",
	})
}
