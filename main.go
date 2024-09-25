package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
)

// Global variable to hold the database connection
var db *sql.DB

// Template cache for better performance
var templates *template.Template

// In-memory session store (for simplicity)
var sessions = map[string]string{} // session_id -> user_id

// User structure
type User struct {
	ID       int
	Email    string
	Username string
	Password string
}

// Post structure
type Post struct {
	ID      int
	Title   string
	Content string
	Author  string
}

// Comment structure
type Comment struct {
	ID      int
	Content string
	Author  string
}

// Initialize the database connection and create tables
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Fatal(err)
	}

	// Ping to test the connection
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Create necessary tables
	createTables()
	log.Println("Database connected and tables created successfully")
}

// Create database tables
func createTables() {
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT UNIQUE NOT NULL,
        username TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        title TEXT NOT NULL,
        content TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(user_id) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        post_id INTEGER,
        user_id INTEGER,
        content TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(post_id) REFERENCES posts(id),
        FOREIGN KEY(user_id) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS likes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        post_id INTEGER,
        user_id INTEGER,
        is_like INTEGER,
        FOREIGN KEY(post_id) REFERENCES posts(id),
        FOREIGN KEY(user_id) REFERENCES users(id)
    );
    `
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("Error creating tables: %s", err)
	}
}

// Helper function to render templates
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
	}
}

// Middleware to check if user is logged in
func getUserIDFromSession(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return "", false
	}
	userID, exists := sessions[cookie.Value]
	return userID, exists
}

// Session handling
func createSession(w http.ResponseWriter, userID string) {
	sessionID := uuid.NewString()
	sessions[sessionID] = userID
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})
}

func destroySession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return
	}
	delete(sessions, cookie.Value)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
	})
}

// --- Handlers ---

// Home handler
func homeHandler(w http.ResponseWriter, r *http.Request) {
	userID, isLoggedIn := getUserIDFromSession(r)
	posts, err := getAllPosts()
	if err != nil {
		http.Error(w, "Unable to load posts", http.StatusInternalServerError)
		return
	}
	data := struct {
		Title      string
		IsLoggedIn bool
		UserID     string
		Posts      []Post
	}{
		Title:      "Home",
		IsLoggedIn: isLoggedIn,
		UserID:     userID,
		Posts:      posts,
	}
	renderTemplate(w, "home", data)
}

// Register handler
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, "register", nil)
	} else if r.Method == http.MethodPost {
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Unable to register", http.StatusInternalServerError)
			return
		}
		err = createUser(email, username, string(hashedPassword))
		if err != nil {
			http.Error(w, "Unable to register", http.StatusBadRequest)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

// Login handler
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, "login", nil)
	} else if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		user, err := getUserByEmail(email)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
			http.Error(w, "Invalid login", http.StatusUnauthorized)
			return
		}

		createSession(w, user.ID)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// Create post handler
func createPostHandler(w http.ResponseWriter, r *http.Request) {
	userID, loggedIn := getUserIDFromSession(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		renderTemplate(w, "create_post", nil)
	} else if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")
		err := createPost(userID, title, content)
		if err != nil {
			http.Error(w, "Unable to create post", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// View post handler
func viewPostHandler(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("id")
	post, err := getPostByID(postID)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	comments, err := getCommentsByPostID(postID)
	if err != nil {
		http.Error(w, "Unable to load comments", http.StatusInternalServerError)
		return
	}
	data := struct {
		Post     Post
		Comments []Comment
	}{
		Post:     post,
		Comments: comments,
	}
	renderTemplate(w, "view_post", data)
}

// --- Main Function ---

func main() {
	// Initialize the database
	initDB()

	// Load templates
	templates = template.Must(template.ParseGlob("templates/*.html"))

	// Routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		destroySession(w, r)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	http.HandleFunc("/createPost", createPostHandler)
	http.HandleFunc("/viewPost", viewPostHandler)

	// Serve static files (CSS, images)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start server
	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}

