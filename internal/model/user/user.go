package user

type User struct {
	Username string `json:"username" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}
