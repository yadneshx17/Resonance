package spotify

// `json:<key>` these tags maps YOUR field name to THEIR key


type UserProfile struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Product     string `json:"product"`
	Country     string `json:"country"`
	Images      []struct {
		URL string `json:"url"`
	} `json:"images"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}
