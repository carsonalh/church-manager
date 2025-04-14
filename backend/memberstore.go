package main

type MemberStore interface {
	// If the id field is populated, then this is an update, else it is a create
	Save(*Member) error
	// Get a user by id
	GetById(id uint64) (*Member, error)
	// Application clients should re-download their whole member database each
	// time, system designed for up approx. 10,000 members per church
	Get(pageSize uint, page uint) ([]Member, error)
	// Soft deletes by id
	Delete(id uint64) (bool, error)
}
