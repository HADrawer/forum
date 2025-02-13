package handlers

import (
	"Forum/models"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

// renderTemplate helper function
func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	tmplPath := filepath.Join("templates", tmpl+".html")

	// Check if the requested template file exists
	if _, err := os.Stat(tmplPath); os.IsNotExist(err) {
		// Template not found, check if the 404 page exists
		if _, err404 := os.Stat(filepath.Join("templates", "404.html")); os.IsNotExist(err404) {
			// 404.html is missing, directly return 404 error
			http.Error(w, "404 page not found", http.StatusNotFound)
			return
		}

		// Render the 404 error page if the requested template is missing and 404 page exists
		err404 := templates.ExecuteTemplate(w, "404.html", nil)
		if err404 != nil {
			// If rendering 404 template fails, fallback to default 404 message
			http.Error(w, "404 page not found", http.StatusNotFound)
		}
		return
	}

	// Attempt to execute the requested template
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		log.Print(err)

		// Render the 500 error page if there's an internal server error
		err500 := templates.ExecuteTemplate(w, "500.html", nil)
		if err500 != nil {
			// If rendering 500 template fails, fallback to default 500 message
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}


// BaseHandler serves pages with the base layout (base.html)
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" &&  r.URL.Path != "/home" {
		w.WriteHeader(http.StatusNotFound) // Set the 404 status code
		RenderTemplate(w, "404", nil)      // Render custom 404 page
		return
	}
	if r.Method == http.MethodGet {
		
	
	
	userID, isLoggedIn := GetUserIDFromSession(r)
	// templateName = "base"
	pageData := make(map[string]interface{})
	// Common data across all templates using base.html
	pageData["IsLoggedIn"] = isLoggedIn
	pageData["UserID"] = userID

	Catagories, err := models.GetAllCategories()
	if err != nil {
		w.WriteHeader(http.StatusNotFound) // Set the 404 status code
		RenderTemplate(w, "404", nil)      // Render custom 404 page
		return
	}
	var CatagoryDetails []map[string]interface{}
	for _, Catagory := range Catagories {
		CatagoryDetail := map[string]interface{}{
			"Catagory": Catagory.Name,
		}
		CatagoryDetails = append(CatagoryDetails, CatagoryDetail)
	}
	pageData["Catagories"] = CatagoryDetails




	posts, err := models.GetAllPosts()
	if err != nil {
		http.Error(w, "Unable to load posts", http.StatusInternalServerError)
		RenderTemplate(w, "500", nil)   // 500
		return
	}
	isExist := true
	if posts == nil {
		isExist = false
	}

	// Render the template with posts
	var postDetails []map[string]interface{}
	for _, post := range posts {
		postDetail := map[string]interface{}{
			"Id":         post.ID,
			"Author":     post.Author,
			"Title":      post.Title,
			"created_at": post.Created_at,
		}
		postDetails = append(postDetails, postDetail)
	}
	
	pageData["isExist"] = isExist
	pageData["IsLoggedIn"] = isLoggedIn
	pageData["Title"] = "Liked"
	if isExist == false {
		pageData["NoPosts"] = "No Liked posts found."
	}
	pageData["Posts"] = postDetails


	

	// Render the template with base.html as the layout
	if pageData == nil {
		templates, err := template.ParseFiles("templates/500.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Println("Error loading 500 template:", err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		templates.Execute(w, nil)
		return
	}
	RenderTemplate(w, "base", pageData)
	// ExecuteTemplate
}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		
		RenderTemplate(w, "register", nil)
	} 
	if r.Method == http.MethodPost {
		// Extract form data
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")
		// Validate inputs
		if email == "" || username == "" || password == "" {

			t, err := template.ParseFiles("templates/500.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Println("Error loading 500 template:", err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		t.Execute(w, nil)
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
				pageData := map[string]interface{}{
					"InvalidRegister": "Email or username already exists",
				}
				RenderTemplate(w, "register", pageData)
				return
			} else {
				log.Println("Error creating user:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}
		user, err := models.GetUserByEmail(email)
		if err != nil {
			http.Error(w, "Invalid login", http.StatusUnauthorized)
			return
		}

		// Redirect to the login page or home page
		CreateSession(w, user.Username)
		http.Redirect(w, r, "/", http.StatusSeeOther)

	}
}
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		RenderTemplate(w, "login", nil)
		return
	} 
	
	if r.Method == http.MethodPost {
		Email_UserName := r.FormValue("email")
		password := r.FormValue("password")
		user, err := models.GetUserByEmail(Email_UserName)
		if err != nil {
			user, err = models.GetUserByUserName(Email_UserName)
		}
		
		if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
			pageData := map[string]interface{}{
				"InvalidLogin": "The Username or Password is Uncorrect",
			}
			RenderTemplate(w, "login", pageData)
			return
		}

		CreateSession(w, user.Username)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

//-----------------------------------------------------------------------


func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	
	userID, loggedIn := GetUserIDFromSession(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther) // 303
		return
	}
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	if r.Method == http.MethodGet {
		
		Catagories, _ := models.GetAllCategories()
		pageData := make(map[string]interface{})
		var postDetails []map[string]interface{}
	for _, Catagory := range Catagories {
		postDetail := map[string]interface{}{
			"Catagory": Catagory.Name,
		}
		postDetails = append(postDetails, postDetail)
	}
		pageData["Catagories"] = postDetails
		RenderTemplate(w, "createPost", pageData)
		
		return
	}
	 if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")
		categories := r.Form["categories[]"]
		stringCategories := strings.Join(categories, ",")
		


	

		// Check if title and content are provided
		if title == "" || content == "" || len(categories) == 0 {

			http.Error(w, "Bad request: Missing PostID or Comment", http.StatusBadRequest) 

		}

		// Attempt to create the post
		err := models.CreatePost(userID, title, content, stringCategories)
		if err != nil {
			http.Error(w, err.Error() , http.StatusInternalServerError) // 500
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther) // 303
	}
}

//-----------------------------------------------------------------------


func CreatedPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID, isLoggedIn := GetUserIDFromSession(r)

	// Check if user is logged in
	if !isLoggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther) // 303
		return
	}

	// Get posts for the logged-in user
	posts, err := models.GetPostsFromUserID(userID)
	if err != nil {
		http.Error(w, "Unable to load posts", http.StatusInternalServerError) // 500
		return
	}
	isExist := true
	if posts == nil {
		isExist = false
	}

	// Render the template with posts
	var postDetails []map[string]interface{}
	for _, post := range posts {
		postDetail := map[string]interface{}{
			"Id":         post.ID,
			"Author":     post.Author,
			"Title":      post.Title,
			"created_at": post.Created_at,
		}
		postDetails = append(postDetails, postDetail)
	}
	pageData := make(map[string]interface{})
	pageData["isExist"] = isExist
	pageData["IsLoggedIn"] = isLoggedIn
	pageData["Title"] = "My Created"
	if isExist == false {
		pageData["NoPosts"] = "No created posts found."
	}
	pageData["Posts"] = postDetails
	RenderTemplate(w, "ListsViewer", pageData)
	
}

func LikedPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID, isLoggedIn := GetUserIDFromSession(r)

	// Check if user is logged in
	if !isLoggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther) // 303
		return
	}

	// Get posts for the logged-in user
	posts, err := models.GetPostsFromLiked(userID)
	if err != nil {
		http.Error(w, "Unable to load posts", http.StatusInternalServerError) // 500
		return
	}
	isExist := true
	if posts == nil {
		isExist = false
	}

	// Render the template with posts
	var postDetails []map[string]interface{}
	for _, post := range posts {
		postDetail := map[string]interface{}{
			"Id":         post.ID,
			"Author":     post.Author,
			"Title":      post.Title,
			"created_at": post.Created_at,
		}
		postDetails = append(postDetails, postDetail)
	}
	pageData := make(map[string]interface{})
	pageData["isExist"] = isExist
	pageData["IsLoggedIn"] = isLoggedIn
	pageData["Title"] = "Liked"
	if isExist == false {
		pageData["NoPosts"] = "No Liked posts found."
	}
	pageData["Posts"] = postDetails
	RenderTemplate(w, "ListsViewer", pageData)
	
}

func ViewPostHandler(w http.ResponseWriter, r *http.Request) {
	_, isLoggedIn := GetUserIDFromSession(r)
	isExist := true
	id := r.URL.Query().Get("id")

	// Retrieve post by ID
	post, err0 := models.GetPostByID(id)
	comments, err := models.GetCommentsByPostID(id)
	if err0 != nil {
		w.WriteHeader(http.StatusNotFound) // 404
		RenderTemplate(w, "404", nil)      // Render custom 404 page if post not found
		return
	}

	// Check for errors in retrieving comments
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // 500
		RenderTemplate(w, "500", nil)                 // Render custom 500 page for internal error
		return
	}


	// Populate comments for the template
	var CommentDetails []map[string]interface{}
	for _, comment := range comments {
		
		CommentlikeCount , _ := models.CommentLikeCounter(strconv.Itoa(comment.ID))
		CommentDislikeCount , _ := models.CommentDisLikeCounter(strconv.Itoa(comment.ID))
		
		commentDetail := map[string]interface{}{
			"PostID" :		id,
			"id":			 comment.ID,
			"Author":        comment.Author,
			"comment":       comment.Content,
			"created_at":    comment.Created_at,
			"CommentUserID": comment.User_ID,
			"IsLoggedIn": 	isLoggedIn,
			"likes" : 		CommentlikeCount,
			"DisLikes" : 	CommentDislikeCount,

		}
		CommentDetails = append(CommentDetails, commentDetail)
	}
	likeCount , _ := models.LikeCounter(id)
	DislikeCount , _ := models.DisLikeCounter(id)	
	// Prepare page data with post details and comments
	pageData := make(map[string]interface{})
	pageData["id"] = id
	pageData["Author"] = post.Author
	pageData["Title"] = post.Title
	pageData["Content"] = post.Content
	pageData["IsLoggedIn"] = isLoggedIn
	pageData["isExist"] = isExist
	pageData["Comments"] = CommentDetails
	pageData["likes"] = likeCount
	pageData["DisLikes"] = DislikeCount

	// Render the view post template
	RenderTemplate(w, "viewPost", pageData)
	
}
func CatagoryHandler(w http.ResponseWriter, r *http.Request) {
	_, isLoggedIn := GetUserIDFromSession(r)
	catagory := r.FormValue("Catagory")
	isExist := true

	// Retrieve all posts in the category
	posts, err := models.GetAllCategoryPosts(catagory)
	if err != nil {
		
		w.WriteHeader(http.StatusNotFound) // 404
		RenderTemplate(w, "404", nil)      // Render custom 404 page for category not found
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // 500
		RenderTemplate(w, "500", nil)                 // Render custom 500 page for internal error
		return
	}

	pageData := make(map[string]interface{})

	// Check if any posts exist in the category
	if posts == nil {
		isExist = false
	}

	// Create a slice to hold the post details for the template
	var postDetails []map[string]interface{}
	for _, post := range posts {
		postDetail := map[string]interface{}{
			"IsLoggedIn": isLoggedIn,
			"Id":         post.ID,
			"Author":     post.Author,
			"Title":      post.Title,
			"created_at": post.Created_at,
		}
		postDetails = append(postDetails, postDetail)
	}

	// Populate page data with the relevant info
	pageData["IsLoggedIn"] = isLoggedIn
	pageData["Posts"] = postDetails
	pageData["isExist"] = isExist
	pageData["Title"] = catagory
	if isExist == false {
		pageData["NoPosts"] = "This Catagory is Empty."
	}

	// Render the category view template
	RenderTemplate(w, "ListsViewer", pageData)
	
}

//-----------------------------------------------------------------------
// Set a flag for cooldown period



func LikeHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := GetUserIDFromSession(r)
	postID := r.URL.Query().Get("post_id")
	like := r.URL.Query().Get("like") // "1" for like, "0" for dislike

	// Logic to update the like/dislike in the database
	if postID == "" || (like != "1" && like != "-1") {
		http.Error(w, "Bad request: Missing PostID or Comment", http.StatusBadRequest) // 400 Bad Request
		return
	}

	if like == "1" {
		if models.IsLike(postID, userID) {
			models.RemoveLike(postID, userID)
			
		} else if models.IsDisLike(postID, userID) {
			models.UpdateLike(postID, userID, "1")

		} else {
			models.AddLike(postID, userID, "1")
		}
	} else if like == "-1" {
		if models.IsDisLike(postID, userID) {
			models.RemoveLike(postID, userID)

		} else if models.IsLike(postID, userID) {
			models.UpdateLike(postID, userID, "-1")

		} else {
			models.AddLike(postID, userID, "-1")
		}
	}

	http.Redirect(w, r, "/Post?id="+postID, http.StatusSeeOther)
}

//-----------------------------------------------------------------------

func LikeCommentHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := GetUserIDFromSession(r)
	commentID := r.URL.Query().Get("Comment_id")
	like := r.URL.Query().Get("like") // "1" for like, "0" for dislike
	postID := r.URL.Query().Get("post_id")

	// Logic to update the like/dislike in the database
	if commentID == "" || (like != "1" && like != "-1") {
		http.Error(w, "Bad request: Missing PostID or Comment", http.StatusBadRequest) // 400 Bad Request
		return
	}

	if like == "1" {
		if models.CommentIsLike(commentID, userID) {
			models.CommentRemoveLike(commentID, userID)
			
		} else if models.CommentIsDisLike(commentID, userID) {
			models.CommentUpdateLike(commentID, userID, "1")

		} else {
			models.CommentAddLike(commentID, userID, "1")
		}
	} else if like == "-1" {
		if models.CommentIsDisLike(commentID, userID) {
			models.CommentRemoveLike(commentID, userID)

		} else if models.CommentIsLike(commentID, userID) {
			models.CommentUpdateLike(commentID, userID, "-1")

		} else {
			models.CommentAddLike(commentID, userID, "-1")
		}
	}

	http.Redirect(w, r, "/Post?id="+postID, http.StatusSeeOther)


}

//-----------------------------------------------------------------------

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Destroy the session
	DestroySession(w, r)
	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

//-----------------------------------------------------------------------


func CommentHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := GetUserIDFromSession(r)

	// Extract form values
	postId := r.FormValue("PostID")
	comment := r.FormValue("PostComment")

	// Check if required fields are present
	if postId == "" || comment == "" {
		http.Error(w, "Bad request: Missing PostID or Comment", http.StatusBadRequest) // 400
		return
	}

	// Attempt to create comment
	err := models.CreateComment(userID, postId, comment)
	if err != nil {
		http.Error(w, "Internal server error 500", http.StatusInternalServerError) // 500
		RenderTemplate(w, "500", nil)  
		return
	}

	// Redirect to the post page after successful comment creation
	http.Redirect(w, r, "/Post?id="+postId, http.StatusFound)
}

//-----------------------------------------------------------------------
