package handler

import "time"

type User struct {
	Photo     string    `firestore:"photo" json:"photo"`
	Username  string    `firestore:"username" json:"username"`
	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	Location  string    `firestore:"location" json:"location"`
}

type UserDB struct {
	Email     string    `firestore:"email" json:"email"`
	Photo     string    `firestore:"photo" json:"photo"`
	Username  string    `firestore:"username" json:"username"`
	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	Location  string    `firestore:"location" json:"location"`
}

type PrivateUser struct {
	User UserDB `json:"user"`
}

type UserResp struct {
	User User `json:"user"`
}
