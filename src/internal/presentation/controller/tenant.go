package controller

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TenantController struct {
	db *sql.DB
}

func NewTenantController(db *sql.DB) *TenantController {
	return &TenantController{db}
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

func (c *TenantController) Groups(ctx *gin.Context) {
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
		SELECT g.id, g."name", g.description, r."name" FROM "group" g 
		LEFT JOIN group_role gr on g.id = gr.group_id 
		LEFT JOIN "role" r on gr.role_id = r.id 
		LEFT JOIN tenant t on g.application_id = t.app_id
		WHERE t.id=$1;`, params.TenantId)

	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusInternalServerError, "Failed to get groups")
		return
	}

	type Group struct {
		Id        string   `json:"id"`
		Name      string   `json:"name"`
		GroupDesc string   `json:"description"`
		Roles     []string `json:"roles"`
	}
	var groups []Group = make([]Group, 0)

	var prevGroup Group
	for rows.Next() {
		var g Group
		var roleName string

		err = rows.Scan(&g.Id, &g.Name, &g.GroupDesc, &roleName)
		if err != nil {
			log.Printf("Error: %v", err)
			ctx.String(http.StatusInternalServerError, "Failed to get groups")
			return
		}

		if prevGroup.Id != g.Id {
			g.Roles = []string{roleName}
			groups = append(groups, g)

			prevGroup = g
		} else {
			groups[len(groups)-1].Roles = append(groups[len(groups)-1].Roles, roleName)
		}
	}

	ctx.JSON(http.StatusOK, groups)
}
