package domain

type MemberResponseDTO struct {
	Id           uint64  `json:"id" example:"81996"`
	FirstName    *string `json:"firstName" example:"Augustinus"`
	LastName     *string `json:"lastName" example:"Hipponensis"`
	EmailAddress *string `json:"emailAddress" example:"aug.of.hippo@live.roma"`
	PhoneNumber  *string `json:"phoneNumber" example:"0434579344"`
	Notes        string  `json:"notes" example:"Fluent in Latin and Greek."`
} // @name MemberResponse
