package main

import (
    "database/sql"
    "net/http"
    "github.com/gin-gonic/gin"
    _ "github.com/mattn/go-sqlite3"
    "golang.org/x/crypto/bcrypt"
    "log"
)


var (
    db       *sql.DB
    loggedIn bool
    userID   int
)

func main() {
    var err error
    db, err = sql.Open("sqlite3", "./webapp.db")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    // Initialize the database schema
    initializeDatabase()

    r := gin.Default()

    // Serve HTML files
    r.LoadHTMLGlob("templates/*")

    // Routes
	r.GET("/register", showRegisterPage)
    r.POST("/register", handleRegister)
    r.GET("/login", showLoginPage)
    r.POST("/login", handleLogin)
    
    r.GET("/logout", handleLogout)
    r.GET("/", showHomePage)
    r.GET("/categories", showCategories)
    r.GET("/category/:id", showPostsInCategory)
    r.POST("/post/:id/like", likePost)
    r.POST("/post/:id/dislike", dislikePost)
    r.POST("/post/:id/comment", commentOnPost)

    r.Run(":8080")
}

// Initialize the database schema
func initializeDatabase() {
    schema := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS categories (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT UNIQUE NOT NULL
    );

    CREATE TABLE IF NOT EXISTS posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        category_id INTEGER,
        title TEXT NOT NULL,
        content TEXT,
        FOREIGN KEY (category_id) REFERENCES categories(id)
    );

    CREATE TABLE IF NOT EXISTS likes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        post_id INTEGER,
        is_liked BOOLEAN,
        FOREIGN KEY (user_id) REFERENCES users(id),
        FOREIGN KEY (post_id) REFERENCES posts(id)
    );

    CREATE TABLE IF NOT EXISTS comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        post_id INTEGER,
        user_id INTEGER,
        content TEXT,
        FOREIGN KEY (post_id) REFERENCES posts(id),
        FOREIGN KEY (user_id) REFERENCES users(id)
    );
    `

    _, err := db.Exec(schema)
    if err != nil {
        log.Fatalf("Error initializing database: %v", err)
    }
}

// Handle user authentication
func authenticate(email, password string) (int, bool) {
    var id int
    var hashedPassword string

    err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&id, &hashedPassword)
    if err != nil {
        return 0, false
    }

    if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
        return 0, false
    }

    return id, true
}

// Handle user registration
func registerUser(email, password string) bool {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return false
    }

    _, err = db.Exec("INSERT INTO users (email, password) VALUES (?, ?)", email, hashedPassword)
    return err == nil
}

// Show login page
func showLoginPage(c *gin.Context) {
    c.HTML(http.StatusOK, "login.html", nil)
}

// Handle login
func handleLogin(c *gin.Context) {
    email := c.PostForm("email")
    password := c.PostForm("password")

    var ok bool
    userID, ok = authenticate(email, password)
    if ok {
        loggedIn = true
        c.Redirect(http.StatusSeeOther, "/")
    } else {
        c.HTML(http.StatusUnauthorized, "login.html", gin.H{"Error": "Invalid email or password"})
    }
}

// Show register page
func showRegisterPage(c *gin.Context) {
    c.HTML(http.StatusOK, "register.html", nil)
}

// Handle registration
func handleRegister(c *gin.Context) {
    email := c.PostForm("email")
    password := c.PostForm("password")

    if registerUser(email, password) {
        c.Redirect(http.StatusSeeOther, "/login")
    } else {
        c.HTML(http.StatusInternalServerError, "register.html", gin.H{"Error": "Failed to register"})
    }
}

// Handle logout
func handleLogout(c *gin.Context) {
    loggedIn = false
    userID = 0
    c.Redirect(http.StatusSeeOther, "/login")
}

// Show home page
func showHomePage(c *gin.Context) {
    if !loggedIn {
        c.Redirect(http.StatusSeeOther, "/login")
        return
    }
    c.HTML(http.StatusOK, "home.html", nil)
}

// Show categories
func showCategories(c *gin.Context) {
    rows, err := db.Query("SELECT id, name FROM categories")
    if err != nil {
        c.String(http.StatusInternalServerError, "Failed to retrieve categories")
        return
    }
    defer rows.Close()

    var categories []map[string]interface{}
    for rows.Next() {
        var id int
        var name string
        if err := rows.Scan(&id, &name); err != nil {
            c.String(http.StatusInternalServerError, "Failed to scan categories")
            return
        }
        categories = append(categories, map[string]interface{}{
            "id":   id,
            "name": name,
        })
    }

    c.HTML(http.StatusOK, "categories.html", gin.H{
        "categories": categories,
    })
}

// Show posts in a category
func showPostsInCategory(c *gin.Context) {
    categoryID := c.Param("id")

    rows, err := db.Query("SELECT id, title, content FROM posts WHERE category_id = ?", categoryID)
    if err != nil {
        c.String(http.StatusInternalServerError, "Failed to retrieve posts")
        return
    }
    defer rows.Close()

    var posts []map[string]interface{}
    for rows.Next() {
        var id int
        var title, content string
        if err := rows.Scan(&id, &title, &content); err != nil {
            c.String(http.StatusInternalServerError, "Failed to scan posts")
            return
        }
        posts = append(posts, map[string]interface{}{
            "id":      id,
            "title":   title,
            "content": content,
        })
    }

    c.HTML(http.StatusOK, "posts.html", gin.H{
        "posts": posts,
    })
}

// Like a post
func likePost(c *gin.Context) {
    postID := c.Param("id")
    _, err := db.Exec("INSERT INTO likes (user_id, post_id, is_liked) VALUES (?, ?, true) ON CONFLICT(user_id, post_id) DO UPDATE SET is_liked = true", userID, postID)
    if err != nil {
        c.String(http.StatusInternalServerError, "Failed to like post")
        return
    }
    c.Redirect(http.StatusSeeOther, "/category/"+postID)
}

// Dislike a post
func dislikePost(c *gin.Context) {
    postID := c.Param("id")
    _, err := db.Exec("INSERT INTO likes (user_id, post_id, is_liked) VALUES (?, ?, false) ON CONFLICT(user_id, post_id) DO UPDATE SET is_liked = false", userID, postID)
    if err != nil {
        c.String(http.StatusInternalServerError, "Failed to dislike post")
        return
    }
    c.Redirect(http.StatusSeeOther, "/category/"+postID)
}

// Comment on a post
func commentOnPost(c *gin.Context) {
    postID := c.Param("id")
    content := c.PostForm("content")

    _, err := db.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)", postID, userID, content)
    if err != nil {
        c.String(http.StatusInternalServerError, "Failed to comment on post")
        return
    }
    c.Redirect(http.StatusSeeOther, "/category/"+postID)
}

// package main

// import (
//     "database/sql"
//     "net/http"
//     "github.com/gin-gonic/gin"
//     _ "github.com/mattn/go-sqlite3"
//     "golang.org/x/crypto/bcrypt"
//     "log"
// )

// var (
//     db       *sql.DB
//     loggedIn bool
//     userID   int
// )

// func main() {
//     var err error
//     db, err = sql.Open("sqlite3", "./webapp.db")
//     if err != nil {
//         panic(err)
//     }
//     defer db.Close()

//     // Initialize the database schema
//     initializeDatabase()

//     r := gin.Default()

//     // Serve HTML files
//     r.LoadHTMLGlob("templates/*")

//     // Routes
//     r.GET("/login", showLoginPage)
//     r.POST("/login", handleLogin)
//     r.GET("/register", showRegisterPage)
//     r.POST("/register", handleRegister)
//     r.GET("/logout", handleLogout)
//     r.GET("/", showHomePage)
//     r.GET("/categories", showCategories)
//     r.GET("/category/:id", showPostsInCategory)
//     r.POST("/post/:id/like", likePost)
//     r.POST("/post/:id/dislike", dislikePost)
//     r.POST("/post/:id/comment", commentOnPost)

//     r.Run(":8080")
// }

// // Initialize the database schema
// func initializeDatabase() {
//     schema := `
//     CREATE TABLE IF NOT EXISTS users (
//         id INTEGER PRIMARY KEY AUTOINCREMENT,
//         email TEXT UNIQUE NOT NULL,
//         password TEXT NOT NULL
//     );

//     CREATE TABLE IF NOT EXISTS categories (
//         id INTEGER PRIMARY KEY AUTOINCREMENT,
//         name TEXT UNIQUE NOT NULL
//     );

//     CREATE TABLE IF NOT EXISTS posts (
//         id INTEGER PRIMARY KEY AUTOINCREMENT,
//         category_id INTEGER,
//         title TEXT NOT NULL,
//         content TEXT,
//         FOREIGN KEY (category_id) REFERENCES categories(id)
//     );

//     CREATE TABLE IF NOT EXISTS likes (
//         id INTEGER PRIMARY KEY AUTOINCREMENT,
//         user_id INTEGER,
//         post_id INTEGER,
//         is_liked BOOLEAN,
//         FOREIGN KEY (user_id) REFERENCES users(id),
//         FOREIGN KEY (post_id) REFERENCES posts(id)
//     );

//     CREATE TABLE IF NOT EXISTS comments (
//         id INTEGER PRIMARY KEY AUTOINCREMENT,
//         post_id INTEGER,
//         user_id INTEGER,
//         content TEXT,
//         FOREIGN KEY (post_id) REFERENCES posts(id),
//         FOREIGN KEY (user_id) REFERENCES users(id)
//     );
//     `

//     _, err := db.Exec(schema)
//     if err != nil {
//         log.Fatalf("Error initializing database: %v", err)
//     }
// }

// // Helper function to handle user authentication
// func authenticate(email, password string) (int, bool) {
//     var id int
//     var hashedPassword string

//     err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&id, &hashedPassword)
//     if err != nil {
//         return 0, false
//     }

//     if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
//         return 0, false
//     }

//     return id, true
// }

// // Helper function to handle user registration
// func registerUser(email, password string) bool {
//     hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
//     if err != nil {
//         return false
//     }

//     _, err = db.Exec("INSERT INTO users (email, password) VALUES (?, ?)", email, hashedPassword)
//     return err == nil
// }

// func showLoginPage(c *gin.Context) {
//     c.HTML(http.StatusOK, "login.html", nil)
// }

// func handleLogin(c *gin.Context) {
//     email := c.PostForm("email")
//     password := c.PostForm("password")

//     var ok bool
//     userID, ok = authenticate(email, password)
//     if ok {
//         loggedIn = true
//         c.Redirect(http.StatusSeeOther, "/")
//     } else {
//         c.HTML(http.StatusUnauthorized, "login.html", gin.H{"Error": "Invalid email or password"})
//     }
// }

// func showRegisterPage(c *gin.Context) {
//     c.HTML(http.StatusOK, "register.html", nil)
// }

// func handleRegister(c *gin.Context) {
//     email := c.PostForm("email")
//     password := c.PostForm("password")

//     if registerUser(email, password) {
//         c.Redirect(http.StatusSeeOther, "/login")
//     } else {
//         c.HTML(http.StatusInternalServerError, "register.html", gin.H{"Error": "Failed to register"})
//     }
// }

// func handleLogout(c *gin.Context) {
//     loggedIn = false
//     userID = 0
//     c.Redirect(http.StatusSeeOther, "/login")
// }

// func showHomePage(c *gin.Context) {
//     if !loggedIn {
//         c.Redirect(http.StatusSeeOther, "/login")
//         return
//     }
//     c.HTML(http.StatusOK, "home.html", nil)
// }

// func showCategories(c *gin.Context) {
//     rows, err := db.Query("SELECT id, name FROM categories")
//     if err != nil {
//         c.String(http.StatusInternalServerError, "Failed to retrieve categories")
//         return
//     }
//     defer rows.Close()

//     var categories []map[string]interface{}
//     for rows.Next() {
//         var id int
//         var name string
//         if err := rows.Scan(&id, &name); err != nil {
//             c.String(http.StatusInternalServerError, "Failed to scan categories")
//             return
//         }
//         categories = append(categories, map[string]interface{}{
//             "id":   id,
//             "name": name,
//         })
//     }

//     c.HTML(http.StatusOK, "categories.html", gin.H{
//         "categories": categories,
//     })
// }

// func showPostsInCategory(c *gin.Context) {
//     categoryID := c.Param("id")

//     rows, err := db.Query("SELECT id, title, content FROM posts WHERE category_id = ?", categoryID)
//     if err != nil {
//         c.String(http.StatusInternalServerError, "Failed to retrieve posts")
//         return
//     }
//     defer rows.Close()

//     var posts []map[string]interface{}
//     for rows.Next() {
//         var id int
//         var title, content string
//         if err := rows.Scan(&id, &title, &content); err != nil {
//             c.String(http.StatusInternalServerError, "Failed to scan posts")
//             return
//         }
//         posts = append(posts, map[string]interface{}{
//             "id":      id,
//             "title":   title,
//             "content": content,
//         })
//     }

//     c.HTML(http.StatusOK, "posts.html", gin.H{
//         "posts": posts,
//     })
// }

// func likePost(c *gin.Context) {
//     postID := c.Param("id")
//     _, err := db.Exec("INSERT INTO likes (user_id, post_id, is_liked) VALUES (?, ?, true) ON CONFLICT(user_id, post_id) DO UPDATE SET is_liked = true", userID, postID)
//     if err != nil {
//         c.String(http.StatusInternalServerError, "Failed to like post")
//         return
//     }
//     c.Redirect(http.StatusSeeOther, "/category/"+postID)
// }

// func dislikePost(c *gin.Context) {
//     postID := c.Param("id")
//     _, err := db.Exec("INSERT INTO likes (user_id, post_id, is_liked) VALUES (?, ?, false) ON CONFLICT(user_id, post_id) DO UPDATE SET is_liked = false", userID, postID)
//     if err != nil {
//         c.String(http.StatusInternalServerError, "Failed to dislike post")
//         return
//     }
//     c.Redirect(http.StatusSeeOther, "/category/"+postID)
// }

// func commentOnPost(c *gin.Context) {
//     postID := c.Param("id")
//     content := c.PostForm("content")

//     _, err := db.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)", postID, userID, content)
//     if err != nil {
//         c.String(http.StatusInternalServerError, "Failed to comment on post")
//         return
//     }
//     c.Redirect(http.StatusSeeOther, "/category/"+postID)
// }

// func handleLogin(c *gin.Context) {
//     email := c.PostForm("email")
//     password := c.PostForm("password")

//     var ok bool
//     userID, ok = authenticate(email, password)
//     if ok {
//         loggedIn = true
//         c.Redirect(http.StatusSeeOther, "/")
//     } else {
//         c.HTML(http.StatusUnauthorized, "login.html", gin.H{"Error": "Invalid email or password"})
//     }
// }

// func authenticate(email, password string) (int, bool) {
//     var id int
//     var hashedPassword string

//     err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&id, &hashedPassword)
//     if err != nil {
//         return 0, false
//     }

//     if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
//         return 0, false
//     }

//     return id, true
// }

// func registerUser(email, password string) bool {
//     hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
//     if err != nil {
//         return false
//     }

//     _, err = db.Exec("INSERT INTO users (email, password) VALUES (?, ?)", email, hashedPassword)
//     return err == nil
// }


// package main

// import (
//     "database/sql"
//     "net/http"
//     "github.com/gin-gonic/gin"
//     _ "github.com/mattn/go-sqlite3"
//     "golang.org/x/crypto/bcrypt"
//     "log"
//     "github.com/satori/go.uuid"
// )

// var (
//     db       *sql.DB
//     loggedIn bool
//     userID   int
// )

// func main() {
//     var err error
//     db, err = sql.Open("sqlite3", "./webapp.db")
//     if err != nil {
//         panic(err)
//     }
//     defer db.Close()

//     // Initialize the database schema
//     initializeDatabase()

//     r := gin.Default()

//     // Serve HTML files
//     r.LoadHTMLGlob("templates/*")

//     // Routes
//     r.GET("/login", showLoginPage)
//     r.POST("/login", handleLogin)
//     r.GET("/register", showRegisterPage)
//     r.POST("/register", handleRegister)
//     r.GET("/logout", handleLogout)
//     r.GET("/", showHomePage)
//     r.GET("/categories", showCategories)
//     r.GET("/category/:id", showPostsInCategory)
//     r.POST("/post/:id/like", likePost)
//     r.POST("/post/:id/dislike", dislikePost)
//     r.POST("/post/:id/comment", commentOnPost)

//     r.Run(":8080")
// }

// // Initialize the database schema
// func initializeDatabase() {
//     schema := `
//     CREATE TABLE IF NOT EXISTS users (
//         id INTEGER PRIMARY KEY AUTOINCREMENT,
//         email TEXT UNIQUE NOT NULL,
//         password TEXT NOT NULL
//     );

//     CREATE TABLE IF NOT EXISTS categories (
//         id INTEGER PRIMARY KEY AUTOINCREMENT,
//         name TEXT UNIQUE NOT NULL
//     );

//     CREATE TABLE IF NOT EXISTS posts (
//         id INTEGER PRIMARY KEY AUTOINCREMENT,
//         category_id INTEGER,
//         title TEXT NOT NULL,
//         content TEXT,
//         FOREIGN KEY (category_id) REFERENCES categories(id)
//     );

//     CREATE TABLE IF NOT EXISTS likes (
//         id INTEGER PRIMARY KEY AUTOINCREMENT,
//         user_id INTEGER,
//         post_id INTEGER,
//         is_liked BOOLEAN,
//         FOREIGN KEY (user_id) REFERENCES users(id),
//         FOREIGN KEY (post_id) REFERENCES posts(id)
//     );

//     CREATE TABLE IF NOT EXISTS comments (
//         id INTEGER PRIMARY KEY AUTOINCREMENT,
//         post_id INTEGER,
//         user_id INTEGER,
//         content TEXT,
//         FOREIGN KEY (post_id) REFERENCES posts(id),
//         FOREIGN KEY (user_id) REFERENCES users(id)
//     );
//     `

//     _, err := db.Exec(schema)
//     if err != nil {
//         log.Fatalf("Error initializing database: %v", err)
//     }
// }

// // Helper function to handle user authentication
// func authenticate(email, password string) (int, bool) {
//     var id int
//     var hashedPassword string

//     err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&id, &hashedPassword)
//     if err != nil {
//         return 0, false
//     }

//     if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
//         return 0, false
//     }

//     return id, true
// }

// // Helper function to handle user registration
// func registerUser(email, password string) bool {
//     hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
//     if err != nil {
//         return false
//     }

//     _, err = db.Exec("INSERT INTO users (email, password) VALUES (?, ?)", email, hashedPassword)
//     return err == nil
// }

// func showLoginPage(c *gin.Context) {
//     c.HTML(http.StatusOK, "login.html", nil)
// }

// func handleLogin(c *gin.Context) {
//     email := c.PostForm("email")
//     password := c.PostForm("password")

//     var ok bool
//     userID, ok = authenticate(email, password)
//     if ok {
//         loggedIn = true
//         c.Redirect(http.StatusSeeOther, "/")
//     } else {
//         c.HTML(http.StatusUnauthorized, "login.html", gin.H{"Error": "Invalid email or password"})
//     }
// }

// func showRegisterPage(c *gin.Context) {
//     c.HTML(http.StatusOK, "register.html", nil)
// }

// func handleRegister(c *gin.Context) {
//     email := c.PostForm("email")
//     password := c.PostForm("password")

//     if registerUser(email, password) {
//         c.Redirect(http.StatusSeeOther, "/login")
//     } else {
//         c.HTML(http.StatusInternalServerError, "register.html", gin.H{"Error": "Failed to register"})
//     }
// }

// func handleLogout(c *gin.Context) {
//     loggedIn = false
//     userID = 0
//     c.Redirect(http.StatusSeeOther, "/login")
// }

// func showHomePage(c *gin.Context) {
//     if !loggedIn {
//         c.Redirect(http.StatusSeeOther, "/login")
//         return
//     }
//     c.HTML(http.StatusOK, "home.html", nil)
// }

// func showCategories(c *gin.Context) {
//     rows, err := db.Query("SELECT id, name FROM categories")
//     if err != nil {
//         c.String(http.StatusInternalServerError, "Failed to retrieve categories")
//         return
//     }
//     defer rows.Close()

//     var categories []map[string]interface{}
//     for rows.Next() {
//         var id int
//         var name string
//         if err := rows.Scan(&id, &name); err != nil {
//             c.String(http.StatusInternalServerError, "Failed to scan categories")
//             return
//         }
//         categories = append(categories, map[string]interface{}{
//             "id":   id,
//             "name": name,
//         })
//     }

//     c.HTML(http.StatusOK, "categories.html", gin.H{
//         "categories": categories,
//     })
// }

// func showPostsInCategory(c *gin.Context) {
//     categoryID := c.Param("id")

//     rows, err := db.Query("SELECT id, title, content FROM posts WHERE category_id = ?", categoryID)
//     if err != nil {
//         c.String(http.StatusInternalServerError, "Failed to retrieve posts")
//         return
//     }
//     defer rows.Close()

//     var posts []map[string]interface{}
//     for rows.Next() {
//         var id int
//         var title, content string
//         if err := rows.Scan(&id, &title, &content); err != nil {
//             c.String(http.StatusInternalServerError, "Failed to scan posts")
//             return
//         }
//         posts = append(posts, map[string]interface{}{
//             "id":      id,
//             "title":   title,
//             "content": content,
//         })
//     }

//     c.HTML(http.StatusOK, "posts.html", gin.H{
//         "posts": posts,
//     })
// }

// func likePost(c *gin.Context) {
//     postID := c.Param("id")
//     _, err := db.Exec("INSERT INTO likes (user_id, post_id, is_liked) VALUES (?, ?, true) ON CONFLICT(user_id, post_id) DO UPDATE SET is_liked = true", userID, postID)
//     if err != nil {
//         c.String(http.StatusInternalServerError, "Failed to like post")
//         return
//     }
//     c.Redirect(http.StatusSeeOther, "/category/"+postID)
// }

// func dislikePost(c *gin.Context) {
//     postID := c.Param("id")
//     _, err := db.Exec("INSERT INTO likes (user_id, post_id, is_liked) VALUES (?, ?, false) ON CONFLICT(user_id, post_id) DO UPDATE SET is_liked = false", userID, postID)
//     if err != nil {
//         c.String(http.StatusInternalServerError, "Failed to dislike post")
//         return
//     }
//     c.Redirect(http.StatusSeeOther, "/category/"+postID)
// }

// func commentOnPost(c *gin.Context) {
//     postID := c.Param("id")
//     content := c.PostForm("content")

//     _, err := db.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)", postID, userID, content)
//     if err != nil {
//         c.String(http.StatusInternalServerError, "Failed to comment on post")
//         return
//     }
//     c.Redirect(http.StatusSeeOther, "/category/"+postID)
// }

// // package main

// // import (
// //     "database/sql"
// //     "net/http"
// //     "github.com/gin-gonic/gin"
// //     _ "github.com/go-sql-driver/mysql"
// //     "golang.org/x/crypto/bcrypt"
// //     "log"
// // )

// // var (
// //     db        *sql.DB
// //     loggedIn  bool
// //     userID    int
// // )

// // func main() {
// //     var err error
// //     db, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/webapp")
// //     if err != nil {
// //         panic(err)
// //     }
// //     defer db.Close()

// //     r := gin.Default()

// //     // Serve HTML files
// //     r.LoadHTMLGlob("templates/*")

// //     // Routes
// //     r.GET("/login", showLoginPage)
// //     r.POST("/login", handleLogin)
// //     r.GET("/register", showRegisterPage)
// //     r.POST("/register", handleRegister)
// //     r.GET("/logout", handleLogout)
// //     r.GET("/", showHomePage)
// //     r.GET("/categories", showCategories)
// //     r.GET("/category/:id", showPostsInCategory)
// //     r.POST("/post/:id/like", likePost)
// //     r.POST("/post/:id/dislike", dislikePost)
// //     r.POST("/post/:id/comment", commentOnPost)

// //     r.Run(":8080")
// // }

// // // Helper function to handle user authentication
// // func authenticate(email, password string) (int, bool) {
// //     var id int
// //     var hashedPassword string

// //     err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&id, &hashedPassword)
// //     if err != nil {
// //         return 0, false
// //     }

// //     if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
// //         return 0, false
// //     }

// //     return id, true
// // }

// // // Helper function to handle user registration
// // func registerUser(email, password string) bool {
// //     hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// //     if err != nil {
// //         return false
// //     }

// //     _, err = db.Exec("INSERT INTO users (email, password) VALUES (?, ?)", email, hashedPassword)
// //     return err == nil
// // }

// // func showLoginPage(c *gin.Context) {
// //     c.HTML(http.StatusOK, "login.html", nil)
// // }

// // func handleLogin(c *gin.Context) {
// //     email := c.PostForm("email")
// //     password := c.PostForm("password")

// //     var ok bool
// //     userID, ok = authenticate(email, password)
// //     if ok {
// //         loggedIn = true
// //         c.Redirect(http.StatusSeeOther, "/")
// //     } else {
// //         c.HTML(http.StatusUnauthorized, "login.html", gin.H{"Error": "Invalid email or password"})
// //     }
// // }

// // func showRegisterPage(c *gin.Context) {
// //     c.HTML(http.StatusOK, "register.html", nil)
// // }

// // func handleRegister(c *gin.Context) {
// //     email := c.PostForm("email")
// //     password := c.PostForm("password")

// //     if registerUser(email, password) {
// //         c.Redirect(http.StatusSeeOther, "/login")
// //     } else {
// //         c.HTML(http.StatusInternalServerError, "register.html", gin.H{"Error": "Failed to register"})
// //     }
// // }

// // func handleLogout(c *gin.Context) {
// //     loggedIn = false
// //     userID = 0
// //     c.Redirect(http.StatusSeeOther, "/login")
// // }

// // func showHomePage(c *gin.Context) {
// //     if !loggedIn {
// //         c.Redirect(http.StatusSeeOther, "/login")
// //         return
// //     }
// //     c.HTML(http.StatusOK, "home.html", nil)
// // }

// // func showCategories(c *gin.Context) {
// //     rows, err := db.Query("SELECT id, name FROM categories")
// //     if err != nil {
// //         c.String(http.StatusInternalServerError, "Failed to retrieve categories")
// //         return
// //     }
// //     defer rows.Close()

// //     var categories []map[string]interface{}
// //     for rows.Next() {
// //         var id int
// //         var name string
// //         if err := rows.Scan(&id, &name); err != nil {
// //             c.String(http.StatusInternalServerError, "Failed to scan categories")
// //             return
// //         }
// //         categories = append(categories, map[string]interface{}{
// //             "id":   id,
// //             "name": name,
// //         })
// //     }

// //     c.HTML(http.StatusOK, "categories.html", gin.H{
// //         "categories": categories,
// //     })
// // }

// // func showPostsInCategory(c *gin.Context) {
// //     categoryID := c.Param("id")

// //     rows, err := db.Query("SELECT id, title, content FROM posts WHERE category_id = ?", categoryID)
// //     if err != nil {
// //         c.String(http.StatusInternalServerError, "Failed to retrieve posts")
// //         return
// //     }
// //     defer rows.Close()

// //     var posts []map[string]interface{}
// //     for rows.Next() {
// //         var id int
// //         var title, content string
// //         if err := rows.Scan(&id, &title, &content); err != nil {
// //             c.String(http.StatusInternalServerError, "Failed to scan posts")
// //             return
// //         }
// //         posts = append(posts, map[string]interface{}{
// //             "id":      id,
// //             "title":   title,
// //             "content": content,
// //         })
// //     }

// //     c.HTML(http.StatusOK, "posts.html", gin.H{
// //         "posts": posts,
// //     })
// // }

// // func likePost(c *gin.Context) {
// //     postID := c.Param("id")
// //     _, err := db.Exec("INSERT INTO likes (user_id, post_id, is_liked) VALUES (?, ?, true) ON DUPLICATE KEY UPDATE is_liked = true", userID, postID)
// //     if err != nil {
// //         c.String(http.StatusInternalServerError, "Failed to like post")
// //         return
// //     }
// //     c.Redirect(http.StatusSeeOther, "/category/"+postID)
// // }

// // func dislikePost(c *gin.Context) {
// //     postID := c.Param("id")
// //     _, err := db.Exec("INSERT INTO likes (user_id, post_id, is_liked) VALUES (?, ?, false) ON DUPLICATE KEY UPDATE is_liked = false", userID, postID)
// //     if err != nil {
// //         c.String(http.StatusInternalServerError, "Failed to dislike post")
// //         return
// //     }
// //     c.Redirect(http.StatusSeeOther, "/category/"+postID)
// // }

// // func commentOnPost(c *gin.Context) {
// //     postID := c.Param("id")
// //     content := c.PostForm("content")

// //     _, err := db.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)", postID, userID, content)
// //     if err != nil {
// //         c.String(http.StatusInternalServerError, "Failed to comment on post")
// //         return
// //     }
// //     c.Redirect(http.StatusSeeOther, "/category/"+postID)
// // }

// // // package main

// // // import (
// // //     "database/sql"
// // //     "net/http"
// // //     "github.com/gin-gonic/gin"
// // //     _ "github.com/go-sql-driver/mysql"
// // // )

// // // var (
// // //     db        *sql.DB
// // //     loggedIn  bool
// // // )

// // // func main() {
// // //     var err error
// // //     db, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/webapp")
// // //     if err != nil {
// // //         panic(err)
// // //     }
// // //     defer db.Close()

// // //     r := gin.Default()

// // //     // Serve HTML files
// // //     r.LoadHTMLGlob("templates/*")

// // //     // Login route
// // //     r.GET("/login", func(c *gin.Context) {
// // //         c.HTML(http.StatusOK, "login.html", nil)
// // //     })

// // //     r.POST("/login", func(c *gin.Context) {
// // //         username := c.PostForm("username")
// // //         password := c.PostForm("password")

// // //         // Simple hardcoded check - replace with real authentication
// // //         if username == "admin" && password == "password" {
// // //             loggedIn = true
// // //             c.Redirect(http.StatusSeeOther, "/")
// // //         } else {
// // //             c.HTML(http.StatusUnauthorized, "login.html", gin.H{
// // //                 "Error": "Invalid username or password",
// // //             })
// // //         }
// // //     })

// // //     // Middleware to check login status
// // //     r.Use(func(c *gin.Context) {
// // //         if !loggedIn && c.Request.URL.Path != "/login" {
// // //             c.Redirect(http.StatusSeeOther, "/login")
// // //             return
// // //         }
// // //         c.Next()
// // //     })

// // //     // Home route
// // //     r.GET("/", func(c *gin.Context) {
// // //         c.HTML(http.StatusOK, "index.html", nil)
// // //     })

// // //     // Submit route
// // //     r.POST("/submit", func(c *gin.Context) {
// // //         content := c.PostForm("content")

// // //         _, err := db.Exec("INSERT INTO messages (content) VALUES (?)", content)
// // //         if err != nil {
// // //             c.String(http.StatusInternalServerError, "Failed to insert data")
// // //             return
// // //         }
// // //         c.Redirect(http.StatusSeeOther, "/")
// // //     })

// // //     // Messages route
// // //     r.GET("/messages", func(c *gin.Context) {
// // //         rows, err := db.Query("SELECT id, content FROM messages")
// // //         if err != nil {
// // //             c.String(http.StatusInternalServerError, "Failed to retrieve messages")
// // //             return
// // //         }
// // //         defer rows.Close()

// // //         var messages []map[string]interface{}
// // //         for rows.Next() {
// // //             var id int
// // //             var content string
// // //             if err := rows.Scan(&id, &content); err != nil {
// // //                 c.String(http.StatusInternalServerError, "Failed to scan rows")
// // //                 return
// // //             }
// // //             messages = append(messages, map[string]interface{}{
// // //                 "id":      id,
// // //                 "content": content,
// // //             })
// // //         }

// // //         c.HTML(http.StatusOK, "messages.html", gin.H{
// // //             "messages": messages,
// // //         })
// // //     })

// // //     r.Run(":8080")
	
// // // }
