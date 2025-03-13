package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"talknet/Database"
	"talknet/server/sessions"
	"talknet/structs"
)

type EditPostData struct {
    IsLoggedIn     bool
    Post           structs.Post
    AllCategories  []structs.Category
    PostCategories []structs.Category
}

func EditPostHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
    // Ensure the user is logged in
    userID, loggedIn := sessions.GetSessionUserID(r)
    if !loggedIn {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    switch r.Method {
    case http.MethodGet:
        // 1. Parse the post_id from query
        postIDStr := r.URL.Query().Get("post_id")
        postID, err := strconv.Atoi(postIDStr)
        if err != nil {
            log.Printf("Error converting post_id (%s): %v", postIDStr, err)
            RenderErrorPage(w, "Invalid post ID", http.StatusBadRequest)
            return
        }

        // 2. Fetch the post
        post, err := Database.GetPostByID(db, postID)
        if err != nil {
            log.Printf("Error fetching post (ID %d): %v", postID, err)
            RenderErrorPage(w, "Post not found", http.StatusNotFound)
            return
        }

        // 3. Only allow the post owner to edit
        if post.UserID != userID {
            log.Printf("Unauthorized edit attempt: user %d trying to edit post %d owned by %d", userID, postID, post.UserID)
            RenderErrorPage(w, "You are not authorized to edit this post", http.StatusForbidden)
            return
        }

        // 4. Fetch all categories for the category list
        allCategories, err := Database.GetAllGategories(db)
        if err != nil {
            log.Printf("Failed to fetch all categories: %v", err)
            RenderErrorPage(w, "Failed to load categories", http.StatusInternalServerError)
            return
        }

        // 5. Fetch categories currently associated with this post
        postCategories, err := Database.GetCategoryNamesByPostID(db, postID)
        if err != nil {
            log.Printf("Failed to fetch categories for post %d: %v", postID, err)
            RenderErrorPage(w, "Failed to load post categories", http.StatusInternalServerError)
            return
        }

        // 6. Build data for template
        data := EditPostData{
            IsLoggedIn:     loggedIn,
            Post:           post,
            AllCategories:  allCategories,
            PostCategories: postCategories,
        }

        // 7. Render the edit-post template
        log.Printf("Rendering edit template for post: %+v", post)
        err = templates.ExecuteTemplate(w, "edit-post.html", data)
        if err != nil {
            log.Printf("Failed to render edit post template: %v", err)
            RenderErrorPage(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }

    case http.MethodPost:
        // 1. Parse multipart form (for image upload)
        err := r.ParseMultipartForm(20 << 20) // 20MB
        if err != nil {
            log.Printf("Error parsing form data: %v", err)
            RenderErrorPage(w, "Failed to parse form data", http.StatusBadRequest)
            return
        }

        // 2. Extract form fields
        postIDStr := r.FormValue("post_id")
        oldImage := r.FormValue("old_image") // The existing image path
        title := r.FormValue("title")
        content := r.FormValue("content")
        categoryIDs := r.Form["category[]"] // All selected categories

        postID, err := strconv.Atoi(postIDStr)
        if err != nil {
            log.Printf("Error converting post_id from form (%s): %v", postIDStr, err)
            RenderErrorPage(w, "Invalid post ID", http.StatusBadRequest)
            return
        }

        // Basic validation
        if title == "" || content == "" {
            log.Printf("Title or content empty: title=%q, content=%q", title, content)
            RenderErrorPage(w, "Title and content cannot be empty", http.StatusBadRequest)
            return
        }

        // 3. Verify ownership
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

        // 4. Handle image upload (if any)
        newImagePath := oldImage // default to the old image if none is uploaded
        file, fileHeader, fileErr := r.FormFile("image")
        if fileErr == nil {
            // A new image was uploaded
            defer file.Close()

            // Validate extension
            ext := filepath.Ext(fileHeader.Filename)
            if !allowedExtensions[ext] {
                RenderErrorPage(w, "Invalid file type. Only JPEG, PNG, and GIF are allowed.", http.StatusBadRequest)
                return
            }

            // Build unique file name
            newImagePath = generateUniqueFileName(userID, ext)

            // Ensure the upload directory exists
            if _, statErr := os.Stat(uploadDir); os.IsNotExist(statErr) {
                if mkdirErr := os.Mkdir(uploadDir, os.ModePerm); mkdirErr != nil {
                    RenderErrorPage(w, "Failed to create upload directory", http.StatusInternalServerError)
                    return
                }
            }

            // Save the file
            outFile, createErr := os.Create(newImagePath)
            if createErr != nil {
                RenderErrorPage(w, "Failed to save image", http.StatusInternalServerError)
                return
            }
            defer outFile.Close()

            _, copyErr := io.Copy(outFile, file)
            if copyErr != nil {
                RenderErrorPage(w, "Failed to write image", http.StatusInternalServerError)
                return
            }

            // (Optional) remove the old file if it existed
            if oldImage != "" {
                os.Remove(oldImage)
            }
        }

        // 5. Update the Posts table
        _, err = db.Exec(`
            UPDATE Posts
            SET title = ?, content = ?, image_url = ?
            WHERE id = ?
        `, title, content, newImagePath, postID)
        if err != nil {
            log.Printf("Error updating post %d: %v", postID, err)
            RenderErrorPage(w, "Failed to update post", http.StatusInternalServerError)
            return
        }

        // 6. Update categories: remove old, insert new
        _, err = db.Exec("DELETE FROM Post_Categories WHERE post_id = ?", postID)
        if err != nil {
            log.Printf("Failed to clear old categories for post %d: %v", postID, err)
            RenderErrorPage(w, "Failed to update categories", http.StatusInternalServerError)
            return
        }

        for _, catIDStr := range categoryIDs {
            catID, convErr := strconv.Atoi(catIDStr)
            if convErr != nil {
                RenderErrorPage(w, "Invalid category ID", http.StatusBadRequest)
                return
            }
            _, insErr := db.Exec(`
                INSERT INTO Post_Categories (post_id, category_id)
                VALUES (?, ?)
            `, postID, catID)
            if insErr != nil {
                log.Printf("Failed to insert category %d for post %d: %v", catID, postID, insErr)
                RenderErrorPage(w, "Failed to update categories", http.StatusInternalServerError)
                return
            }
        }

        log.Printf("Post %d updated successfully (title, content, categories, image)", postID)
        // Redirect to the post details page
        http.Redirect(w, r, fmt.Sprintf("/post-details?post_id=%d", postID), http.StatusSeeOther)

    default:
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
