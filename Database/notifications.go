package Database

import (
	"database/sql"
	"talknet/structs"
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
		var n structs.Notification
		err := rows.Scan(&n.CreatedAt, &n.PostTitle, &n.PostOwner, &n.ActionTaker, &n.Action, &n.CommentText)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

