package Database

import (
	"database/sql"
	"talknet/structs"
	"talknet/utils"
	"time"
)

// GetNotifications returns notifications for the given user (post owner)
// by unioning likes, dislikes, and comments on their posts.
func GetNotifications(db *sql.DB, userID int) ([]structs.Notification, error) {
    query := `
    SELECT created_at, post_title, post_owner, action_taker, action, comment_text
    FROM (
      SELECT ld.created_at,
             p.title as post_title,
             u.username as post_owner,
             u2.username as action_taker,
             'like' as action,
             NULL as comment_text
      FROM Likes_Dislikes ld
      JOIN Posts p ON ld.post_id = p.id
      JOIN Users u ON p.user_id = u.id
      JOIN Users u2 ON ld.user_id = u2.id
      WHERE ld.like_dislike = 1 AND p.user_id = ? AND ld.user_id <> ?
      
      UNION ALL
      
      SELECT ld.created_at,
             p.title as post_title,
             u.username as post_owner,
             u2.username as action_taker,
             'dislike' as action,
             NULL as comment_text
      FROM Likes_Dislikes ld
      JOIN Posts p ON ld.post_id = p.id
      JOIN Users u ON p.user_id = u.id
      JOIN Users u2 ON ld.user_id = u2.id
      WHERE ld.like_dislike = 0 AND p.user_id = ? AND ld.user_id <> ?
      
      UNION ALL
      
      SELECT c.created_at,
             p.title as post_title,
             u.username as post_owner,
             u2.username as action_taker,
             'comment' as action,
             c.content as comment_text
      FROM Comments c
      JOIN Posts p ON c.post_id = p.id
      JOIN Users u ON p.user_id = u.id
      JOIN Users u2 ON c.user_id = u2.id
      WHERE p.user_id = ? AND c.user_id <> ?
    ) AS notifications
    ORDER BY created_at DESC;
    `
    // We pass userID six times:
    rows, err := db.Query(query, userID, userID, userID, userID, userID, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var notifications []structs.Notification

    for rows.Next() {
        var (
            dbCreatedAt time.Time
            postTitle   string
            postOwner   string
            actionTaker string
            action      string
            commentText *string
        )

        err := rows.Scan(&dbCreatedAt, &postTitle, &postOwner, &actionTaker, &action, &commentText)
        if err != nil {
            return nil, err
        }

        // Format the time as you like. Example: "YYYY-MM-DD HH:MM:SS"
		formattedTime := utils.TimeAgo(dbCreatedAt)


        notifications = append(notifications, structs.Notification{
            CreatedAt:   formattedTime,
            PostTitle:   postTitle,
            PostOwner:   postOwner,
            ActionTaker: actionTaker,
            Action:      action,
            CommentText: commentText, // pointer is okay if it might be NULL
        })
    }

    if err = rows.Err(); err != nil {
        return nil, err
    }

    return notifications, nil
}
