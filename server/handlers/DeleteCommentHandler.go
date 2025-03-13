package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"talknet/Database"
	"talknet/server/sessions"
)

func DeleteCommentHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// User must be logged in
	userID, isLoggedIn := sessions.GetSessionUserID(r)
	if !isLoggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if r.Method != http.MethodPost {
		RenderErrorPage(w, "Invalid Request Method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		RenderErrorPage(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	commentIDStr := r.FormValue("comment_id")
	postIDStr := r.FormValue("post_id")

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		RenderErrorPage(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		RenderErrorPage(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Verify that the comment belongs to the logged in user (or check admin privileges)
	comments, err := Database.GetCommentsByPostID(db, postID)
	if err != nil {
		RenderErrorPage(w, "Failed to retrieve comment", http.StatusInternalServerError)
		return
	}

	var isOwner bool
	for _, c := range comments {
		if c.ID == commentID && c.UserID == userID {
			isOwner = true
			break
		}
	}
	if !isOwner {
		RenderErrorPage(w, "You are not authorized to delete this comment", http.StatusForbidden)
		return
	}

	err = Database.DeleteComment(db, commentID)
	if err != nil {
		RenderErrorPage(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/post-details?post_id="+postIDStr, http.StatusSeeOther)
}
