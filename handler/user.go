package handler

type User struct {
	Photo    string `json:"photo"`
	Username string `json:"username"`
}

type UserDB struct {
	Email    string `firestore:"email" json:"email"`
	Photo    string `firestore:"photo" json:"photo"`
	Username string `firestore:"username" json:"username"`
}

type PrivateUser struct {
	User UserDB `json:"user"`
}

type UserResp struct {
	User User `json:"user"`
}
