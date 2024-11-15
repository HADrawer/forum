package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
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
	ID         int
	UserID     int
	Title      string
	Content    string
	Author     string
	Category   []Category
    Likes      int
    Dislikes   int
	Created_at string
}

// Comment structure
type Comment struct {
	ID         int
	Content    string
	User_ID    string
	Author     string
	Created_at string
}
type Category struct {
	ID   int
	Name string
}
type Like struct {
	ID     int
	PostID string
	UserID string
	IsLike string
}
type CommentLike struct {
	ID     int
	CommentID string
	UserID string
	IsLike string
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
	CreateCategory(db)
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
        Author Text NOT NULL,
        Category TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(user_id) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        post_id INTEGER,
        user_id INTEGER,
        Author Text NOT NULL,
        comment TEXT NOT NULL,
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
	CREATE TABLE IF NOT EXISTS commentlikes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        comment_id INTEGER,
        user_id INTEGER,
        is_like INTEGER,
        FOREIGN KEY(comment_id) REFERENCES comments(id),
        FOREIGN KEY(user_id) REFERENCES users(id)
    );
	
    CREATE TABLE IF NOT EXISTS categories (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT UNIQUE NOT NULL
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
		if err.Error() == "constraint failed: UNIQUE constraint failed: users.username (2067)" {
			return ErrUserExists
		}else if err.Error() == "constraint failed: UNIQUE constraint failed: users.email (2067)"{
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
func GetUserByUserName(username string) (*User, error) {
	var user User
	err := db.QueryRow("SELECT id, email, username, password FROM users WHERE username = ?", username).
		Scan(&user.ID, &user.Email, &user.Username, &user.Password)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

// Get all posts
func GetAllPosts() ([]Post, error) {
	var posts []Post
	rows, err := db.Query("SELECT  id ,user_id, title, content ,Author , created_at   FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		var createdAt time.Time
		if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.Author, &createdAt); err != nil {
			return nil, err
		}
		post.Created_at = createdAt.Format("2006-01-02 15:04:05")
		posts = append(posts, post)
	}
	return posts, nil
}

func GetAllCategoryPosts(isCategory string) ([]Post, error) {
	var posts []Post
	query := "SELECT id, user_id, title, content, Author , created_at FROM posts WHERE Category LIKE ?"
	// Prepare the LIKE pattern to search for the category
	pattern := "%" + isCategory + "%"

	rows, err := db.Query(query, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		var createdAt time.Time
		if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.Author, &createdAt); err != nil {
			return nil, err
		}
		post.Created_at = createdAt.Format("2006-01-02 15:04:05")
		posts = append(posts, post)
	}

	return posts, nil
}

// Get post by ID
func GetPostByID(postID string) (*Post, error) {
	var post Post
	var createdAt time.Time
	err := db.QueryRow("SELECT id ,user_id, title, content ,Author , created_at FROM posts WHERE id = ?", postID).
		Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.Author, &createdAt)
	if err != nil {
		return nil, errors.New("post not found")
	}
	post.Created_at = createdAt.Format("2006-01-02 15:04:05")
	return &post, nil
}

// Create post
func CreatePost(userID, title, content, categories string) error {
	stmt, err := db.Prepare("INSERT INTO posts (user_id, title, content ,Author, Category) VALUES(?, ?, ?,?,?)")
	if err != nil {
		return err
	}
	user, _ := GetUserByUserName(userID)
	_, err = stmt.Exec(user.ID, title, content, user.Username, categories)
	return err
}
func CreateComment(userID, postID, comment string) error {
	stmt, err := db.Prepare("INSERT INTO comments (post_id , user_id, Author , comment) VALUES(?,?,?,?)")
	if err != nil {
		return err
	}
	user, _ := GetUserByUserName(userID)
	_, err = stmt.Exec(postID, user.ID, user.Username, comment)
	return err
}

// Get comments by post ID
func GetCommentsByPostID(postID string) ([]Comment, error) {
	var comments []Comment
	rows, err := db.Query("SELECT id, user_id, Author ,comment , created_at FROM comments WHERE post_id = ?", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment Comment
		var createdAt time.Time
		if err := rows.Scan(&comment.ID, &comment.User_ID, &comment.Author, &comment.Content, &createdAt); err != nil {
			return nil, err
		}
		comment.Created_at = createdAt.Format("2006-01-02 15:04:05")
		comments = append(comments, comment)
	}
	return comments, nil
}

var ErrCategoryExists = errors.New("category already exists")

// CreateCategory adds a new category to the database
func CreateCategory(database *sql.DB) {
	statement, _ := database.Prepare("INSERT INTO categories (name) SELECT ? WHERE NOT EXISTS (SELECT 1 FROM categories WHERE name = ?)")
	statement.Exec("General", "General")
	statement.Exec("Technology", "Technology")
	statement.Exec("Science", "Science")
	statement.Exec("Sports", "Sports")
	statement.Exec("Gaming", "Gaming")
	statement.Exec("Music", "Music")
	statement.Exec("Books", "Books")
	statement.Exec("Movies", "Movies")
	statement.Exec("TV", "TV")
	statement.Exec("Food", "Food")
	statement.Exec("Travel", "Travel")
	statement.Exec("Photography", "Photography")
	statement.Exec("Art", "Art")
	statement.Exec("Writing", "Writing")
	statement.Exec("Programming", "Programming")
	statement.Exec("Other", "Other")
}

// GetAllCategories retrieves all categories from the database
func GetAllCategories() ([]Category, error) {
	rows, err := db.Query("SELECT id, name FROM categories")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}
	return categories, nil
}

// GetPostsFromUserID retrieves posts created by the user with the given userID
func GetPostsFromUserID(userID string) ([]Post, error) {
	var posts []Post
	user , _ := GetUserByUserName(userID)
	rows, err := db.Query("SELECT id, user_id, title, content, Author , created_at FROM posts WHERE user_id = ?", user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		var createdAt time.Time
		if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.Author, &createdAt); err != nil {
			
			return nil, err
		}
		post.Created_at = createdAt.Format("2006-01-02 15:04:05")
		posts = append(posts, post)
	}

	return posts, nil
}



func GetPostsFromLiked(userID string) ([]Post, error) {
	var posts []Post
	user , _ := GetUserByUserName(userID)

	query := `
		SELECT p.id, p.user_id, p.title, p.content, p.Author, p.created_at
		FROM posts p
		JOIN likes l ON p.id = l.post_id
		WHERE l.user_id = ? AND l.is_like = 1
	`

	rows, err := db.Query(query, user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		var createdAt time.Time
		if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.Author, &createdAt); err != nil {
			
			return nil, err
		}
		post.Created_at = createdAt.Format("2006-01-02 15:04:05")
		posts = append(posts, post)
	}

	return posts, nil
}
func LikeCounter(postID string) (int ,error){
	var likes []Like
	rows, err := db.Query("SELECT is_like  FROM likes WHERE post_id = ? AND is_like = 1 ", postID)
if err != nil {
	return 0, errors.New(" not found")
}
defer rows.Close()

	for rows.Next() {
		var like Like
		if err := rows.Scan(&like.IsLike); err != nil {
			return 0, err
		}
		likes = append(likes, like)
	}
return len(likes), nil
}
func DisLikeCounter(postID string) (int ,error){
	var likes []Like
	rows, err := db.Query("SELECT is_like  FROM likes WHERE post_id = ? AND is_like = -1 ", postID)
	if err != nil {
		return 0, errors.New(" not found")
	}
	defer rows.Close()
	
		for rows.Next() {
			var like Like
			if err := rows.Scan(&like.IsLike); err != nil {
				return 0, err
			}
			likes = append(likes, like)
		}
	return len(likes), nil
}

func AddLike(postID, userID, Liked string) {
	stat, _ := db.Prepare("INSERT INTO likes (post_id,user_id, is_like) VALUES (?,?,?)")
	user, _ := GetUserByUserName(userID)
	stat.Exec(postID, user.ID, Liked)
}
func RemoveLike(postID, userID string) {
	stat, _ := db.Prepare("DELETE FROM likes WHERE post_id = ? AND user_id = ?")
	user, _ := GetUserByUserName(userID)
	stat.Exec(postID, user.ID)
}
func UpdateLike(postID, userID, Liked string) {
	statement, _ := db.Prepare("UPDATE likes SET is_like = ? WHERE post_id = ? AND user_id = ?")
	user, _ := GetUserByUserName(userID)
	statement.Exec(Liked, postID, user.ID)
}

func IsLike(postID, userID string) bool {
	user, _ := GetUserByUserName(userID)

	rows, _ := db.Query("SELECT is_like FROM likes WHERE user_id = ? AND post_id = ? AND is_like = 1", user.ID, postID)
	like := 0
	for rows.Next() {
		rows.Scan(&like)
	}
	if like != 0 {
		return true
	}
	return false
}
func IsDisLike(postID, userID string) bool {
	user, _ := GetUserByUserName(userID)

	rows, _ := db.Query("SELECT is_like FROM likes WHERE user_id = ? AND post_id = ? AND is_like = -1", user.ID, postID)
	like := 0
	for rows.Next() {
		rows.Scan(&like)
	}
	if like != 0 {
		return true
	}
	return false
}
func CommentLikeCounter(CommentID string) (int ,error){
	var likes []CommentLike
	rows, err := db.Query("SELECT is_like  FROM Commentlikes WHERE comment_id = ? AND is_like = 1 ", CommentID)
if err != nil {
	return 0, errors.New(" not found")
}
defer rows.Close()

	for rows.Next() {
		var like CommentLike
		if err := rows.Scan(&like.IsLike); err != nil {
			return 0, err
		}
		likes = append(likes, like)
	}
return len(likes), nil
}
func CommentDisLikeCounter(CommentID string) (int ,error){
	var likes []CommentLike
	rows, err := db.Query("SELECT is_like  FROM Commentlikes WHERE comment_id = ? AND is_like = -1 ", CommentID)
	if err != nil {
		return 0, errors.New(" not found")
	}
	defer rows.Close()
	
		for rows.Next() {
			var like CommentLike
			if err := rows.Scan(&like.IsLike); err != nil {
				return 0, err
			}
			likes = append(likes, like)
		}
	return len(likes), nil
}

func CommentAddLike(CommentID, userID, Liked string) {
	stat, _ := db.Prepare("INSERT INTO Commentlikes (comment_id , user_id, is_like) VALUES (?,?,?)")
	user, _ := GetUserByUserName(userID)
	stat.Exec(CommentID, user.ID, Liked)
}
func CommentRemoveLike(CommentID, userID string) {
	stat, _ := db.Prepare("DELETE FROM Commentlikes WHERE comment_id = ? AND user_id = ?")
	user, _ := GetUserByUserName(userID)
	stat.Exec(CommentID, user.ID)
}
func CommentUpdateLike(CommentID, userID, Liked string) {
	statement, _ := db.Prepare("UPDATE Commentlikes SET is_like = ? WHERE comment_id = ? AND user_id = ?")
	user, _ := GetUserByUserName(userID)
	statement.Exec(Liked, CommentID, user.ID)
}

func CommentIsLike(CommentID, userID string) bool {
	user, _ := GetUserByUserName(userID)

	rows, _ := db.Query("SELECT is_like FROM Commentlikes WHERE user_id = ? AND comment_id = ? AND is_like = 1", user.ID, CommentID)
	like := 0
	for rows.Next() {
		rows.Scan(&like)
	}
	if like != 0 {
		return true
	}
	return false
}
func CommentIsDisLike(CommentID, userID string) bool {
	user, _ := GetUserByUserName(userID)

	rows, _ := db.Query("SELECT is_like FROM Commentlikes WHERE user_id = ? AND comment_id = ? AND is_like = -1", user.ID, CommentID)
	like := 0
	for rows.Next() {
		rows.Scan(&like)
	}
	if like != 0 {
		return true
	}
	return false
}
