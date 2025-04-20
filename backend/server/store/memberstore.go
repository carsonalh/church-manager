package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/carsonalh/churchmanagerbackend/server/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MemberStore struct {
	pool *pgxpool.Pool
}

func CreateMemberStore(pool *pgxpool.Pool) *MemberStore {
	return &MemberStore{pool}
}

// Ignores member's Id field
func (store *MemberStore) Create(createDto *domain.MemberUpdateDTO) (*domain.Member, error) {
	var row domain.MemberRow
	err := store.pool.QueryRow(
		context.Background(),
		"INSERT INTO member (first_name, last_name, email_address, phone_number, notes)\n"+
			"VALUES ($1, $2, $3, $4, $5)\n"+
			"RETURNING id, first_name, last_name, email_address, phone_number, notes;",
		createDto.FirstName, createDto.LastName, createDto.EmailAddress, createDto.PhoneNumber, createDto.Notes).
		Scan(&row.Id, &row.FirstName, &row.LastName, &row.EmailAddress, &row.PhoneNumber, &row.Notes)
	if err != nil {
		return nil, err
	}
	member, err := row.ToMember()
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (store *MemberStore) Update(id uint64, updateDto *domain.MemberUpdateDTO) (*domain.Member, error) {
	row := domain.MemberRow{}
	err := store.pool.QueryRow(
		context.Background(),
		"UPDATE member SET first_name = $1, last_name = $2, email_address = $3, phone_number = $4, notes = $5 WHERE id = $6\n"+
			"RETURNING id, first_name, last_name, email_address, phone_number, notes;",
		updateDto.FirstName, updateDto.LastName, updateDto.EmailAddress, updateDto.PhoneNumber, updateDto.Notes,
		id,
	).Scan(
		&row.Id, &row.FirstName, &row.LastName, &row.EmailAddress, &row.PhoneNumber, &row.Notes,
	)
	if err != nil {
		return nil, err
	}
	member, err := row.ToMember()
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (store *MemberStore) FindById(id uint64) (*domain.Member, error) {
	var row domain.MemberRow
	err := store.pool.QueryRow(
		context.Background(),
		"SELECT id, first_name, last_name, email_address, phone_number, notes FROM member WHERE id = $1;",
		id,
	).Scan(
		&row.Id,
		&row.FirstName,
		&row.LastName,
		&row.EmailAddress,
		&row.PhoneNumber,
		&row.Notes,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	member, err := row.ToMember()
	if err != nil {
		return nil, err
	}

	return member, nil
}

func (store *MemberStore) GetPage(pageSize uint, page uint) ([]domain.Member, error) {
	rows, err := store.pool.Query(
		context.Background(),
		"SELECT id, first_name, last_name, email_address, phone_number, notes FROM member ORDER BY id OFFSET $1 LIMIT $2;",
		page*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	members := make([]domain.Member, 0)
	i := 0
	for rows.Next() {
		var row domain.MemberRow
		err = rows.Scan(&row.Id, &row.FirstName, &row.LastName, &row.EmailAddress, &row.PhoneNumber, &row.Notes)
		if err != nil {
			return nil, fmt.Errorf("scanning row %d: %v", i, err)
		}
		member, err := row.ToMember()
		if err != nil {
			return nil, fmt.Errorf("converting row to member at row %d: %v", i, err)
		}
		members = append(members, *member)
		i += 1
	}
	return members, nil
}

func (store *MemberStore) DeleteById(id uint64) (bool, error) {
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
