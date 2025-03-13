package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"talknet/Database"
	"talknet/server/sessions"
	"talknet/structs"
	"talknet/utils"
)

var postDetailTemplate = template.Must(template.ParseFiles("static/pages/post-details.html"))

type CommentWithUser struct {
	structs.Comment
	Username     string
	CreatedAt    string
	LikeCount    int
	DislikeCount int
	CommentCount int
	Reaction     int
	IsOwner      bool // Newly added field
}

func PostDetailsHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Get the post ID from the URL query
	userSessionID, isLoggedIn := sessions.GetSessionUserID(r)

	postIDStr := r.URL.Query().Get("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		RenderErrorPage(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Fetch the post by ID
	post, err := Database.GetPostByID(db, postID)
	if err != nil {
		RenderErrorPage(w, "Post not found", http.StatusNotFound)
		return
	}

	// Fetch the user who created the post
	user, err := Database.GetUserByID(db, post.UserID)
	if err != nil {
		RenderErrorPage(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Fetch comments for the post
	comments, err := Database.GetCommentsByPostID(db, postID)
	if err != nil {
		log.Printf("Failed to get comments: %v", err)
		RenderErrorPage(w, "Failed to load comments", http.StatusInternalServerError)
		return
	}

	// Prepare comments with usernames and ownership info
	var commentsWithUser []CommentWithUser
	for _, comment := range comments {
		commentUser, err := Database.GetUserByID(db, comment.UserID)
		if err != nil {
			log.Printf("Failed to get user for comment: %v", err)
			continue
		}

		likes, dislikes, err := Database.GetReactionsByCommentID(db, comment.ID)
		if err != nil {
			log.Printf("Failed to get reactions: %v", err)
			continue
		}
		likeCount := len(likes)
		dislikeCount := len(dislikes)

		reaction := -1
		if isLoggedIn {
			reaction, err = Database.CheckReactionExists(db, comment.ID, userSessionID, "comment")
			if err != nil {
				log.Printf("Failed to check reaction: %v", err)
				continue
			}
		}

		// Determine if the current user is the owner of the comment
		isOwner := false
		if isLoggedIn && comment.UserID == userSessionID {
			isOwner = true
		}

		commentsWithUser = append(commentsWithUser, CommentWithUser{
			Comment:      comment,
			Username:     commentUser.Username,
			CreatedAt:    utils.TimeAgo(comment.CreatedAt),
			LikeCount:    likeCount,
			DislikeCount: dislikeCount,
			Reaction:     reaction,
			IsOwner:      isOwner,
		})
	}

	// Render the post details template,
	// passing CurrentUserID so the template can check ownership
	err = postDetailTemplate.Execute(w, struct {
		Post          structs.Post
		Username      string
		Comments      []CommentWithUser
		CurrentUserID int
	}{
		Post:          post, // Now contains ImageURL
		Username:      user.Username,
		Comments:      commentsWithUser,
		CurrentUserID: userSessionID,
	})
	if err != nil {
		log.Printf("Failed to render template: %v", err)
		RenderErrorPage(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
