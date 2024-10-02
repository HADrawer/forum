package handlers


import(

    "html/template"
    "golang.org/x/crypto/bcrypt"
	"net/http"
    "log"
    "Forum/models"
    
)
var templates = template.Must(template.ParseGlob("templates/*.html"))

// renderTemplate helper function
func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
    err := templates.ExecuteTemplate(w, tmpl+".html", data)
    if err != nil {
        http.Error(w, "Unable to load template", http.StatusInternalServerError)
    }
}

// BaseHandler serves pages with the base layout (base.html)
func BaseHandler(w http.ResponseWriter, r *http.Request,templateName string, data interface{}) {
    userID, isLoggedIn := GetUserIDFromSession(r)
    templateName = "base"
    
    pageData:= make(map[string]interface{})
    
    
    // Common data across all templates using base.html
    pageData["IsLoggedIn"] = isLoggedIn
    pageData["UserID"] = userID
    
    // Render the template with base.html as the layout
    err := templates.ExecuteTemplate(w, templateName+".html", pageData)
    if err != nil {
        http.Error(w, "Unable to render page", http.StatusInternalServerError)
    }
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	userID, isLoggedIn := GetUserIDFromSession(r)
	posts, err := models.GetAllPosts()
	if err != nil {
		http.Error(w, "Unable to load posts", http.StatusInternalServerError)
		return
	}
	data := struct {
		Title      string
		IsLoggedIn bool
		UserID     string
		Posts      []models.Post
	}{
		Title:      "Home",
		IsLoggedIn: isLoggedIn,
		UserID:     userID,
		Posts:      posts,
	}
	BaseHandler(w, r, "home", data)
}


func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        data := struct {
            Title string
        }{
            Title: "Register",
        }
        RenderTemplate(w, "register", data)
    } else if r.Method == http.MethodPost {
        // Extract form data
        email := r.FormValue("email")
        username := r.FormValue("username")
        password := r.FormValue("password")

        // Validate inputs
        if email == "" || username == "" || password == "" {
            http.Error(w, "All fields are required", http.StatusBadRequest)
            return
        }

        // Hash the password
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        if err != nil {
            log.Println("Error hashing password:", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        // Create a new user object
        newUser := models.User{
            Email:    email,
            Username: username,
            Password: string(hashedPassword),
        }

        // Save the user to the database
        err = models.CreateUser(newUser)
        if err != nil {
            if err == models.ErrUserExists {
                http.Error(w, "Email or username already exists", http.StatusBadRequest)
            } else {
                log.Println("Error creating user:", err)
                http.Error(w, "Internal server error", http.StatusInternalServerError)
            }
            return
        }
		
        // Redirect to the login page or home page
        http.Redirect(w, r, "/login", http.StatusSeeOther)
    }
}
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		RenderTemplate(w, "login", nil)
	} else if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		user, err := models.GetUserByEmail(email)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
			http.Error(w, "Invalid login", http.StatusUnauthorized)
			return
		}

		CreateSession(w, user.Username)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}


func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	userID, loggedIn := GetUserIDFromSession(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		RenderTemplate(w, "create_post", nil)
	} else if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")
		err := models.CreatePost(userID, title, content)
		if err != nil {
			http.Error(w, "Unable to create post", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}


func ViewPostHandler(w http.ResponseWriter, r *http.Request) {
    postID := r.URL.Query().Get("id")
    post, err := models.GetPostByID(postID)
    if err != nil {
        w.WriteHeader(http.StatusNotFound) // Set the 404 status code
        RenderTemplate(w, "404", nil) // Render custom 404 page
        return
    }

    comments, err := models.GetCommentsByPostID(postID)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError) // Set the 500 status code
        RenderTemplate(w, "500", nil) // Render custom 500 page
        return
    }

    data := struct {
        Post     *models.Post
        Comments []models.Comment
    }{
        Post:     post,
        Comments: comments,
    }

    RenderTemplate(w, "view_post", data)
}



func LikeHandler(w http.ResponseWriter, r *http.Request) {
    postID := r.URL.Query().Get("post_id")
    like := r.URL.Query().Get("like") // "1" for like, "0" for dislike

    // Logic to update the like/dislike in the database
    if postID == "" || (like != "1" && like != "0") {
        http.Error(w, "Invalid like or post ID", http.StatusBadRequest) // 400 Bad Request
        return
    }
    
    log.Printf("Post %s liked: %s", postID, like)
    http.Redirect(w, r, "/viewPost?id="+postID, http.StatusSeeOther)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
    // Destroy the session
    DestroySession(w, r)

    // Redirect to home page
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func CategoryHandler(w http.ResponseWriter, r *http.Request){
    
}
