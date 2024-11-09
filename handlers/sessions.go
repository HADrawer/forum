package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	// "strconv"
)
var sessions = map[string]string{} // key is the session ID, value is the user ID

var userSessions = map[string]string{}


func CreateSession(w http.ResponseWriter, userID string) {
    // Generate a new UUID for the session ID
    sessionID := uuid.NewString()

    if existingSessionID, exists := userSessions[userID]; exists {
        delete(sessions, existingSessionID)
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    "",
			Expires:  time.Now().Add(-1 * time.Hour), // Expire the old session cookie immediately
			HttpOnly: true,
		})
    }
    
    // Store the session ID and associated userID in the session store
    sessions[sessionID] = userID
    userSessions[userID] = sessionID

    // Set a cookie with the session ID
    http.SetCookie(w, &http.Cookie{
        Name:     "session_id",
        Value:    sessionID,
        Expires:  time.Now().Add(24 * time.Hour), // Set the expiration to 24 hours
        HttpOnly: true,                           // Make it inaccessible via JavaScript
    })
}

func GetUserIDFromSession(r *http.Request) (string, bool) {
    cookie, err := r.Cookie("session_id")
    if err != nil {
        return "", false
    }
    
    // Check if session ID exists in the session store
    userID, exists := sessions[cookie.Value]
    if !exists {
        return "", false
    }

    return userID, true
}

// after the user log we delete 
func DestroySession(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("session_id")
    if err != nil {
        return
    }

    userID, exists := sessions[cookie.Value]
	if !exists {
		return
	}
    // Delete session from the session store
    delete(sessions, cookie.Value)

    delete(userSessions, userID)
    // Expire the cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "session_id",
        Value:    "",
        Expires:  time.Now().Add(-1 * time.Hour), // Expire the cookie immediately
        HttpOnly: true,
    })
}
