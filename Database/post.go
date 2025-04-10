package Database

import (
	"database/sql"
	"talknet/structs" // Adjust import path as needed
	"time"
)

// CreatePost inserts a new post into the database.
func CreatePost(db *sql.DB, userID int, title, content string) error {
	_, err := db.Exec("INSERT INTO posts (user_id, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		userID, title, content, time.Now(), time.Now())
	return err
}

func GetPostByID(db *sql.DB, id int) (structs.Post, error) {
    row := db.QueryRow("SELECT id, user_id, title, content, image_url, created_at FROM posts WHERE id = ?", id)
    
    var post structs.Post
    var imageURL sql.NullString // Handle NULL values

    err := row.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &imageURL, &post.CreatedAt)
    if err != nil {
        return post, err
    }

    if imageURL.Valid {
        post.ImageURL = imageURL.String
    } else {
        post.ImageURL = "" // Set empty string if NULL
    }

    return post, nil
}




func GetAllPosts(db *sql.DB) ([]structs.Post, error) {
	rows, err := db.Query("SELECT id, user_id, title, content, image_url, created_at FROM posts ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []structs.Post
	for rows.Next() {
		var post structs.Post
		var imageURL sql.NullString // Handle NULL image_url

		err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &imageURL, &post.CreatedAt)
		if err != nil {
			return nil, err
		}

		// ✅ Handle NULL image_url
		if imageURL.Valid {
			post.ImageURL = imageURL.String
		} else {
			post.ImageURL = ""
		}

		posts = append(posts, post)
	}
	return posts, nil
}





// Other post-related functions (e.g., UpdatePost, DeletePost) go here.
func GetPostsByCategory(db *sql.DB, category string) ([]structs.Post, error) {
	rows, err := db.Query(`
        SELECT p.id, p.user_id, p.title, p.content, p.created_at
        FROM Posts p
        JOIN Post_Categories pc ON p.id = pc.post_id
        JOIN Categories c ON pc.category_id = c.id
        WHERE c.name = ?`, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []structs.Post
	for rows.Next() {
		var post structs.Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// DeletePost removes a post from the database.
func DeletePost(db *sql.DB, postID int) error {
	// Delete associated comments
	_, err := db.Exec("DELETE FROM Comments WHERE post_id = ?", postID)
	if err != nil {
		return err
	}

	// Delete likes/dislikes related to the post
	_, err = db.Exec("DELETE FROM Likes_Dislikes WHERE post_id = ?", postID)
	if err != nil {
		return err
	}

	// Delete the post itself
	_, err = db.Exec("DELETE FROM Posts WHERE id = ?", postID)
	return err
}

// DeleteComment removes a comment from the database.
func DeleteComment(db *sql.DB, commentID int) error {
	// Delete likes/dislikes related to the comment
	_, err := db.Exec("DELETE FROM Likes_Dislikes WHERE comment_id = ?", commentID)
	if err != nil {
		return err
	}

	// Delete the comment itself
	_, err = db.Exec("DELETE FROM Comments WHERE id = ?", commentID)
	return err
}


func GetPostByUserID(db *sql.DB, user_id int) ([]structs.Post, error) {
	rows, err := db.Query(`
		SELECT id, user_id, title, content, image_url, created_at 
		FROM posts 
		WHERE user_id = ?`, user_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []structs.Post
	for rows.Next() {
		var post structs.Post
		var imageURL sql.NullString // Handle NULL values safely

		err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &imageURL, &post.CreatedAt)
		if err != nil {
			return nil, err
		}

		// ✅ Assign image URL properly
		if imageURL.Valid {
			post.ImageURL = imageURL.String
		} else {
			post.ImageURL = ""
		}

		posts = append(posts, post)
	}

	return posts, nil
}







func GetLikedPosts(db *sql.DB, user_id int) ([]structs.Post, error) {
	rows, err := db.Query(`
		SELECT Posts.id, Posts.user_id, Posts.title, Posts.content, Posts.image_url, Posts.created_at 
		FROM Posts 
		INNER JOIN Likes_Dislikes ON Posts.id = Likes_Dislikes.post_id 
		WHERE Likes_Dislikes.user_id = ?  
		AND Likes_Dislikes.like_dislike = 1 
		AND Likes_Dislikes.post_id IS NOT NULL;
	`, user_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []structs.Post
	for rows.Next() {
		var post structs.Post
		var imageURL sql.NullString // Handle NULL values safely

		err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &imageURL, &post.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Assign image URL properly
		if imageURL.Valid {
			post.ImageURL = imageURL.String
		} else {
			post.ImageURL = ""
		}

		posts = append(posts, post)
	}

	return posts, nil
}


// UpdatePost updates an existing post's title and content.
func UpdatePost(db *sql.DB, postID int, title, content string) error {
	_, err := db.Exec(`
        UPDATE Posts
        SET title = ?, content = ?, updated_at = ?
        WHERE id = ?
    `, title, content, time.Now(), postID)
	return err
}