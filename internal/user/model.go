package user

type User struct {
	Email string `validate:"required,email"`
	Pass  string `validate:"required,min=8,max=60"`
}
