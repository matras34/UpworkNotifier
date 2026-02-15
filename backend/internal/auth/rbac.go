package auth

import (
	"errors"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

type Permission string

const (
	PermissionReadConnections   Permission = "read:connections"
	PermissionWriteConnections  Permission = "write:connections"
	PermissionDeleteConnections Permission = "delete:connections"
	PermissionCreateSessions    Permission = "create:sessions"
	PermissionReadSessions      Permission = "read:sessions"
	PermissionDeleteSessions    Permission = "delete:sessions"
	PermissionReadUsers         Permission = "read:users"
	PermissionWriteUsers        Permission = "write:users"
	PermissionReadAuditLogs     Permission = "read:audit_logs"
	PermissionManageOrg         Permission = "manage:organization"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleUser   Role = "user"
	RoleViewer Role = "viewer"
)

var rolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermissionReadConnections,
		PermissionWriteConnections,
		PermissionDeleteConnections,
		PermissionCreateSessions,
		PermissionReadSessions,
		PermissionDeleteSessions,
		PermissionReadUsers,
		PermissionWriteUsers,
		PermissionReadAuditLogs,
		PermissionManageOrg,
	},
	RoleUser: {
		PermissionReadConnections,
		PermissionWriteConnections,
		PermissionDeleteConnections,
		PermissionCreateSessions,
		PermissionReadSessions,
		PermissionDeleteSessions,
	},
	RoleViewer: {
		PermissionReadConnections,
		PermissionReadSessions,
	},
}

func HasPermission(role string, permission Permission) bool {
	r := Role(role)
	permissions, ok := rolePermissions[r]
	if !ok {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

func RequirePermission(role string, permission Permission) error {
	if !HasPermission(role, permission) {
		return ErrForbidden
	}
	return nil
}

func IsAdmin(role string) bool {
	return role == string(RoleAdmin)
}
