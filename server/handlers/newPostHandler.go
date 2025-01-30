package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"talknet/Database"
	"talknet/server/sessions"
	"talknet/structs"
)

// Allowed file extensions
var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
}

const maxUploadSize = 20 * 1024 * 1024 // Max file size (20MB)

func NewPostHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	userID, isLoggedIn := sessions.GetSessionUserID(r)
	if !isLoggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Fetch the username for the logged-in user
	username, err := Database.GetUsername(db, userID)
	if err != nil {
		log.Printf("Failed to get username: %v", err)
		http.Error(w, "Failed to load user details", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		allCategories, err := Database.GetAllGategories(db)
		if err != nil {
			log.Printf("Failed to get categories: %v", err)
			http.Error(w, "Failed to load categories", http.StatusInternalServerError)
			return
		}

		err = templates.ExecuteTemplate(w, "new-post.html", struct {
			Username      string
			AllCategories []structs.Category
			Post          structs.Post
			PostCategories []structs.Category // ✅ Added to fix the error
		}{
			Username:       username,
			AllCategories:  allCategories,
			Post:           structs.Post{},
			PostCategories: nil, // ✅ Ensuring it's included to prevent template errors
		})
		if err != nil {
			log.Printf("Failed to render template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Handle POST request (image upload & DB insertion)
	err = r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	categories := r.Form["category[]"]

	if title == "" || content == "" || len(categories) == 0 {
		http.Error(w, "Title, content, and category are required", http.StatusBadRequest)
		return
	}

	var imagePath string

	// Handle image upload
	file, fileHeader, err := r.FormFile("image")
	if err == nil { // If an image was uploaded
		defer file.Close()

		// Validate file size
		if fileHeader.Size > maxUploadSize {
			http.Error(w, "File size exceeds 20MB limit", http.StatusBadRequest)
			return
		}

		// Validate file extension
		ext := filepath.Ext(fileHeader.Filename)
		if !allowedExtensions[ext] {
			http.Error(w, "Invalid file type. Only JPEG, PNG, and GIF are allowed.", http.StatusBadRequest)
			return
		}

		// Save the image in the uploads folder
		uploadDir := "./uploads"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.Mkdir(uploadDir, os.ModePerm)
		}

		imagePath = fmt.Sprintf("%s/%d%s", uploadDir, userID, ext)
		outFile, err := os.Create(imagePath)
		if err != nil {
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}
		defer outFile.Close()

		_, err = outFile.ReadFrom(file)
		if err != nil {
			http.Error(w, "Failed to write image", http.StatusInternalServerError)
			return
		}
	}

	// Insert post into database
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}

	res, err := tx.Exec("INSERT INTO Posts (user_id, title, content, image_url) VALUES (?, ?, ?, ?)", userID, title, content, imagePath)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	postID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to retrieve post ID", http.StatusInternalServerError)
		return
	}

	// Associate post with categories
	for _, categoryIDStr := range categories {
		categoryID, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}

		_, err = tx.Exec("INSERT INTO Post_Categories (post_id, category_id) VALUES (?, ?)", postID, categoryID)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to insert post categories", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
