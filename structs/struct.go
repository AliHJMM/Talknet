package structs

import "time"

type ErrorData struct {
	ErrorMessage string
	Code         string
}

// User represents a user in the forum.
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

// Post represents a forum post.
type Post struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Comment represents a comment on a forum post.
type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Like represents a like on a post or comment.
type Like struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	PostID    *int      `json:"post_id,omitempty"`    // Nullable for likes on comments
	CommentID *int      `json:"comment_id,omitempty"` // Nullable for likes on posts
	CreatedAt time.Time `json:"created_at"`
}
type Dislike struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	PostID    *int      `json:"post_id,omitempty"`    // Nullable for likes on comments
	CommentID *int      `json:"comment_id,omitempty"` // Nullable for likes on posts
	CreatedAt time.Time `json:"created_at"`
}


// Category represents a category for posts.
type Category struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// PostCategory represents the association between posts and categories.
type PostCategories struct {
	ID            int    `json:"id"`
	Post_id       int    `json:"post_id"`
	Category_id   int    `json:"category_id"`
	Category_name string `json:"category_name"`
}

type Notification struct {
	CreatedAt   time.Time `json:"created_at"`
	PostTitle   string    `json:"post_title"`
	PostOwner   string    `json:"post_owner"`
	ActionTaker string    `json:"action_taker"`
	Action      string    `json:"action"` // "like", "dislike", "comment"
	CommentText *string   `json:"comment_text,omitempty"`
}