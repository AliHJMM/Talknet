package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"talknet/Database"
	"talknet/server/sessions"
)

// EditPostHandler handles both displaying the edit form (GET)
// and updating the post (POST).
func EditPostHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Ensure the user is logged in
	userID, loggedIn := sessions.GetSessionUserID(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if r.Method == http.MethodGet {
		// Display the edit form.
		postIDStr := r.URL.Query().Get("post_id")
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			log.Printf("Error converting post_id (%s): %v", postIDStr, err)
			RenderErrorPage(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		// Fetch the post.
		post, err := Database.GetPostByID(db, postID)
		if err != nil {
			log.Printf("Error fetching post (ID %d): %v", postID, err)
			RenderErrorPage(w, "Post not found", http.StatusNotFound)
			return
		}

		// Only allow the post owner to edit.
		if post.UserID != userID {
			log.Printf("Unauthorized edit attempt: user %d trying to edit post %d owned by %d", userID, postID, post.UserID)
			RenderErrorPage(w, "You are not authorized to edit this post", http.StatusForbidden)
			return
		}

		// Log the post details for debugging.
		log.Printf("Rendering edit template for post: %+v", post)

		// Render the edit-post template.
		err = templates.ExecuteTemplate(w, "edit-post.html", post)
		if err != nil {
			log.Printf("Failed to render edit post template: %v", err)
			RenderErrorPage(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else if r.Method == http.MethodPost {
		// Process the form submission.
		err := r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form data: %v", err)
			RenderErrorPage(w, "Failed to parse form data", http.StatusBadRequest)
			return
		}

		postIDStr := r.FormValue("post_id")
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			log.Printf("Error converting post_id from form (%s): %v", postIDStr, err)
			RenderErrorPage(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		title := r.FormValue("title")
		content := r.FormValue("content")
		if title == "" || content == "" {
			log.Printf("Title or content empty: title=%q, content=%q", title, content)
			RenderErrorPage(w, "Title and content cannot be empty", http.StatusBadRequest)
			return
		}

		// Check if the logged in user is the owner.
		post, err := Database.GetPostByID(db, postID)
		if err != nil {
			log.Printf("Error fetching post for update (ID %d): %v", postID, err)
			RenderErrorPage(w, "Post not found", http.StatusNotFound)
			return
		}
		if post.UserID != userID {
			log.Printf("Unauthorized update attempt: user %d trying to update post %d owned by %d", userID, postID, post.UserID)
			RenderErrorPage(w, "You are not authorized to edit this post", http.StatusForbidden)
			return
		}

		// Update the post.
		err = Database.UpdatePost(db, postID, title, content)
		if err != nil {
			log.Printf("Error updating post %d: %v", postID, err)
			RenderErrorPage(w, "Failed to update post", http.StatusInternalServerError)
			return
		}

		log.Printf("Post %d updated successfully", postID)
		// Redirect to the post details page.
		http.Redirect(w, r, fmt.Sprintf("/post-details?post_id=%d", postID), http.StatusSeeOther)
	} else {
		RenderErrorPage(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}


// DeletePostHandler handles post deletion.
func DeletePostHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Ensure the user is logged in.
	userID, loggedIn := sessions.GetSessionUserID(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if r.Method != http.MethodPost {
		RenderErrorPage(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		RenderErrorPage(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	postIDStr := r.FormValue("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		RenderErrorPage(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Fetch the post.
	post, err := Database.GetPostByID(db, postID)
	if err != nil {
		RenderErrorPage(w, "Post not found", http.StatusNotFound)
		return
	}

	// Only allow the owner to delete.
	if post.UserID != userID {
		RenderErrorPage(w, "You are not authorized to delete this post", http.StatusForbidden)
		return
	}

	

	// Delete the post.
	err = Database.DeletePost(db, postID)
	if err != nil {
		RenderErrorPage(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	// Redirect to the home page (or profile page).
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
