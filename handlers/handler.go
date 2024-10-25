package handlers

import (
	"Forum/models"
	"html/template"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

// renderTemplate helper function
func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		log.Print(err)
		http.Error(w, "Internal server error 500", http.StatusInternalServerError)
	}
}

// BaseHandler serves pages with the base layout (base.html)
func HomeHandler(w http.ResponseWriter, r *http.Request) {
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
	var postDetails []map[string]interface{}
	for _, Catagory := range Catagories {
		postDetail := map[string]interface{}{
			"Catagory": Catagory.Name,
		}
		postDetails = append(postDetails, postDetail)
	}
	pageData["Catagories"] = postDetails

	// Render the template with base.html as the layout
	err1 := templates.ExecuteTemplate(w, "base.html", pageData)
	if err1 != nil {
		log.Print(err)
		http.Error(w, "Internal server error 500", http.StatusInternalServerError)
	}

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
	} else if r.Method == http.MethodPost {
		Email_UserName := r.FormValue("email")
		password := r.FormValue("password")
		user, err := models.GetUserByEmail(Email_UserName)
		if err != nil {
			user, err = models.GetUserByUserName(Email_UserName)
		}
		if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
			http.Error(w, "Invalid login", http.StatusUnauthorized)
			return
		}
		CreateSession(w, user.Username)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

//-----------------------------------------------------------------------

// func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
// 	userID, loggedIn := GetUserIDFromSession(r)
// 	if !loggedIn {
// 		http.Redirect(w, r, "/login", http.StatusSeeOther)
// 		return
// 	}
// 	if r.Method == http.MethodGet {
// 		RenderTemplate(w, "createPost", nil)
// 	} else if r.Method == http.MethodPost {

// 		title := r.FormValue("title")

// 		content := r.FormValue("content")

// 		categories := r.Form["categories[]"]

// 		stringCategories := strings.Join(categories, ",")

//			err := models.CreatePost(userID, title, content, stringCategories)
//			if err != nil {
//				http.Error(w, "Unable to create post", http.StatusInternalServerError)
//				return
//			}
//			http.Redirect(w, r, "/", http.StatusSeeOther)
//		}
//	}
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	userID, loggedIn := GetUserIDFromSession(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther) // 303
		return
	}

	if r.Method == http.MethodGet {
		RenderTemplate(w, "createPost", nil) // Render form for creating a post
	} else if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")
		categories := r.Form["categories[]"]
		stringCategories := strings.Join(categories, ",")

		// Check if title and content are provided
		if title == "" || content == "" {
			http.Error(w, "Bad request: Missing title or content", http.StatusBadRequest) // 400
			return
		}

		// if len(categories) == 0 { // Check if categories are selected
		// 	http.Error(w, "Bad request: At least one category must be selected", http.StatusBadRequest) // 400
		// 	return
		// }

		// Attempt to create the post
		err := models.CreatePost(userID, title, content, stringCategories)
		if err != nil {
			http.Error(w, "Unable to create post", http.StatusInternalServerError) // 500
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther) // 303
	}
}

//-----------------------------------------------------------------------

//deleted the original cbecause i changed nothing just added spaces

func CreatedPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID, loggedIn := GetUserIDFromSession(r)

	// Check if user is logged in
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther) // 303
		return
	}

	// Get posts for the logged-in user
	posts, err := models.GetPostsFromUserID(userID)
	if err != nil {
		http.Error(w, "Unable to load posts", http.StatusInternalServerError) // 500
		return
	}

	// Render the template with posts
	data := struct {
		Title string
		Posts []models.Post
	}{
		Title: "My Created Posts",
		Posts: posts,
	}

	RenderTemplate(w, "myposts", data) // Render the "myposts" template
}

//-----------------------------------------------------------------------

// func ViewPostHandler(w http.ResponseWriter, r *http.Request) {
// 		_, isLoggedIn := GetUserIDFromSession(r)
// 		isExist := true
// 		id := r.URL.Query().Get("id")

// 	post, err0 := models.GetPostByID(id)
// 	Comments , err := models.GetCommentsByPostID(id)
// 	if err0 != nil {
// 		w.WriteHeader(http.StatusNotFound) // Set the 404 status code
// 		RenderTemplate(w, "404", nil)      // Render custom 404 page
// 		return
// 	}
// 	// comments, err := models.GetCommentsByPostID(postID)
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError) // Set the 500 status code
// 		RenderTemplate(w, "500", nil)                 // Render custom 500 page
// 		return
// 	}

// 	var CommentDetails []map[string]interface{}

// 	for _, comment := range Comments {

// 		CommentDetail := map[string]interface{}{

// 			"Author":  comment.Author,
// 			"comment":   comment.Content,
// 			"created_at":comment.Created_at,
// 			"PostUserID": post.UserID,
// 			"CommentUserID": comment.User_ID,

// 		}
// 		CommentDetails = append(CommentDetails, CommentDetail)
// 	}
// 	pageData := make(map[string]interface{})
// 	pageData["id"] = id
// 	pageData["Author"]= post.Author
// 	pageData["Title"]= post.Title
// 	pageData["Content"]= post.Content
// 	pageData["IsLoggedIn"]= isLoggedIn
// 	pageData["isExist"] = isExist
// 	pageData["Comments"] = CommentDetails
// 	pageData["likes"]	= post.Likes
// 	pageData["DisLikes"] = post.Dislikes

//				err1 := templates.ExecuteTemplate(w, "viewPost.html", pageData)
//				if err1 != nil {
//					http.Error(w, "Internal server error 500", http.StatusInternalServerError)
//				}
//	}
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
		commentDetail := map[string]interface{}{
			"Author":        comment.Author,
			"comment":       comment.Content,
			"created_at":    comment.Created_at,
			"PostUserID":    post.UserID,
			"CommentUserID": comment.User_ID,
		}
		CommentDetails = append(CommentDetails, commentDetail)
	}

	// Prepare page data with post details and comments
	pageData := make(map[string]interface{})
	pageData["id"] = id
	pageData["Author"] = post.Author
	pageData["Title"] = post.Title
	pageData["Content"] = post.Content
	pageData["IsLoggedIn"] = isLoggedIn
	pageData["isExist"] = isExist
	pageData["Comments"] = CommentDetails
	pageData["likes"] = post.Likes
	pageData["DisLikes"] = post.Dislikes

	// Render the view post template
	err1 := templates.ExecuteTemplate(w, "viewPost.html", pageData)
	if err1 != nil {
		http.Error(w, "Internal server error 500", http.StatusInternalServerError) // 500
	}
}

//-----------------------------------------------------------------------

// func CatagoryHandler(w http.ResponseWriter, r *http.Request) {
// 	_, isLoggedIn := GetUserIDFromSession(r)
// 	Catagory := r.FormValue("Catagory")
// 	isExist := true
// 	posts, err := models.GetAllCategoryPosts(Catagory)
// 	if err != nil {
// 		log.Print(err)
// 		w.WriteHeader(http.StatusNotFound) // Set the 404 status code
// 		RenderTemplate(w, "404", nil)      // Render custom 404 page
// 		return
// 	}
// 	// comments, err := models.GetCommentsByPostID(postID)
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError) // Set the 500 status code
// 		RenderTemplate(w, "500", nil)                 // Render custom 500 page
// 		return
// 	}

// 	pageData := make(map[string]interface{})

// 	if posts == nil {
// 		isExist = false
// 	}
// 	// Create a slice to hold the post details for the template
// 	var postDetails []map[string]interface{}
// 	for _, post := range posts {

// 		postDetail := map[string]interface{}{
// 			"IsLoggedIn": isLoggedIn,
// 			"Id"	:  post.ID,
// 			"Author":  post.Author,
// 			"Title":   post.Title,
// 			"created_at":post.Created_at,

// 		}
// 		postDetails = append(postDetails, postDetail)
// 	}

// 	pageData["IsLoggedIn"]= isLoggedIn
// 	pageData["Posts"] = postDetails
// 	pageData["isExist"] = isExist
// 	pageData["CateName"] = Catagory

//		err1 := templates.ExecuteTemplate(w, "CategoryViewer.html", pageData)
//		if err1 != nil {
//			http.Error(w, "Internal server error 500", http.StatusInternalServerError)
//		}
//	}
func CatagoryHandler(w http.ResponseWriter, r *http.Request) {
	_, isLoggedIn := GetUserIDFromSession(r)
	catagory := r.FormValue("Catagory")
	isExist := true

	// Retrieve all posts in the category
	posts, err := models.GetAllCategoryPosts(catagory)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound) // 404
		RenderTemplate(w, "404", nil)      // Render custom 404 page for category not found
		return
	}

	// comments, err := models.GetCommentsByPostID(postID)
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
	pageData["CateName"] = catagory

	// Render the category view template
	err1 := templates.ExecuteTemplate(w, "CategoryViewer.html", pageData)
	if err1 != nil {
		http.Error(w, "Internal server error 500", http.StatusInternalServerError) // 500
	}
}

//-----------------------------------------------------------------------
// Set a flag for cooldown period

// A simple map to track user actions; for real applications, consider a more persistent method.
var userActionStatus = make(map[string]bool)

func LikeHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := GetUserIDFromSession(r)
	postID := r.URL.Query().Get("post_id")
	like := r.URL.Query().Get("like") // "1" for like, "-1" for dislike

	// Validate inputs
	if postID == "" || (like != "1" && like != "-1") {
		http.Redirect(w, r, "/Post?id="+postID, http.StatusSeeOther)
		return
	}

	// Check if the user has recently performed an action
	if userActionStatus[userID] {
		http.Redirect(w, r, "/Post?id="+postID, http.StatusSeeOther)
		return
	}

	// Mark the user as having performed an action
	userActionStatus[userID] = true

	// Process like/dislike logic
	if like == "1" {
		if models.IsLike(postID, userID) {
			models.RemoveLike(postID, userID)
			models.DecraseLike(postID)
		} else if models.IsDisLike(postID, userID) {
			models.RemoveLike(postID, userID) // Remove the dislike first
			models.DecraseDisLike(postID)     // Decrease dislike count
			models.IncreaseLike(postID)        // Increase like count
			models.UpdateLike(postID, userID, "1")
		} else {
			models.IncreaseLike(postID)
			models.AddLike(postID, userID, "1")
		}
	} else if like == "-1" {
		if models.IsDisLike(postID, userID) {
			models.RemoveLike(postID, userID)
			models.DecraseDisLike(postID)
		} else if models.IsLike(postID, userID) {
			models.RemoveLike(postID, userID) // Remove the like first
			models.DecraseLike(postID)        // Decrease like count
			models.IncreaseDisLike(postID)    // Increase dislike count
			models.UpdateLike(postID, userID, "-1")
		} else {
			models.IncreaseDisLike(postID)
			models.AddLike(postID, userID, "-1")
		}
	}

	// Redirect to the post page after action
	http.Redirect(w, r, "/Post?id="+postID, http.StatusSeeOther)
}

// Reset the action status on the next page load
func ResetUserAction(w http.ResponseWriter, r *http.Request) {
	userID, _ := GetUserIDFromSession(r)
	delete(userActionStatus, userID)
}
// func LikeHandler(w http.ResponseWriter, r *http.Request) {
// 	userID, _ := GetUserIDFromSession(r)
// 	postID := r.URL.Query().Get("post_id")
// 	like := r.URL.Query().Get("like") // "1" for like, "0" for dislike

// 	// Logic to update the like/dislike in the database
// 	if postID == "" || (like != "1" && like != "-1") {
// 		http.Error(w, "Invalid like or post ID", http.StatusBadRequest) // 400 Bad Request
// 		return
// 	}

// 	if like == "1" {
// 		if models.IsLike(postID, userID) {
// 			models.RemoveLike(postID, userID)
// 			models.DecraseLike(postID)

// 		} else if models.IsDisLike(postID, userID) {
// 			models.DecraseDisLike(postID)
// 			models.IncreaseLike(postID)
// 			models.UpdateLike(postID, userID, "1")

// 		} else {
// 			models.IncreaseLike(postID)
// 			models.AddLike(postID, userID, "1")
// 		}
// 	} else if like == "-1" {
// 		if models.IsDisLike(postID, userID) {
// 			models.RemoveLike(postID, userID)
// 			models.DecraseDisLike(postID)

// 		} else if models.IsLike(postID, userID) {
// 			models.DecraseLike(postID)
// 			models.IncreaseDisLike(postID)
// 			models.UpdateLike(postID, userID, "-1")

// 		} else {
// 			models.IncreaseDisLike(postID)
// 			models.AddLike(postID, userID, "-1")
// 		}
// 	}

// 	http.Redirect(w, r, "/Post?id="+postID, http.StatusSeeOther)
// }

//-----------------------------------------------------------------------

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Destroy the session
	DestroySession(w, r)
	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

//-----------------------------------------------------------------------

// func CommentHandler(w http.ResponseWriter, r *http.Request){
// 	userID, _ := GetUserIDFromSession(r)

// 	postId := r.FormValue("PostID")
// 	Comment := r.FormValue("PostComment")

// 	err := models.CreateComment(userID,postId,Comment)
// 	if err != nil {
// 		http.Error(w, "Internal server error 500", http.StatusInternalServerError)
// 	}

// 	http.Redirect(w, r, "/Post?id="+postId, http.StatusFound)
// }

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
		return
	}

	// Redirect to the post page after successful comment creation
	http.Redirect(w, r, "/Post?id="+postId, http.StatusFound)
}

//-----------------------------------------------------------------------
