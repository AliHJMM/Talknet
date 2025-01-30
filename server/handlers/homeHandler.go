package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"talknet/Database"
	"talknet/server/sessions"
	"talknet/structs"
	"time"
)

type StaticPageData struct {
	IsLoggedIn    bool
	AllCategories []structs.Category
}

type PostData struct {
	ID             int
	Username       string
	Title          string
	Content        string
	CreatedAt      string
	PostCategories []structs.Category
	ImageURL       string
	LikeCount      int
	DislikeCount   int
	CommentCount   int
	Reaction       int
}

var templates = template.Must(template.ParseGlob("static/pages/*.html"))

func HomeHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		RenderErrorPage(w, "Not Found", http.StatusNotFound)
		return
	}

	userSessionID, isLoggedIn := sessions.GetSessionUserID(r)

	// Fetch categories (this is used in the sidebar or category filters)
	allCategories, err := Database.GetAllGategories(db)
	if err != nil {
		log.Printf("Failed to get all categories: %v", err)
		RenderErrorPage(w, "Failed to load categories", http.StatusInternalServerError)
		return
	}

	// Fetch posts based on selected category
	category := r.URL.Query().Get("category")
	var posts []structs.Post

	if category != "" && category != "All" {
		// If a category is selected, fetch the posts for that category
		posts, err = Database.GetPostsByCategory(db, category)
		if err != nil {
			log.Printf("Failed to get posts by category: %v", err)
			RenderErrorPage(w, "Failed to load posts", http.StatusInternalServerError)
			return
		}
	} else {
		// Otherwise, fetch all posts
		posts, err = Database.GetAllPosts(db)
		if err != nil {
			log.Printf("Failed to get all posts: %v", err)
			RenderErrorPage(w, "Failed to load posts", http.StatusInternalServerError)
			return
		}
	}

	staticData := StaticPageData{
		IsLoggedIn:    isLoggedIn,
		AllCategories: allCategories,
	}

	// Prepare the dynamic post data for rendering
	var postDataList []PostData

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
			log.Printf("Failed to get likes: %v", err)
			continue
		}
		likeCount := len(likes)
		dislikeCount := len(dislikes)

		comments, err := Database.GetCommentsByPostID(db, post.ID)
		if err != nil {
			log.Printf("Failed to get comments: %v", err)
			continue
		}

		reaction := -1
		if isLoggedIn {
			reaction, err = Database.CheckReactionExists(db, post.ID, userSessionID, "post")
			if err != nil {
				log.Printf("Failed to check reaction: %v", err)
				continue
			}
		}

		postDataList = append(postDataList, PostData{
			ID:             post.ID,
			Username:       user.Username,
			Title:          post.Title,
			Content:        post.Content,
			CreatedAt:      timeAgo(post.CreatedAt),
			PostCategories: postCategories,
			ImageURL:       post.ImageURL, // Now setting ImageURL
			LikeCount:      likeCount,
			DislikeCount:   dislikeCount,
			CommentCount:   len(comments),
			Reaction:       reaction,
		})
	}

	// Sort the posts by CreatedAt in descending order
	sort.Slice(postDataList, func(i, j int) bool {
		createdAtI, _ := time.Parse(time.RFC3339, postDataList[i].CreatedAt) // Assuming the date format is RFC3339
		createdAtJ, _ := time.Parse(time.RFC3339, postDataList[j].CreatedAt)
		return createdAtI.After(createdAtJ) // Sort by most recent
	})

	// Render the template with both static (categories, isLoggedIn) and dynamic (posts) data
	err = templates.ExecuteTemplate(w, "index.html", struct {
		StaticData StaticPageData
		Posts      []PostData
	}{
		StaticData: staticData,
		Posts:      postDataList,
	})
	if err != nil {
		log.Printf("Failed to render template: %v", err)
		RenderErrorPage(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// timeAgo function to format time
func timeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%d minute%s ago", minutes, pluralize(minutes))
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%d hour%s ago", hours, pluralize(hours))
	case diff < 30*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d day%s ago", days, pluralize(days))
	default:
		return t.Format("2006-01-02") // Fallback to a specific date format
	}
}

func pluralize(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
