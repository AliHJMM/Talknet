package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"talknet/Database"
	"talknet/server/sessions"
)

// EditCommentHandler handles editing an existing comment.
func EditCommentHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
    // User must be logged in
    userID, isLoggedIn := sessions.GetSessionUserID(r)
    if !isLoggedIn {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    // Only handle POST requests
    if r.Method != http.MethodPost {
        RenderErrorPage(w, "Invalid Request Method", http.StatusMethodNotAllowed)
        return
    }

    // Parse form data
    err := r.ParseForm()
    if err != nil {
        RenderErrorPage(w, "Failed to parse form data", http.StatusBadRequest)
        return
    }

    commentIDStr := r.FormValue("comment_id")
    newContent := r.FormValue("content")
    postIDStr := r.FormValue("post_id") // so we can redirect back after edit

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

    // Make sure comment belongs to user (or handle admin roles if needed)
    existingComments, err := Database.GetCommentsByPostID(db, postID)
    if err != nil {
        RenderErrorPage(w, "Failed to retrieve comment", http.StatusInternalServerError)
        return
    }

    // Check ownership
    var isOwner bool
    for _, c := range existingComments {
        if c.ID == commentID && c.UserID == userID {
            isOwner = true
            break
        }
    }
    if !isOwner {
        RenderErrorPage(w, "You are not authorized to edit this comment", http.StatusForbidden)
        return
    }

    // Update the comment content
    if err := Database.EditComment(db, commentID, newContent); err != nil {
        RenderErrorPage(w, "Failed to update comment", http.StatusInternalServerError)
        return
    }

    // Redirect back to the post details page
    http.Redirect(w, r, "/post-details?post_id="+postIDStr, http.StatusSeeOther)
}
