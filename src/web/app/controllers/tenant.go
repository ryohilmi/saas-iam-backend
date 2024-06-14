package controllers

import (
	"database/sql"
	"log"
	"net/http"

	"iyaem/platform/authenticator"

	"github.com/gin-gonic/gin"
)

type TenantController struct {
	auth *authenticator.Authenticator
	db   *sql.DB
}

func NewTenantController(auth *authenticator.Authenticator, db *sql.DB) *TenantController {
	return &TenantController{auth, db}
}

func (c *TenantController) TenantList(ctx *gin.Context) {
	type Params struct {
		OrganizationId string `form:"organization_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindQuery(&params)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Organization ID is required",
		})
		return
	}

	rows, err := c.db.Query(`
		SELECT t.id, a."name" FROM tenant t 
		LEFT JOIN application a on t.app_id = a.id
		WHERE t.org_id = $1`, params.OrganizationId)

	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusInternalServerError, "Failed to get tenants")
		return
	}

	type Tenant struct {
		TenantId string `json:"tenant_id"`
		Name     string `json:"name"`
	}
	var users []Tenant = make([]Tenant, 0)

	for rows.Next() {
		var t Tenant

		err = rows.Scan(&t.TenantId, &t.Name)
		if err != nil {
			log.Printf("Error: %v", err)
			ctx.String(http.StatusInternalServerError, "Failed to get tenants")
			return
		}

		users = append(users, t)
	}

	ctx.JSON(http.StatusOK, users)
}
