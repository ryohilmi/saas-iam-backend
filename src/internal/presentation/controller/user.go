package controller

import (
	"database/sql"
	"iyaem/internal/app/queries"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/lib/pq"
)

type UserController struct {
	db        *sql.DB
	userQuery queries.UserQuery
}

func NewUserController(db *sql.DB, userQuery queries.UserQuery) *UserController {
	return &UserController{db, userQuery}
}

func (c *UserController) UserLevel(ctx *gin.Context) {
	var params struct {
		OrganizationId string `form:"organization_id" binding:"required"`
	}

	if err := ctx.ShouldBindQuery(&params); err != nil {
		log.Printf("Error: %v", err)
		ctx.Error(err)
	}

	token := GetBearerToken(ctx)
	claims, err := DecodeJWT(token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	level, err := c.userQuery.UserLevel(ctx, claims["email"].(string), params.OrganizationId)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get user level",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"level": level,
	})

}

func (c *UserController) UserDetails(ctx *gin.Context) {

	type Params struct {
		UserOrgId string `form:"user_org_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindQuery(&params)

	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err,
		})
		return
	}

	user := struct {
		Id         string         `json:"id"`
		Name       string         `json:"name"`
		Picture    string         `json:"picture"`
		Email      string         `json:"email"`
		Level      string         `json:"level"`
		Identities pq.StringArray `json:"identities"`
	}{}

	row := c.db.QueryRow(`
		select distinct uo.id, u."name", u.picture, u.email, array_remove(array_agg(ui.idp_id), NULL) identities, uo."level"
		from public."user" u 
		left join user_organization uo on uo.user_id = u.id 
		left join user_identity ui on u.id = ui.user_id
		where uo.id=$1
		group by uo.id, u."name", u.picture, u.email;`, params.UserOrgId)
	err = row.Scan(&user.Id, &user.Name, &user.Picture, &user.Email, &user.Identities, &user.Level)

	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "User not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func (c *UserController) UserRoles(ctx *gin.Context) {

	type Params struct {
		UserOrgId string `form:"user_org_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindQuery(&params)

	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err,
		})
		return
	}

	type Role struct {
		TenantName  string         `json:"tenant_name"`
		RoleName    string         `json:"name"`
		RoleId      string         `json:"id"`
		Permissions pq.StringArray `json:"permissions"`
	}

	rows, err := c.db.Query(`
		select distinct a."name", r."name", r.id , array_remove(array_agg(p."name"), NULL) permissions
		from user_role ur
		left join tenant t on ur.tenant_id = t.id
		left join application a on t.app_id = a.id 
		left join "role" r on r.id = ur.role_id 
		left join role_permission rp on rp.role_id = r.id 
		left join "permission" p on rp.permission_id = p.id 
		where ur.user_org_id=$1
		group by a.name, r.name, r.id`, params.UserOrgId)

	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusInternalServerError, "Failed to get user roles")
	}

	var roles []Role = make([]Role, 0)

	for rows.Next() {
		r := Role{}

		err = rows.Scan(&r.TenantName, &r.RoleName, &r.RoleId, &r.Permissions)
		if err != nil {
			log.Printf("Error: %v", err)
			ctx.String(http.StatusInternalServerError, "Failed to get user role")
		}

		roles = append(roles, r)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"roles": roles,
	})
}

func (c *UserController) UserGroups(ctx *gin.Context) {

	type Params struct {
		UserOrgId string `form:"user_org_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindQuery(&params)

	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err,
		})
		return
	}

	type Group struct {
		TenantName string         `json:"tenant_name"`
		GroupName  string         `json:"name"`
		GroupId    string         `json:"id"`
		Roles      pq.StringArray `json:"roles"`
	}

	rows, err := c.db.Query(`
		select distinct a."name", g."name", g.id , array_remove(array_agg(r."name"), NULL) groups
		from user_group ug
		left join tenant t on ug.tenant_id = t.id
		left join application a on t.app_id = a.id 
		left join "group" g  on g.id = ug.group_id 
		left join group_role gr on gr.group_id = g.id 
		left join "role" r on gr.role_id = r.id 
		where ug.user_org_id=$1
		group by a.name, g.name, g.id`, params.UserOrgId)

	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusInternalServerError, "Failed to get user groups")
	}

	var groups []Group = make([]Group, 0)

	for rows.Next() {
		r := Group{}

		err = rows.Scan(&r.TenantName, &r.GroupName, &r.GroupId, &r.Roles)
		if err != nil {
			log.Printf("Error: %v", err)
			ctx.String(http.StatusInternalServerError, "Failed to get user groups")
		}

		groups = append(groups, r)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"groups": groups,
	})
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

func (c *UserController) AssignRole(ctx *gin.Context) {
	type Params struct {
		OrganizationId string `json:"organization_id" binding:"required"`
		UserOrgId      string `json:"user_org_id" binding:"required"`
		TenantId       string `json:"tenant_id" binding:"required"`
		RoleId         string `json:"role_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindBodyWith(&params, binding.JSON)
	if err != nil {
		log.Printf("Error 1301: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Error 1302: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign role",
		})
		return
	}

	// Check if user exists in organization
	row := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM user_organization WHERE id=$1 AND organization_id=$2);", params.UserOrgId, params.OrganizationId)
	var exists bool
	err = row.Scan(&exists)
	if err != nil {
		log.Printf("Error 1303: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign role",
		})
		return
	}

	if !exists {
		log.Printf("Error 1304: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "User does not exist in organization",
		})
		return
	}

	// Insert user role
	_, err = tx.Exec("INSERT INTO user_role (user_org_id, role_id, tenant_id) VALUES ($1, $2, $3);", params.UserOrgId, params.RoleId, params.TenantId)
	if err != nil {
		log.Printf("Error 1305: %v", err)

		if err.Error() == "pq: duplicate key value violates unique constraint \"user_role_user_org_id_idx\"" {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "User already has this role",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign role",
		})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error 1306: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign role",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User assigned role",
	})
}

func (c *UserController) RemoveRole(ctx *gin.Context) {
	type Params struct {
		OrganizationId string `json:"organization_id" binding:"required"`
		UserOrgId      string `json:"user_org_id" binding:"required"`
		TenantId       string `json:"tenant_id" binding:"required"`
		RoleId         string `json:"role_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindBodyWith(&params, binding.JSON)
	if err != nil {
		log.Printf("Error 1301: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Error 1302: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove role",
		})
		return
	}

	// Check if user exists in organization
	row := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM user_organization WHERE id=$1 AND organization_id=$2);", params.UserOrgId, params.OrganizationId)
	var exists bool
	err = row.Scan(&exists)
	if err != nil {
		log.Printf("Error 1303: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove role",
		})
		return
	}

	if !exists {
		log.Printf("Error 1304: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "User does not exist in organization",
		})
		return
	}

	// Insert user role
	_, err = tx.Exec("DELETE FROM user_role WHERE user_org_id=$1 AND role_id=$2 AND tenant_id=$3", params.UserOrgId, params.RoleId, params.TenantId)
	if err != nil {
		log.Printf("Error 1305: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove role",
		})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error 1306: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove role",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Role removed from the user",
	})
}

func (c *UserController) AssignGroup(ctx *gin.Context) {
	type Params struct {
		OrganizationId string `json:"organization_id" binding:"required"`
		UserOrgId      string `json:"user_org_id" binding:"required"`
		TenantId       string `json:"tenant_id" binding:"required"`
		GroupId        string `json:"group_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindBodyWith(&params, binding.JSON)
	if err != nil {
		log.Printf("Error 1301: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Error 1302: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign group",
		})
		return
	}

	// Check if user exists in organization
	row := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM user_organization WHERE id=$1 AND organization_id=$2);", params.UserOrgId, params.OrganizationId)
	var exists bool
	err = row.Scan(&exists)
	if err != nil {
		log.Printf("Error 1303: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign group",
		})
		return
	}

	if !exists {
		log.Printf("Error 1304: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "User does not exist in organization",
		})
		return
	}

	// Insert user group
	_, err = tx.Exec("INSERT INTO user_group (user_org_id, group_id, tenant_id) VALUES ($1, $2, $3);", params.UserOrgId, params.GroupId, params.TenantId)
	if err != nil {
		log.Printf("Error 1305: %v", err)

		if err.Error() == "pq: duplicate key value violates unique constraint \"user_group_user_org_id_idx\"" {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "User already has this group",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign group",
		})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error 1306: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign group",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User assigned group",
	})
}

func (c *UserController) RemoveGroup(ctx *gin.Context) {
	type Params struct {
		OrganizationId string `json:"organization_id" binding:"required"`
		UserOrgId      string `json:"user_org_id" binding:"required"`
		TenantId       string `json:"tenant_id" binding:"required"`
		GroupId        string `json:"group_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindBodyWith(&params, binding.JSON)
	if err != nil {
		log.Printf("Error 1301: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Error 1302: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove group",
		})
		return
	}

	// Check if user exists in organization
	row := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM user_organization WHERE id=$1 AND organization_id=$2);", params.UserOrgId, params.OrganizationId)
	var exists bool
	err = row.Scan(&exists)
	if err != nil {
		log.Printf("Error 1303: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove group",
		})
		return
	}

	if !exists {
		log.Printf("Error 1304: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "User does not exist in organization",
		})
		return
	}

	// Insert user role
	_, err = tx.Exec("DELETE FROM user_group WHERE user_org_id=$1 AND group_id=$2 AND tenant_id=$3", params.UserOrgId, params.GroupId, params.TenantId)
	if err != nil {
		log.Printf("Error 1305: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove group",
		})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error 1306: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove group",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Group removed from the user",
	})
}

func (c *UserController) Promote(ctx *gin.Context) {
	type Params struct {
		OrganizationId string `json:"organization_id" binding:"required"`
		UserOrgId      string `json:"user_org_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindBodyWith(&params, binding.JSON)
	if err != nil {
		log.Printf("Error 1301: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Error 1302: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign role",
		})
		return
	}

	// Check if user exists in organization
	row := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM user_organization WHERE id=$1 AND organization_id=$2);", params.UserOrgId, params.OrganizationId)
	var exists bool
	err = row.Scan(&exists)
	if err != nil {
		log.Printf("Error 1303: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to promote user",
		})
		return
	}

	if !exists {
		log.Printf("Error 1304: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "User does not exist in organization",
		})
		return
	}

	// Update user level
	_, err = tx.Exec("UPDATE user_organization SET level='manager' WHERE id=$1;", params.UserOrgId)
	if err != nil {
		log.Printf("Error 1305: %v", err)

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to promote user",
		})
		return
	}

	// Increment the member count
	_, err = tx.Exec("UPDATE organization SET manager_count = manager_count + 1 WHERE id=$1;", params.OrganizationId)
	if err != nil {
		log.Printf("Error 3131: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to add user to organization",
		})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error 1306: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to promote user",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User assigned role",
	})
}

func (c *UserController) Demote(ctx *gin.Context) {
	type Params struct {
		OrganizationId string `json:"organization_id" binding:"required"`
		UserOrgId      string `json:"user_org_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindBodyWith(&params, binding.JSON)
	if err != nil {
		log.Printf("Error 1301: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Error 1302: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign role",
		})
		return
	}

	// Check if user exists in organization
	row := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM user_organization WHERE id=$1 AND organization_id=$2);", params.UserOrgId, params.OrganizationId)
	var exists bool
	err = row.Scan(&exists)
	if err != nil {
		log.Printf("Error 1303: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to promote user",
		})
		return
	}

	if !exists {
		log.Printf("Error 1304: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "User does not exist in organization",
		})
		return
	}

	// Update user level
	_, err = tx.Exec("UPDATE user_organization SET level='member' WHERE id=$1;", params.UserOrgId)
	if err != nil {
		log.Printf("Error 1305: %v", err)

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to promote user",
		})
		return
	}

	// Increment the member count
	_, err = tx.Exec("UPDATE organization SET manager_count = manager_count - 1 WHERE id=$1;", params.OrganizationId)
	if err != nil {
		log.Printf("Error 3131: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to add user to organization",
		})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error 1306: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to promote user",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User assigned role",
	})
}
