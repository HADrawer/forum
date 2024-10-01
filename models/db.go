

package models

import (
    "database/sql"
    "errors"
	"log"
    _ "modernc.org/sqlite" 
    "fmt"
    	// "strconv"
)


var ErrUserExists = errors.New("user already exists")
var db *sql.DB

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

// Initialize the database connection
func InitDB() {
	var err error
	db, err = sql.Open("sqlite", "./forum.db")
	if err != nil {
		log.Fatal(err)
	}

	// Ping to test the connection
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Create necessary tables
	CreateTables()
	log.Println("Database connected and tables created successfully")
}

// Create database tables
func CreateTables() {
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

// Create user



// CreateUser adds a new user to the database
func CreateUser(user User) error {
    // Prepare the SQL statement
    stmt, err := db.Prepare("INSERT INTO users (email, username, password) VALUES (?, ?, ?)")
    if err != nil {
        log.Printf("Failed to prepare statement: %v", err)
        return fmt.Errorf("failed to prepare statement: %w", err)
    }
    defer stmt.Close()

    // Execute the statement with user data
    _, err = stmt.Exec(user.Email, user.Username, user.Password)
    if err != nil {
        // Check for unique constraint violation
        if err.Error() == "UNIQUE constraint failed: users.email, users.username" {
            return ErrUserExists
        }
        log.Printf("Failed to execute statement: %v", err)
        return fmt.Errorf("failed to execute statement: %w", err)
    }

    return nil
}

// Get user by email
func GetUserByEmail(email string) (*User, error) {
    var user User
    err := db.QueryRow("SELECT id, email, username, password FROM users WHERE email = ?", email).
        Scan(&user.ID, &user.Email, &user.Username, &user.Password)
    if err != nil {
        return nil, errors.New("user not found")
    }
    return &user, nil
}
// func GetUserByID(id string) (*User, error) {
//     var user User
//     intID , _ := strconv.Atoi(id)
//     err := db.QueryRow("SELECT id, email, username, password FROM users WHERE id = ?", intID).
//         Scan(&user.ID, &user.Email, &user.Username, &user.Password)
//     if err != nil {
//         return nil, errors.New("user not found")
//     }
//     return &user ,err
// }

// Get all posts
func GetAllPosts() ([]Post, error) {
    var posts []Post
    rows, err := db.Query("SELECT id, title, content FROM posts")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var post Post
        if err := rows.Scan(&post.ID, &post.Title, &post.Content); err != nil {
            return nil, err
        }
        posts = append(posts, post)
    }
    return posts, nil
}

// Get post by ID
func GetPostByID(postID string) (*Post, error) {
    var post Post
    err := db.QueryRow("SELECT id, title, content FROM posts WHERE id = ?", postID).
        Scan(&post.ID, &post.Title, &post.Content)
    if err != nil {
        return nil, errors.New("post not found")
    }
    return &post, nil
}

// Create post
func CreatePost(userID, title, content string) error {
    stmt, err := db.Prepare("INSERT INTO posts(user_id, title, content) VALUES(?, ?, ?)")
    if err != nil {
        return err
    }
    _, err = stmt.Exec(userID, title, content)
    return err
}

// Get comments by post ID
func GetCommentsByPostID(postID string) ([]Comment, error) {
    var comments []Comment
    rows, err := db.Query("SELECT id, content FROM comments WHERE post_id = ?", postID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var comment Comment
        if err := rows.Scan(&comment.ID, &comment.Content); err != nil {
            return nil, err
        }
        comments = append(comments, comment)
    }
    return comments, nil
}
