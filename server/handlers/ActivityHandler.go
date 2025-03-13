package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"talknet/Database"
	"talknet/server/sessions"
	"talknet/structs"
)

func ActivityHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Ensure user is logged in.
	userID, loggedIn := sessions.GetSessionUserID(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Get notifications for this user.
	notifications, err := Database.GetNotifications(db, userID)
	if err != nil {
		log.Printf("Failed to get notifications: %v", err)
		RenderErrorPage(w, "Failed to load notifications", http.StatusInternalServerError)
		return
	}

	// Parse the activity template.
	tmpl, err := template.ParseFiles("static/pages/activity.html")
	if err != nil {
		log.Printf("Failed to parse activity template: %v", err)
		RenderErrorPage(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render the template with notifications and logged-in flag.
	err = tmpl.Execute(w, struct {
		Notifications []structs.Notification
		IsLoggedIn    bool
	}{
		Notifications: notifications,
		IsLoggedIn:    true,
	})
	if err != nil {
		log.Printf("Failed to render activity template: %v", err)
		RenderErrorPage(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
