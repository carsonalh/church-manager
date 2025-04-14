package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type MemberPgStore struct {
	conn *pgx.Conn
}

func CreateMemberPgStore(conn *pgx.Conn) *MemberPgStore {
	return &MemberPgStore{conn}
}

// Ignores member's Id field
func (store *MemberPgStore) Insert(member *Member) (*Member, error) {
	rows, err := store.conn.Query(
		context.Background(),
		"INSERT INTO member (first_name, last_name, email_address, phone_number, notes)\n"+
			"VALUES ($1, $2, $3, $4, COALESCE($5, ''))\n"+
			"RETURNING id, first_name, last_name, email_address, phone_number, notes;",
		member.FirstName, member.LastName, member.EmailAddress, member.PhoneNumber, member.Notes)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, fmt.Errorf("query did not return a row when one was expected")
	}
	var result Member
	err = rows.Scan(&result.Id, &result.FirstName, &result.LastName, &result.EmailAddress, &result.PhoneNumber, &result.Notes)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Ignores member's Id field and uses the "id" parameter to id the member to be
// updated
func (store *MemberPgStore) Update(id uint64, member *Member) error {
	_, err := store.conn.Query(
		context.Background(),
		"UPDATE member SET first_name = $1, last_name = $2, email_address = $3, phone_number = $4 WHERE id = $5;",
		member.FirstName, member.LastName, member.EmailAddress, member.PhoneNumber, id)
	return err
}

func (store *MemberPgStore) FindById(id uint64) (*Member, error) {
	rows, err := store.conn.Query(
		context.Background(),
		"SELECT id, first_name, last_name, email_address, phone_number FROM member WHERE id = $1;",
		id)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, fmt.Errorf("query did not return a row when one was expected")
	}

	var result Member
	err = rows.Scan(&result.Id, &result.FirstName, &result.LastName, &result.EmailAddress, &result.PhoneNumber)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (store *MemberPgStore) Get(pageSize uint, page uint) ([]Member, error) {
	rows, err := store.conn.Query(
		context.Background(),
		"SELECT id, first_name, last_name, email_address, phone_number FROM member ORDER BY id OFFSET $1 LIMIT $2;",
		page*pageSize, page)
	if err != nil {
		return nil, err
	}
	members := make([]Member, 0)
	var member Member
	i := 0
	for rows.Next() {
		err = rows.Scan(&member.Id, &member.FirstName, &member.LastName, &member.EmailAddress, &member.PhoneNumber)
		if err != nil {
			return nil, fmt.Errorf("scanning row %d: %v", i, err)
		}
		members = append(members, member)
		i += 1
	}
	return members, nil
}

func (store *MemberPgStore) Delete(id uint64) (bool, error) {
	rows, err := store.conn.Query(context.Background(), "DELETE FROM member WHERE id = $1 RETURNING COUNT(*);", id)
	if err != nil {
		return false, err
	}
	if !rows.Next() {
		return false, fmt.Errorf("expected a row but found none")
	}
	var deleted uint
	err = rows.Scan(&deleted)
	if err != nil {
		return false, err
	}

	if deleted == 0 {
		return false, nil
	} else if deleted == 1 {
		return true, nil
	} else {
		return false, fmt.Errorf("expected up to one row of table 'member' to be deleted but %d were deleted", deleted)
	}
}
