package domain

type MemberUpdateDTO struct {
	FirstName    *string `json:"firstName"`
	LastName     *string `json:"lastName"`
	EmailAddress *string `json:"emailAddress"`
	PhoneNumber  *string `json:"phoneNumber"`
	Notes        string  `json:"notes"`
}

func (dto *MemberUpdateDTO) Validate() []error {
	return []error{}
}
