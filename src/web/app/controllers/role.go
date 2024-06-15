package controllers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RoleController struct {
	db *sql.DB
}

func NewRoleController(db *sql.DB) *RoleController {
	return &RoleController{db}
}

func (c *RoleController) UsersWithRole(ctx *gin.Context) {
	type Params struct {
		OrganizationId string `form:"organization_id" binding:"required"`
		TenantId       string `form:"tenant_id" binding:"required"`
		RoleId         string `form:"role_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindQuery(&params)
	if err != nil {
		log.Printf("Error 1401: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	rows, err := c.db.Query(`
		select distinct uo.id, u."picture", u."name", u.email from user_organization uo 
		left join public."user" u on u.id = uo.user_id 
		left join user_role ur on uo.id = ur.user_org_id
		left join tenant t on ur.tenant_id = t.id 
		where t.id=$1 and ur.role_id =$2;`, params.TenantId, params.RoleId)

	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusInternalServerError, "Failed to get users")
		return
	}

	type User struct {
		UserOrgId string `json:"user_org_id"`
		Picture   string `json:"picture"`
		Name      string `json:"name"`
		Email     string `json:"email"`
	}
	var users []User = make([]User, 0)

	for rows.Next() {
		var u User

		err = rows.Scan(&u.UserOrgId, &u.Picture, &u.Name, &u.Email)
		if err != nil {
			log.Printf("Error: %v", err)
			ctx.String(http.StatusInternalServerError, "Failed to get users")
			return
		}

		users = append(users, u)
	}

	ctx.JSON(http.StatusOK, users)
}
