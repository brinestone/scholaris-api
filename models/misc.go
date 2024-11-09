package models

type CaptchaVerifiable interface {
	GetCaptchaToken() string
}

type OwnerInfo interface {
	GetOwner() uint64
	GetOwnerType() string
}
