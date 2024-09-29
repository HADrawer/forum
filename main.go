package main

import (
	"Forum/handlers"
	 "Forum/models"
	"html/template"
	"log"
	"net/http"
	
)
var templates *template.Template

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    handlers.RenderTemplate(w, "register", nil)
}
// --- Main Function ---

func main() {
	// Initialize the database
	models.InitDB()
	

	
	// Routes
	
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login",handlers.LoginHandler)
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		handlers.DestroySession(w, r)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	http.HandleFunc("/createPost", handlers.CreatePostHandler)
	http.HandleFunc("/viewPost", handlers.ViewPostHandler)

	// Serve static files (CSS, images)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start server
	log.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}

