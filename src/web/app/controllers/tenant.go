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

func (c *TenantController) Roles(ctx *gin.Context) {
	type Params struct {
		TenantId string `form:"tenant_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindQuery(&params)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Tenant ID is required",
		})
		return
	}

	rows, err := c.db.Query(`
		SELECT role.id, role.name, role.description, p."name" as permission  FROM role
		LEFT JOIN role_permission rp on role.id = rp.role_id
		LEFT JOIN "permission" p on rp.permission_id = p.id 
		LEFT JOIN tenant t on role.application_id = t.app_id
		WHERE t.id = $1;`, params.TenantId)

	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusInternalServerError, "Failed to get roles")
		return
	}

	type Role struct {
		Id          string   `json:"id"`
		Name        string   `json:"name"`
		RoleDesc    string   `json:"description"`
		Permissions []string `json:"permissions"`
	}
	var roles []Role = make([]Role, 0)

	var prevRole Role
	for rows.Next() {
		var r Role
		var permName string

		err = rows.Scan(&r.Id, &r.Name, &r.RoleDesc, &permName)
		if err != nil {
			log.Printf("Error: %v", err)
			ctx.String(http.StatusInternalServerError, "Failed to get roles")
			return
		}

		if prevRole.Id != r.Id {
			r.Permissions = []string{permName}
			roles = append(roles, r)

			prevRole = r
		} else {
			roles[len(roles)-1].Permissions = append(roles[len(roles)-1].Permissions, permName)
		}
	}

	ctx.JSON(http.StatusOK, roles)
}
