package domain

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrInvalidEmail      = errors.New("invalid email address")
	ErrInvalidName       = errors.New("name cannot be empty")
	ErrRoleNotFound      = errors.New("role not found")
	ErrRoleAlreadyExists = errors.New("role with this name already exists")
	ErrInvalidRoleName   = errors.New("role name cannot be empty")
	ErrNavItemNotFound   = errors.New("navigation item not found")
	ErrInvalidNavItemName = errors.New("navigation item name cannot be empty")
	ErrInvalidNavItemURL  = errors.New("navigation item URL cannot be empty")
)
