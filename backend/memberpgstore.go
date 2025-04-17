package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MemberPostgresStore struct {
	pool *pgxpool.Pool
}

func CreateMemberPgStore(conn *pgxpool.Pool) *MemberPostgresStore {
	return &MemberPostgresStore{conn}
}

// Ignores member's Id field
func (store *MemberPostgresStore) Insert(member *Member) (*Member, error) {
	var result Member
	err := store.pool.QueryRow(
		context.Background(),
		"INSERT INTO member (first_name, last_name, email_address, phone_number, notes)\n"+
			"VALUES ($1, $2, $3, $4, COALESCE($5, ''))\n"+
			"RETURNING id, first_name, last_name, email_address, phone_number, notes;",
		member.FirstName, member.LastName, member.EmailAddress, member.PhoneNumber, member.Notes).
		Scan(&result.Id, &result.FirstName, &result.LastName, &result.EmailAddress, &result.PhoneNumber, &result.Notes)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Ignores member's Id field and uses the "id" parameter to id the member to be
// updated
func (store *MemberPostgresStore) Update(id uint64, member *Member) error {
	rows, err := store.pool.Query(
		context.Background(),
		"UPDATE member SET first_name = $1, last_name = $2, email_address = $3, phone_number = $4, notes = $5 WHERE id = $6;",
		member.FirstName, member.LastName, member.EmailAddress, member.PhoneNumber, member.Notes,
		id,
	)
	if err != nil {
		return err
	}
	rows.Close()
	return nil
}

func (store *MemberPostgresStore) FindById(id uint64) (*Member, error) {
	var result Member
	err := store.pool.QueryRow(
		context.Background(),
		"SELECT id, first_name, last_name, email_address, phone_number, notes FROM member WHERE id = $1;",
		id,
	).Scan(
		&result.Id,
		&result.FirstName,
		&result.LastName,
		&result.EmailAddress,
		&result.PhoneNumber,
		&result.Notes,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &result, nil
}

func (store *MemberPostgresStore) GetPage(pageSize uint, page uint) ([]Member, error) {
	rows, err := store.pool.Query(
		context.Background(),
		"SELECT id, first_name, last_name, email_address, phone_number, notes FROM member ORDER BY id OFFSET $1 LIMIT $2;",
		page*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	members := make([]Member, 0)
	var member Member
	i := 0
	for rows.Next() {
		err = rows.Scan(&member.Id, &member.FirstName, &member.LastName, &member.EmailAddress, &member.PhoneNumber, &member.Notes)
		if err != nil {
			return nil, fmt.Errorf("scanning row %d: %v", i, err)
		}
		members = append(members, member)
		i += 1
	}
	return members, nil
}

func (store *MemberPostgresStore) Delete(id uint64) (bool, error) {
	rows, err := store.pool.Query(context.Background(), "DELETE FROM member WHERE id = $1;", id)
	if err != nil {
		return false, err
	}
	rows.Close()
	deleted := rows.CommandTag().RowsAffected()

	if deleted == 0 {
		return false, nil
	} else if deleted == 1 {
		return true, nil
	} else {
		return false, fmt.Errorf("expected up to one row of table 'member' to be deleted but %d were deleted", deleted)
	}
}
