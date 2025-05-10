package models

type AccountMembersRole string

const (
	AccountMembersRoleOwner  AccountMembersRole = "owner"
	AccountMembersRoleAdmin  AccountMembersRole = "admin"
	AccountMembersRoleMember AccountMembersRole = "member"
)

type AccountMember struct {
	AccountID int64
	UserID    int64
	Name      string
	Email     string
	Role      AccountMembersRole
}

type Account struct {
	ID   int64
	Name string
}
