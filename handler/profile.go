package handler

type Profile struct {
	Username string `json:"username"`
	Location string `json:"location"`
}

type ProfilePayload struct {
	User Profile `json:"user"`
}
