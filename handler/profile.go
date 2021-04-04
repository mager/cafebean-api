package handler

type Profile struct {
	Username string `json:"username"`
}

type ProfilePayload struct {
	User Profile `json:"user"`
}
