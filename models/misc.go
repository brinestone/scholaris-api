package models

type CaptchaVerifiable interface {
	GetCaptchaToken() string
}

// Represents an entity which can have an ownership relation over other entities.
// Tyically [User], [Institution] and [Tenant] entities.
type OwnerInfo interface {
	GetOwner() uint64
	GetOwnerType() string
}
