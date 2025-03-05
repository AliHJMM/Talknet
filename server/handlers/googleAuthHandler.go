package handlers

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"talknet/Database"
	"talknet/server/sessions"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Hardcoded credentials for development purposes.
var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/auth/google/callback",
	ClientID:     "632142280433-5o1ms2d4djujj0gklfvgooq2ltndmb7e.apps.googleusercontent.com",
	ClientSecret: "YOUR_GOOGLE_CLIENT_SECRET",
	Scopes: []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	},
	Endpoint: google.Endpoint,
}

func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Use a fixed state for simplicity (not recommended for production)
	url := googleOauthConfig.AuthCodeURL("randomstate", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallbackHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != "randomstate" {
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Printf("Token exchange error: %v", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		log.Printf("Error fetching user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading user info response: %v", err)
		http.Error(w, "Failed to read user info", http.StatusInternalServerError)
		return
	}

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
	}
	if err := json.Unmarshal(body, &googleUser); err != nil {
		log.Printf("Error parsing user info: %v", err)
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	user, err := Database.GetUserByEmail(db, googleUser.Email)
	if err != nil {
		username := googleUser.GivenName
		if username == "" {
			username = "googleuser"
		}
		if err := Database.CreateUser(db, username, googleUser.Email, ""); err != nil {
			log.Printf("Error creating user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		user, err = Database.GetUserByEmail(db, googleUser.Email)
		if err != nil {
			log.Printf("Error retrieving new user: %v", err)
			http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
			return
		}
	}

	sessions.CreateSession(w, user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
