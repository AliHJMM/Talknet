package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"talknet/Database"
	"talknet/server/sessions"
)

// CommentData holds the data for a comment including post info.
type CommentData struct {
    ID        int
    Content   string
    CreatedAt string
    PostTitle string
    PostID    int
    Username  string
}

func ProfileHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
    var myPostDataList []PostData
    var likedPostDataList []PostData
    var dislikedPostDataList []PostData
    var commentDataList []CommentData

    if r.URL.Path != "/profile" {
        RenderErrorPage(w, "Not Found", http.StatusNotFound)
        return
    }
    if r.Method != "GET" {
        RenderErrorPage(w, "Not Found", http.StatusNotFound)
        return
    }

    // Get the currently logged-in user
    userID, isLoggedIn := sessions.GetSessionUserID(r)

    // If no "id" param is provided, show the logged-in user's profile.
    // Otherwise, interpret "id" as a user ID (not a post ID).
    var profileID int
    if r.URL.Query().Get("id") == "" {
        // No ID in query, so use the logged-in user's ID
        profileID = userID
    } else {
        // Try to parse the query param as an integer user ID
        requestedID, err := strconv.Atoi(r.URL.Query().Get("id"))
        if err != nil {
            log.Printf("Failed to parse profile user ID: %v", err)
            RenderErrorPage(w, "Invalid user ID", http.StatusBadRequest)
            return
        }
        profileID = requestedID
    }

    // Fetch the username for the profile we want to show
    username, err := Database.GetUsername(db, profileID)
    if err != nil {
        log.Printf("Failed to get username for profile ID %d: %v", profileID, err)
        RenderErrorPage(w, "User not found", http.StatusNotFound)
        return
    }

    // Check if this profile belongs to the logged-in user
    isHisProfile := (profileID == userID)

    // 1. Fetch posts created by the user
    posts, err := Database.GetPostByUserID(db, profileID)
    if err != nil {
        log.Printf("Failed to get posts: %v", err)
        RenderErrorPage(w, "Failed to load posts", http.StatusInternalServerError)
        return
    }

    // 2. Fetch posts liked by the user
    likedPosts, err := Database.GetLikedPosts(db, profileID)
    if err != nil {
        log.Printf("Failed to get liked posts: %v", err)
        RenderErrorPage(w, "Failed to load liked posts", http.StatusInternalServerError)
        return
    }

    // 3. Fetch posts disliked by the user
    dislikedPosts, err := Database.GetDislikedPosts(db, profileID)
    if err != nil {
        log.Printf("Failed to get disliked posts: %v", err)
        RenderErrorPage(w, "Failed to load disliked posts", http.StatusInternalServerError)
        return
    }

    // 4. Fetch comments made by the user
    userComments, err := Database.GetCommentsByUserID(db, profileID)
    if err != nil {
        log.Printf("Failed to get comments: %v", err)
        RenderErrorPage(w, "Failed to load comments", http.StatusInternalServerError)
        return
    }

    // Process each category (My Posts, Liked, Disliked, Comments)

    // My Posts
    for _, post := range posts {
        user, err := Database.GetUserByID(db, post.UserID)
        if err != nil {
            log.Printf("Failed to get user: %v", err)
            continue
        }
        postCategories, err := Database.GetCategoryNamesByPostID(db, post.ID)
        if err != nil {
            log.Printf("Failed to get categories: %v", err)
            continue
        }
        likes, dislikes, err := Database.GetReactionsByPostID(db, post.ID)
        if err != nil {
            log.Printf("Failed to get reactions: %v", err)
            continue
        }
        comments, err := Database.GetCommentsByPostID(db, post.ID)
        if err != nil {
            log.Printf("Failed to get comments: %v", err)
            continue
        }
        reaction := -1
        if isLoggedIn {
            reaction, err = Database.CheckReactionExists(db, post.ID, userID, "post")
            if err != nil {
                log.Printf("Failed to check reaction: %v", err)
                continue
            }
        }
        myPostDataList = append(myPostDataList, PostData{
            ID:             post.ID,
            Username:       user.Username,
            Title:          post.Title,
            Content:        post.Content,
            ImageURL:       post.ImageURL,
            CreatedAt:      timeAgo(post.CreatedAt),
            PostCategories: postCategories,
            LikeCount:      len(likes),
            DislikeCount:   len(dislikes),
            CommentCount:   len(comments),
            Reaction:       reaction,
        })
    }

    // Liked Posts
    for _, post := range likedPosts {
        user, err := Database.GetUserByID(db, post.UserID)
        if err != nil {
            log.Printf("Failed to get user: %v", err)
            continue
        }
        postCategories, err := Database.GetCategoryNamesByPostID(db, post.ID)
        if err != nil {
            log.Printf("Failed to get categories: %v", err)
            continue
        }
        likes, dislikes, err := Database.GetReactionsByPostID(db, post.ID)
        if err != nil {
            log.Printf("Failed to get reactions: %v", err)
            continue
        }
        comments, err := Database.GetCommentsByPostID(db, post.ID)
        if err != nil {
            log.Printf("Failed to get comments: %v", err)
            continue
        }
        reaction := -1
        if isLoggedIn {
            reaction, err = Database.CheckReactionExists(db, post.ID, userID, "post")
            if err != nil {
                log.Printf("Failed to check reaction: %v", err)
                continue
            }
        }
        likedPostDataList = append(likedPostDataList, PostData{
            ID:             post.ID,
            Username:       user.Username,
            Title:          post.Title,
            Content:        post.Content,
            ImageURL:       post.ImageURL,
            CreatedAt:      timeAgo(post.CreatedAt),
            PostCategories: postCategories,
            LikeCount:      len(likes),
            DislikeCount:   len(dislikes),
            CommentCount:   len(comments),
            Reaction:       reaction,
        })
    }

    // Disliked Posts
    for _, post := range dislikedPosts {
        user, err := Database.GetUserByID(db, post.UserID)
        if err != nil {
            log.Printf("Failed to get user: %v", err)
            continue
        }
        postCategories, err := Database.GetCategoryNamesByPostID(db, post.ID)
        if err != nil {
            log.Printf("Failed to get categories: %v", err)
            continue
        }
        likes, dislikes, err := Database.GetReactionsByPostID(db, post.ID)
        if err != nil {
            log.Printf("Failed to get reactions: %v", err)
            continue
        }
        comments, err := Database.GetCommentsByPostID(db, post.ID)
        if err != nil {
            log.Printf("Failed to get comments: %v", err)
            continue
        }
        reaction := -1
        if isLoggedIn {
            reaction, err = Database.CheckReactionExists(db, post.ID, userID, "post")
            if err != nil {
                log.Printf("Failed to check reaction: %v", err)
                continue
            }
        }
        dislikedPostDataList = append(dislikedPostDataList, PostData{
            ID:             post.ID,
            Username:       user.Username,
            Title:          post.Title,
            Content:        post.Content,
            ImageURL:       post.ImageURL,
            CreatedAt:      timeAgo(post.CreatedAt),
            PostCategories: postCategories,
            LikeCount:      len(likes),
            DislikeCount:   len(dislikes),
            CommentCount:   len(comments),
            Reaction:       reaction,
        })
    }

    // Comments
    for _, comment := range userComments {
        // Get the post info
        post, err := Database.GetPostByID(db, comment.PostID)
        if err != nil {
            log.Printf("Failed to get post for comment %d: %v", comment.ID, err)
            continue
        }
        // Get the commenter’s user info
        commenter, err := Database.GetUserByID(db, comment.UserID)
        if err != nil {
            log.Printf("Failed to get user for comment %d: %v", comment.ID, err)
            continue
        }
        commentDataList = append(commentDataList, CommentData{
            ID:        comment.ID,
            Content:   comment.Content,
            CreatedAt: timeAgo(comment.CreatedAt),
            PostTitle: post.Title,
            PostID:    post.ID,
            Username:  commenter.Username,
        })
    }

    // Combine all data for the template
    data := struct {
        MyPosts       []PostData
        LikedPosts    []PostData
        DislikedPosts []PostData
        Comments      []CommentData
        IsHisProfile  bool
        Username      string
    }{
        MyPosts:       myPostDataList,
        LikedPosts:    likedPostDataList,
        DislikedPosts: dislikedPostDataList,
        Comments:      commentDataList,
        IsHisProfile:  isHisProfile,
        Username:      username,
    }

    // Render the template
    err = templates.ExecuteTemplate(w, "Profile.html", data)
    if err != nil {
        log.Printf("Failed to render template: %v", err)
        RenderErrorPage(w, "Internal Server Error", http.StatusInternalServerError)
    }
}
