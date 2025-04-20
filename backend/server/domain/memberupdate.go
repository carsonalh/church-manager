package domain

type MemberUpdateDTO struct {
	FirstName    *string `json:"firstName" example:"Augustinus"`
	LastName     *string `json:"lastName" example:"Hipponensis"`
	EmailAddress *string `json:"emailAddress" validate:"email" example:"aug.of.hippo@live.roma"`
	PhoneNumber  *string `json:"phoneNumber" example:"0434579344"`
	Notes        string  `json:"notes" example:"Fluent in Latin and Greek."`
} // @name MemberUpdate

func (dto *MemberUpdateDTO) Validate() []error {
	return []error{}
}
