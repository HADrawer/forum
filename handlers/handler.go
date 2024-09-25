package handlers


import(

     temp "Forum/templates"
	"net/http"
    "log"
)



func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
    templates := temp.Must(temp.ParseFiles(
        "templates/base.html",
        "templates/" + tmpl + ".html",
    ))
    err := templates.ExecuteTemplate(w, "base", data)
    if err != nil {
        http.Error(w, "Unable to load template", http.StatusInternalServerError)
    }
}


func homeHandler(w http.ResponseWriter, r *http.Request) {
    data := struct {
        Title     string
        IsLoggedIn bool
        Posts     []Post // Assuming Post is a struct you've defined for posts
    }{
        Title: "Home",
        IsLoggedIn: false, // You will update this later with session data
        Posts: []Post{}, // Fetch posts from database here
    }

    renderTemplate(w, "home", data)
}


func registerHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        data := struct {
            Title string
        }{
            Title: "Register",
        }
        renderTemplate(w, "register", data)
    } else if r.Method == http.MethodPost {
        // Handle form submission logic for user registration
    }
}


func loginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        w.Header().Set("Content-Type", "text/html")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`<form action="/login" method="post">
                            Email: <input type="email" name="email"><br>
                            Password: <input type="password" name="password"><br>
                            <input type="submit" value="Login">
                        </form>`))
    } else if r.Method == http.MethodPost {
        email := r.FormValue("email")
        password := r.FormValue("password")

        // Logic to validate credentials will go here (compare email & password with DB)

        log.Printf("User logged in: %s", email)

      //after validation is successful, you'd get the userID here

		createSession(w, userID)

        // For now, just redirect to home page after "login"

        http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}



func createPostHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        w.Header().Set("Content-Type", "text/html")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`<form action="/createPost" method="post">
                            Title: <input type="text" name="title"><br>
                            Content: <textarea name="content"></textarea><br>
                            <input type="submit" value="Create Post">
                        </form>`))
    } else if r.Method == http.MethodPost {
        title := r.FormValue("title")
        content := r.FormValue("content")

        // Logic to store post in the database

        log.Printf("New post created: %s", title)

        // Redirect to home page after post creation
        http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}

func viewPostHandler(w http.ResponseWriter, r *http.Request) {
    postID := r.URL.Query().Get("id")

    // Logic to fetch the post from the database

    w.Header().Set("Content-Type", "text/html")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("<h1>Viewing Post: " + postID + "</h1><p>Post content goes here...</p>"))
}


func likeHandler(w http.ResponseWriter, r *http.Request) {
    postID := r.URL.Query().Get("post_id")
    like := r.URL.Query().Get("like") // "1" for like, "0" for dislike

    // Logic to update the like/dislike in the database

    log.Printf("Post %s liked: %s", postID, like)
    http.Redirect(w, r, "/viewPost?id="+postID, http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
    // Destroy the session
    destroySession(w, r)

    // Redirect to home page
    http.Redirect(w, r, "/", http.StatusSeeOther)
}
