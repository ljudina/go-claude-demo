package domain

// Policy represents a Casbin policy rule
type Policy struct {
	Role     string
	Resource string
	Action   string
}
