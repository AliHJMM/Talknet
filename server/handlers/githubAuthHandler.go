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
	"golang.org/x/oauth2/github"
)

// Hardcoded credentials for GitHub OAuth
var githubOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/auth/github/callback",
	ClientID:     "Ov23liYyZJwKB6A55BjU",     // replace with your GitHub Client ID
	ClientSecret: "32c461c2c985104a9683c5fb9dcd11ff695c6c6d", // replace with your GitHub Client Secret
	Scopes:       []string{"user:email"},
	Endpoint:     github.Endpoint,
}

// GithubLoginHandler redirects the user to GitHub's OAuth consent page.
func GithubLoginHandler(w http.ResponseWriter, r *http.Request) {
	// "state" should be random for security; using a fixed value for simplicity.
	url := githubOauthConfig.AuthCodeURL("randomstate", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GithubCallbackHandler handles the callback from GitHub.
func GithubCallbackHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != "randomstate" {
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	token, err := githubOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Printf("Token exchange error: %v", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Fetch user info from GitHub.
	resp, err := http.Get("https://api.github.com/user?access_token=" + token.AccessToken)
	if err != nil {
		log.Printf("Error fetching GitHub user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading GitHub user info: %v", err)
		http.Error(w, "Failed to read user info", http.StatusInternalServerError)
		return
	}

	// Parse GitHub user data.
	var githubUser struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(body, &githubUser); err != nil {
		log.Printf("Error parsing GitHub user info: %v", err)
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	// GitHub may not return email if it's not public.
	if githubUser.Email == "" {
		req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
		if err != nil {
			log.Printf("Error creating email request: %v", err)
			http.Error(w, "Failed to create email request", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Authorization", "token "+token.AccessToken)
		emailResp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Error fetching emails: %v", err)
			http.Error(w, "Failed to fetch emails", http.StatusInternalServerError)
			return
		}
		defer emailResp.Body.Close()
		emailBody, err := ioutil.ReadAll(emailResp.Body)
		if err != nil {
			log.Printf("Error reading email response: %v", err)
			http.Error(w, "Failed to read email response", http.StatusInternalServerError)
			return
		}
		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}
		if err := json.Unmarshal(emailBody, &emails); err != nil {
			log.Printf("Error parsing emails: %v", err)
			http.Error(w, "Failed to parse email info", http.StatusInternalServerError)
			return
		}
		for _, e := range emails {
			if e.Primary && e.Verified {
				githubUser.Email = e.Email
				break
			}
		}
	}

	// Check if user exists; if not, create a new user.
	user, err := Database.GetUserByEmail(db, githubUser.Email)
	if err != nil {
		username := githubUser.Login
		if username == "" {
			username = "githubuser"
		}
		if err := Database.CreateUser(db, username, githubUser.Email, ""); err != nil {
			log.Printf("Error creating user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		user, err = Database.GetUserByEmail(db, githubUser.Email)
		if err != nil {
			log.Printf("Error retrieving new user: %v", err)
			http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
			return
		}
	}

	// Create a session and redirect to home.
	sessions.CreateSession(w, user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
