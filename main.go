package main
import (
	"Forum/handlers"
	"Forum/models"
	// "html/template"
	"log"
	"net/http"
)
// var templates *template.Template
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    handlers.RenderTemplate(w, "register", nil)
}

// --- Main Function ---
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNotFound)
    handlers.RenderTemplate(w, "404", nil)
}

func main() {
		// Initialize the database
		models.InitDB()
	
    // Routes
    http.HandleFunc("/", handlers.HomeHandler)
    http.HandleFunc("/home", handlers.HomeHandler)
    http.HandleFunc("/register", handlers.RegisterHandler)
    http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		handlers.DestroySession(w, r)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	http.HandleFunc("/createPost", handlers.CreatePostHandler)
	http.HandleFunc("/viewPost", handlers.ViewPostHandler)
	http.HandleFunc("/myposts", handlers.CreatedPostsHandler)

    // Add a fallback for unknown routes
    http.HandleFunc("/404", NotFoundHandler)

    // Serve static files
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    log.Println("Server is running on http://localhost:8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal("Failed to start server: ", err)
    }
}
// package main

// import (
// 	"Forum/handlers"
// 	"Forum/models"
// 	"html/template"
// 	"log"
// 	"net/http"
// )

// var templates *template.Template

// func RegisterHandler(w http.ResponseWriter, r *http.Request) {
// 	handlers.RenderTemplate(w, "register", nil)
// }

// // Wrapper function to handle 404 errors
// func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
// 	handlers.RenderErrorStatusCode(w, http.StatusNotFound)
// }

// func main() {
// 	// Initialize the database
// 	models.InitDB()

// 	// Routes
// 	http.HandleFunc("/", handlers.HomeHandler)
// 	http.HandleFunc("/register", RegisterHandler)
// 	http.HandleFunc("/login", handlers.LoginHandler)
// 	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
// 		handlers.DestroySession(w, r)
// 		http.Redirect(w, r, "/", http.StatusSeeOther)
// 	})
// 	http.HandleFunc("/createPost", handlers.CreatePostHandler)
// 	http.HandleFunc("/viewPost", handlers.ViewPostHandler)

// 	// Handle unknown routes with a 404 error
// 	http.HandleFunc("/404", NotFoundHandler)

// 	// Serve static files
// 	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
// 	log.Println("Server is running on http://localhost:8080")
// 	if err := http.ListenAndServe(":8080", nil); err != nil {
// 		log.Fatal("Failed to start server: ", err)
// 	}
// }
