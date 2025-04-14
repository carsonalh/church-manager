package main

import (
	"github.com/jackc/pgx/v5"
)

type MemberPgStore struct {
	conn *pgx.Conn
}

type MemberPgStoreErrorKind int

const (
	PreconditionFailedError MemberPgStoreErrorKind = iota + 1
	DatabaseError
)

func CreateMemberPgStore(conn *pgx.Conn) *MemberPgStore {
	return &MemberPgStore{conn}
}

// Ignores member's Id field
func (store *MemberPgStore) Insert(member *Member) error {
	return nil
}

// Ignores member's Id field and uses the "id" parameter to id the member to be
// updated
func (store *MemberPgStore) Update(id uint64, member *Member) error {
	return nil
}

func (store *MemberPgStore) FindById(id uint64) (*Member, error) {
	return nil, nil
}

func (store *MemberPgStore) Get(pageSize uint, page uint) ([]Member, error) {
	return make([]Member, 0), nil
}

func (store *MemberPgStore) Delete(id uint64) (bool, error) {
	return false, nil
}
