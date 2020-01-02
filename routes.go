package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"photos-app/models"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const userKey string = "user"

func registerRoutes() *gin.Engine {
	gin.SetMode("release")
	r := gin.Default()
	store := cookie.NewStore([]byte("viErkShjgQP59tgelRXsILXNEarwRA6p"))
	store.Options(sessions.Options{
		MaxAge: 0,
	})
	r.Static("img", "./img")
	r.Use(sessions.Sessions("photos-session", store))
	// r.NoRoute(noroute) // sets a redirect for non-existent routes

	r.GET("/", home)
	r.GET("/ping", ping)
	r.GET("/error", showError)

	r.GET("/signon", signinForm)
	r.POST("/signon", con.signin)
	r.GET("/signoff", signoff, home)

	r.GET("/register", registerForm)
	r.POST("/register", con.register)
	r.GET("/welcome", welcome)
	r.GET("/profile/:user", con.profile)
	r.GET("/view/:id", con.showPost)

	r.POST("/comment", con.comment)

	r.GET("/upload", uploadForm)
	r.POST("/upload", con.uploadPhoto)

	return r
}

func getUserSession(c *gin.Context) gin.H {
	h := gin.H{}
	s := sessions.Default(c)
	if s.Get(userKey) != nil {
		data, err := json.Marshal(s.Get(userKey))
		if err != nil {
			log.Fatal("Failed to marshal to json")
		}
		u := models.User{}
		if err := json.Unmarshal(data, &u); err != nil {
			log.Fatal("Failed to unmarshal to struct")
		}
		h[userKey] = u
	}

	if s.Get("error") != nil {
		h["error"] = s.Get("error")
		s.Delete("error")
		s.Save()
	}

	return h
}

func (con appController) comment(c *gin.Context) {

	h := getUserSession(c)
	if h[userKey] == nil {
		c.Redirect(302, "/signon")
		return
	}

	comment := strings.TrimSpace(c.Request.FormValue("comment"))
	postID := strings.TrimSpace(c.Request.FormValue("postID"))

	if comment == "" {
		h["error"] = "Field cannot be blank"
		s := sessions.Default(c)
		s.Set("error", "Field cannot be blank")
		s.Save()
		c.Redirect(301, "/view/"+postID)
		return
	}

	post, err := con.Mongo.FindPostByID(postID)
	if err != nil {
		log.Printf("That post could not be found, %v\n", err)
		c.Redirect(302, "/error")
		return
	}

	com := models.Comment{
		CreatedBy: h[userKey].(models.User).Username,
		Content:   comment,
	}
	post.Comments = append(post.Comments, com)

	owner, err := con.Mongo.FindUserByUsername(post.CreatedBy)
	if err != nil {
		log.Printf("That user could not be found, %v\n", err)
		c.Redirect(302, "/error")
		return
	}

	newPosts := owner.Posts

	for i, p := range owner.Posts {
		if p.PostID == post.PostID {
			newPosts[i] = post
			break
		}
	}

	owner.Posts = newPosts
	if err := con.Mongo.Upsert(owner); err != nil {
		c.Redirect(302, "/error")
		return
	}

	if err := con.Mongo.Upsert(post); err != nil {
		c.Redirect(302, "/error")
		return
	}

	c.Redirect(302, "/view/"+post.PostID)

}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func noroute(c *gin.Context) {
	c.Redirect(301, "/error")
}

func showError(c *gin.Context) {
	h := getUserSession(c)
	h["title"] = "404 Not Found"
	h["body"] = "Looks like something went wrong :("
	c.HTML(http.StatusOK, "error.html", h)
}

func (con appController) showPost(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	post, err := con.Mongo.FindPostByID(id)
	if err != nil {
		log.Printf("That post could not be found, %v\n", err)
		c.Redirect(302, "/error")
		return
	}

	h := getUserSession(c)
	h["title"] = "View Post"
	h["post"] = post
	c.HTML(http.StatusOK, "post.html", h)
}

func home(c *gin.Context) {
	h := getUserSession(c)
	h["title"] = "Home"
	h["body"] = "Welcome to the Photos App!"
	c.HTML(http.StatusOK, "home.html", h)
}

func signoff(c *gin.Context) {
	s := sessions.Default(c)
	s.Delete(userKey)
	if err := s.Save(); err != nil {
		log.Printf("Failed to save session, %v\n", err)
	}
	c.Next()
}

func registerForm(c *gin.Context) {
	h := getUserSession(c)
	h["title"] = "Register"
	c.HTML(http.StatusOK, "register.html", h)
}

func (con appController) register(c *gin.Context) {
	h := getUserSession(c)
	un := strings.TrimSpace(c.Request.FormValue("username"))
	pw := strings.TrimSpace(c.Request.FormValue("password"))
	pc := strings.TrimSpace(c.Request.FormValue("confirm"))

	if un == "" || pw == "" || pc == "" {
		h["error"] = "Fields cannot be blank"
		c.HTML(http.StatusBadRequest, "register.html", h)
		return
	}

	// TODO: Add failure responses to page render
	if pw != pc {
		h["error"] = "Passwords do not match"
		c.HTML(http.StatusBadRequest, "register.html", h)
		return
	}

	_, err := con.Mongo.FindUserByUsername(un)
	if err == nil {
		h["error"] = "That username has been taken"
		c.HTML(http.StatusFound, "register.html", h)
		return
	}

	u := models.InitUser()

	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), 6)
	if err != nil {
		h["error"] = "Failed to encrypt password"
		c.HTML(http.StatusInternalServerError, "register.html", h)
		return
	}

	u.Username = un
	u.Password = string(hash)

	if err := con.Mongo.Upsert(u); err != nil {
		h["error"] = "Failed to create user"
		c.HTML(http.StatusInternalServerError, "register.html", h)
		return
	}

	s := sessions.Default(c)
	s.Set("user", u)
	s.Save()
	c.Redirect(302, "/welcome")
}

func welcome(c *gin.Context) {
	h := getUserSession(c)
	h["title"] = "Welcome!"
	c.HTML(http.StatusOK, "welcome.html", h)
}

func signinForm(c *gin.Context) {
	h := getUserSession(c)
	if h["user"] != nil {
		c.Redirect(301, "/profile/"+h["user"].(models.User).Username)
		return
	}
	h["title"] = "Sign In Here!"
	c.HTML(http.StatusOK, "login.html", h)
}

func (con appController) signin(c *gin.Context) {
	h := getUserSession(c)
	un := strings.TrimSpace(c.Request.FormValue("username"))
	pw := strings.TrimSpace(c.Request.FormValue("password"))

	if un == "" || pw == "" {
		h["error"] = "Fields cannot be blank"
		c.HTML(http.StatusBadRequest, "login.html", h)
		return
	}

	u, err := con.Mongo.FindUserByUsername(un)
	if err != nil {
		h["error"] = "That user doesn't exist"
		c.HTML(http.StatusNotFound, "login.html", h)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pw)); err != nil {
		h["error"] = "Incorrect Password"
		c.HTML(http.StatusBadRequest, "login.html", h)
		return
	}

	s := sessions.Default(c)
	s.Set(userKey, u)
	s.Save()
	c.Redirect(302, "/profile/"+u.Username)
}

func (con appController) profile(c *gin.Context) {
	un := strings.TrimSpace(c.Param("user"))
	if un == "" {
		log.Println("No User provided")
		c.Redirect(301, "/error")
		return
	}

	u, err := con.Mongo.FindUserByUsername(un)
	if err != nil {
		log.Println("That User doesn't exist")
		c.Redirect(301, "/error")
		return
	}

	h := getUserSession(c)
	h["title"] = un + "'s Profile"
	h["username"] = un
	h["owner"] = u

	c.HTML(http.StatusOK, "profile.html", h)
}

func uploadForm(c *gin.Context) {
	h := getUserSession(c)
	if h[userKey] == nil {
		c.Redirect(302, "/signon")
		return
	}
	h["title"] = "Upload a Photo"
	c.HTML(http.StatusOK, "upload.html", h)
}

func (con appController) uploadPhoto(c *gin.Context) {

	h := getUserSession(c)

	if h[userKey] == nil {
		c.Redirect(302, "/signon")
		return
	}
	userSession := h[userKey].(models.User)

	u, err := con.Mongo.FindUserByUsername(userSession.Username)
	if err != nil {
		c.Redirect(302, "/error")
		return
	}

	file, err := c.FormFile("photo")
	if err != nil {
		h["error"] = "You must select a photo"
		c.HTML(http.StatusBadRequest, "upload.html", h)
		return
	}
	caption := c.Request.FormValue("caption")
	split := strings.Split(file.Filename, ".")
	ext := split[len(split)-1]

	extMatch := false
	for _, a := range []string{"jpg", "jpeg", "png"} {
		if a == ext {
			extMatch = true
			break
		}
	}

	if !extMatch {
		h["error"] = "File must be .jpg, .jpeg, or .png"
		c.HTML(http.StatusBadRequest, "upload.html", h)
		return
	}

	post := models.InitPost()

	// if err := os.MkdirAll("./img/"+u.Username, os.ModePerm); err != nil {
	// 	c.Redirect(302, "/error")
	// 	return
	// }

	// out, err := os.Create("./img/" + u.Username + "/" + post.ImgName)
	// if err != nil {
	// 	log.Printf("Failed to create directory, %v\n", err)
	// 	c.Redirect(302, "/error")
	// 	return
	// }
	// defer out.Close()

	f, err := file.Open()
	if err != nil {
		log.Printf("Failed to open photo, %v\n", err)
		c.Redirect(302, "/error")
		return
	}
	defer f.Close()

	out, err := con.AWS.S3Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("neon-photos"),
		Key:    aws.String(u.Username + "/" + post.PostID),
		Body:   f,
	})
	if err != nil {
		log.Printf("Failed to upload photo to S3 bucket, %v\n", err)
		c.Redirect(302, "/error")
		return
	}

	// log.Printf("S3 Upload Output: %v", out.Location)
	post.CreatedBy = u.Username
	post.ImgName = post.PostID + "." + ext
	post.Caption = caption
	post.CreateDt = time.Now()
	post.ImgLocation = out.Location

	u.Posts = append(u.Posts, post)

	h[userKey] = u

	log.Printf("User Posts: %d", len(u.Posts))

	con.Mongo.Upsert(post)
	con.Mongo.Upsert(u)
	c.HTML(http.StatusCreated, "success.html", h)
}

func uploadSuccess(c *gin.Context) {
	h := getUserSession(c)
	if h[userKey] == nil {
		c.Redirect(302, "/signon")
		return
	}
	h["title"] = "Upload Successful"
	c.HTML(http.StatusOK, "success.html", h)
}

func loadTemplates(tmplDir string) multitemplate.Renderer {
	rend := multitemplate.NewRenderer()

	layouts, err := filepath.Glob(tmplDir + "/layouts/*.html")
	if err != nil {
		log.Panic(err.Error())
	}

	includes, err := filepath.Glob(tmplDir + "/includes/*.html")
	if err != nil {
		log.Panic(err.Error())
	}

	for _, i := range includes {
		lCopy := make([]string, len(layouts))
		copy(lCopy, layouts)
		files := append(lCopy, i)
		rend.AddFromFiles(filepath.Base(i), files...)
	}

	return rend
}
