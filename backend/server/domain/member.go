package domain

import "github.com/carsonalh/churchmanagerbackend/server/util"

type Member struct {
	id           uint64
	firstName    *string
	lastName     *string
	emailAddress *string
	phoneNumber  *string
	notes        string
}

func (member *Member) ToResponseDTO() *MemberResponseDTO {
	return &MemberResponseDTO{
		Id:           member.id,
		FirstName:    member.firstName,
		LastName:     member.lastName,
		EmailAddress: member.emailAddress,
		PhoneNumber:  member.phoneNumber,
		Notes:        member.notes,
	}
}

func (member *Member) Id() uint64 {
	return member.id
}

func (member *Member) FirstName() *string {
	if member.firstName == nil {
		return nil
	}

	return util.NewPtr(*member.firstName)
}

func (member *Member) LastName() *string {
	if member.lastName == nil {
		return nil
	}

	return util.NewPtr(*member.lastName)
}

func (member *Member) EmailAddress() *string {
	if member.emailAddress == nil {
		return nil
	}

	return util.NewPtr(*member.emailAddress)
}

func (member *Member) PhoneNumber() *string {
	if member.phoneNumber == nil {
		return nil
	}

	return util.NewPtr(*member.phoneNumber)
}

func (member *Member) Notes() string {
	return member.notes
}

type MemberRow struct {
	Id           uint64
	FirstName    *string
	LastName     *string
	EmailAddress *string
	PhoneNumber  *string
	Notes        string
}

func (row *MemberRow) ToMember() (*Member, error) {
	member := &Member{
		id:           row.Id,
		firstName:    row.FirstName,
		lastName:     row.LastName,
		emailAddress: row.EmailAddress,
		phoneNumber:  row.PhoneNumber,
		notes:        row.Notes,
	}

	return member, nil
}
