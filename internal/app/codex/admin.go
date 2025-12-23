package codex

type Code int

const (
	Success        Code = 0
	InvalidParams  Code = 10
	InternalError  Code = 99
	TooManyRequest Code = 100
	Unauthorized   Code = 401
	PermDenied     Code = 403
)

const (
	UserNotFound  Code = 101
	CodeExpired   Code = 102
	WrongCode     Code = 103
	WrongPassword Code = 104
	UserFrozen    Code = 105
	UserBanned    Code = 106
)
