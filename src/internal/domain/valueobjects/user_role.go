package valueobjects

type UserRole struct {
	userId UserId
	roleId RoleId
}

func NewUserRole(userId UserId, roleId RoleId) UserRole {
	return UserRole{userId, roleId}
}

func (ur *UserRole) UserId() UserId {
	return ur.userId
}

func (ur *UserRole) RoleId() RoleId {
	return ur.roleId
}
