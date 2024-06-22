package controller

import (
	"database/sql"
	"iyaem/internal/app/commands"
	"iyaem/internal/app/queries"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/lib/pq"
)

type UserController struct {
	db *sql.DB

	promoteUserCommand *commands.PromoteUserCommand
	demoteUserCommand  *commands.DemoteUserCommand
	addRoleCommand     *commands.AddRoleToMemberCommand
	removeRoleCommand  *commands.RemoveRoleFromMemberCommand
	addGroupCommand    *commands.AddGroupToMemberCommand
	removeGroupCommand *commands.RemoveGroupFromMemberCommand

	userQuery queries.UserQuery
}

func NewUserController(
	db *sql.DB,
	promoteUser *commands.PromoteUserCommand,
	demoteUser *commands.DemoteUserCommand,
	addRoleCommand *commands.AddRoleToMemberCommand,
	removeRoleCommand *commands.RemoveRoleFromMemberCommand,
	addGroupCommand *commands.AddGroupToMemberCommand,
	removeGroupCommand *commands.RemoveGroupFromMemberCommand,
	userQuery queries.UserQuery,
) *UserController {
	return &UserController{
		db,
		promoteUser,
		demoteUser,
		addRoleCommand,
		removeRoleCommand,
		addGroupCommand,
		removeGroupCommand,
		userQuery}
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
	var params struct {
		OrganizationId string `json:"organization_id" binding:"required"`
		UserOrgId      string `json:"user_org_id" binding:"required"`
		TenantId       string `json:"tenant_id" binding:"required"`
		RoleId         string `json:"role_id" binding:"required"`
	}

	err := ctx.ShouldBindBodyWith(&params, binding.JSON)
	if err != nil {
		log.Printf("Error 1301: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	req := commands.AddRoleToMemberRequest{
		OrganizationId: params.OrganizationId,
		MembershipId:   params.UserOrgId,
		TenantId:       params.TenantId,
		RoleId:         params.RoleId,
	}
	membershipId, err := c.addRoleCommand.Execute(ctx, req)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "success",
		"data":    membershipId,
	})
}

func (c *UserController) RemoveRole(ctx *gin.Context) {
	var params struct {
		OrganizationId string `json:"organization_id" binding:"required"`
		UserOrgId      string `json:"user_org_id" binding:"required"`
		TenantId       string `json:"tenant_id" binding:"required"`
		RoleId         string `json:"role_id" binding:"required"`
	}

	err := ctx.ShouldBindBodyWith(&params, binding.JSON)
	if err != nil {
		log.Printf("Error 1301: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	req := commands.RemoveRoleFromMemberRequest{
		OrganizationId: params.OrganizationId,
		MembershipId:   params.UserOrgId,
		TenantId:       params.TenantId,
		RoleId:         params.RoleId,
	}
	membershipId, err := c.removeRoleCommand.Execute(ctx, req)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "success",
		"data":    membershipId,
	})
}

func (c *UserController) AssignGroup(ctx *gin.Context) {
	var params struct {
		OrganizationId string `json:"organization_id" binding:"required"`
		UserOrgId      string `json:"user_org_id" binding:"required"`
		TenantId       string `json:"tenant_id" binding:"required"`
		GroupId        string `json:"group_id" binding:"required"`
	}

	err := ctx.ShouldBindBodyWith(&params, binding.JSON)
	if err != nil {
		log.Printf("Error 1301: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	req := commands.AddGroupToMemberRequest{
		OrganizationId: params.OrganizationId,
		MembershipId:   params.UserOrgId,
		TenantId:       params.TenantId,
		GroupId:        params.GroupId,
	}
	membershipId, err := c.addGroupCommand.Execute(ctx, req)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "success",
		"data":    membershipId,
	})
}

func (c *UserController) RemoveGroup(ctx *gin.Context) {
	var params struct {
		OrganizationId string `json:"organization_id" binding:"required"`
		UserOrgId      string `json:"user_org_id" binding:"required"`
		TenantId       string `json:"tenant_id" binding:"required"`
		GroupId        string `json:"group_id" binding:"required"`
	}

	err := ctx.ShouldBindBodyWith(&params, binding.JSON)
	if err != nil {
		log.Printf("Error 1301: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	req := commands.RemoveGroupFromMemberRequest{
		OrganizationId: params.OrganizationId,
		MembershipId:   params.UserOrgId,
		TenantId:       params.TenantId,
		GroupId:        params.GroupId,
	}
	membershipId, err := c.removeGroupCommand.Execute(ctx, req)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "success",
		"data":    membershipId,
	})
}

func (c *UserController) Promote(ctx *gin.Context) {
	var params struct {
		OrganizationId string `json:"organization_id" binding:"required"`
		UserOrgId      string `json:"user_org_id" binding:"required"`
	}

	if err := ctx.ShouldBindBodyWith(&params, binding.JSON); err != nil {
		log.Printf("Error 1301: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	req := commands.PromoteUserRequest{
		OrganizationId: params.OrganizationId,
		MembershipId:   params.UserOrgId,
	}
	membershipId, err := c.promoteUserCommand.Execute(ctx, req)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "success",
		"data":    membershipId,
	})
}

func (c *UserController) Demote(ctx *gin.Context) {
	var params struct {
		OrganizationId string `json:"organization_id" binding:"required"`
		UserOrgId      string `json:"user_org_id" binding:"required"`
	}

	if err := ctx.ShouldBindBodyWith(&params, binding.JSON); err != nil {
		log.Printf("Error 1301: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	req := commands.DemoteUserRequest{
		OrganizationId: params.OrganizationId,
		MembershipId:   params.UserOrgId,
	}
	membershipId, err := c.demoteUserCommand.Execute(ctx, req)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "success",
		"data":    membershipId,
	})
}
