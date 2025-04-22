package model

// RegisterRequest ...
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email" binding:"required,email"`
}

// RegisterResponse ...
type RegisterResponse struct {
	UserID      string `json:"userId"`
	AccessToken string `json:"accessToken"`
}
