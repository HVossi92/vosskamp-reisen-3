package server

import (
	"html/template"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"vosskamp-reisen-3/internal/helpers"
	"vosskamp-reisen-3/internal/models"

	quill "github.com/dchenk/go-render-quill"
	"github.com/google/uuid"
)

func (s *Server) RegisterPostRoutes(mux *http.ServeMux) {
	mux.Handle("GET /admin/posts", s.middleWareService.CheckSession(http.HandlerFunc(s.getPostsPageHandler)))
	mux.Handle("GET /admin/posts/rows", s.middleWareService.CheckSession(http.HandlerFunc(s.fetchPostsRows)))
	mux.Handle("GET /admin/post", s.middleWareService.CheckSession(http.HandlerFunc(s.fetchPost)))
	mux.Handle("GET /admin/post/create", s.middleWareService.CheckSession(http.HandlerFunc(s.fetchCreatePostFormHandler)))
	mux.Handle("GET /admin/post/update", s.middleWareService.CheckSession(http.HandlerFunc(s.fetchUpdatePostFormHandler)))
	mux.Handle("POST /admin/post", s.middleWareService.CheckSession(http.HandlerFunc(s.createPostHandler)))
	mux.Handle("PUT /admin/post", s.middleWareService.CheckSession(http.HandlerFunc(s.updatePostHandler)))
	mux.Handle("DELETE /admin/post", s.middleWareService.CheckSession(http.HandlerFunc(s.deletePostHandler)))
}

func (s *Server) getPostsPageHandler(w http.ResponseWriter, r *http.Request) {
	err := s.tmpl.ExecuteTemplate(w, "posts", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) createPostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	post := models.Posts{
		Title: r.FormValue("title"),
		Body:  r.FormValue("body"),
	}

	var errorMessages = map[string]bool{
		"Title": false,
		"Body":  false,
	}
	hasErrors := false
	if post.Title == "" {
		errorMessages["Title"] = true
		hasErrors = true
	}
	if post.Body == "" {
		errorMessages["Body"] = true
		hasErrors = true
	}

	if hasErrors {
		errorFormData := map[string]interface{}{
			"ErrorMessages": errorMessages,
			"Post":          post,
		}
		err := s.tmpl.ExecuteTemplate(w, "postForm", errorFormData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	post.Picture, err = writeImage(w, r, err, post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = s.postService.CreatePost(post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Location", "/admin/posts")
	w.WriteHeader((http.StatusNoContent))
}

func (s *Server) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	postIdStr := r.URL.Query().Get("id")
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	post := models.Posts{
		Title: r.FormValue("title"),
		Body:  r.FormValue("body"),
	}

	var errorMessages = map[string]bool{
		"Title": false,
		"Body":  false,
	}
	hasErrors := false
	if post.Title == "" {
		errorMessages["Title"] = true
		hasErrors = true
	}
	if post.Body == "" {
		errorMessages["Body"] = true
		hasErrors = true
	}

	if hasErrors {
		errorFormData := map[string]interface{}{
			"ErrorMessages": errorMessages,
			"Post":          post,
		}
		err = s.tmpl.ExecuteTemplate(w, "postForm", errorFormData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	existingPost, err := s.postService.FetchPostById(postId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if existingPost.Picture != "" && post.Picture != "" {
		err := removePicture(w, err, existingPost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	post.Picture, err = writeImage(w, r, err, post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	post.Id = postId
	if post.Picture == "" {
		post.Picture = existingPost.Picture
	}

	_, err = s.postService.UpdatePost(post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Location", "/admin/posts")
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	postIdStr := r.URL.Query().Get("id")
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	existingPost, err := s.postService.FetchPostById(postId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if existingPost.Picture != "" {
		err := removePicture(w, err, existingPost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	err = s.postService.DeletePost(postId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Location", "/admin/posts")
	w.WriteHeader(200)
}

func (s *Server) fetchCreatePostFormHandler(w http.ResponseWriter, r *http.Request) {
	err := s.tmpl.ExecuteTemplate(w, "postForm", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) fetchUpdatePostFormHandler(w http.ResponseWriter, r *http.Request) {
	postIdStr := r.URL.Query().Get("id")
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	post, err := s.postService.FetchPostById(postId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Post          *models.Posts
		ErrorMessages map[string]bool
		IsUpdate      bool
	}{
		Post: post,
		ErrorMessages: map[string]bool{
			"Title": false,
			"Body":  false,
		},
		IsUpdate: true,
	}
	err = s.tmpl.ExecuteTemplate(w, "postForm", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) fetchPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	post, err := s.postService.FetchPostById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	start := strings.Index(post.Body, `"ops":[`) + len(`"ops":[`)
	end := strings.Index(post.Body[start:], `]`) + start

	// Extract the ops array
	opsArray := "[" + post.Body[start:end] + "]"

	html, err := quill.Render([]byte(opsArray))
	if err != nil {
		panic(err)
	}

	data := struct {
		Id        int
		Title     string
		Body      template.HTML
		CreatedAt string
		UpdatedAt string
		Picture   string
	}{
		Id:        post.Id,
		Title:     post.Title,
		Body:      template.HTML(html),
		CreatedAt: post.CreatedAt,
		Picture:   post.Picture,
	}

	err = s.tmpl.ExecuteTemplate(w, "post", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) fetchPostsRows(w http.ResponseWriter, r *http.Request) {
	page, limit := helpers.GetPagination(r)
	posts, totalPosts, err := s.postService.FetchPaginatedPosts(page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(totalPosts) / float64(limit)))
	data := struct {
		Posts            *[]models.Posts
		CurrentPage      int
		TotalPages       int
		Limit            int
		PreviousPage     int
		NextPage         int
		PageButtonsRange []int
	}{
		Posts:            posts,
		CurrentPage:      page,
		TotalPages:       totalPages,
		Limit:            limit,
		PreviousPage:     page - 1,
		NextPage:         page + 1,
		PageButtonsRange: makeRange(1, totalPages),
	}

	err = s.tmpl.ExecuteTemplate(w, "postRows", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func writeImage(w http.ResponseWriter, r *http.Request, err error, post models.Posts) (string, error) {
	file, handler, err := r.FormFile("picture")
	if err == nil {
		// Generate a unique filename to prevent overwriting and conflicts
		random, err := uuid.NewRandom()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return "", err
		}
		filename := random.String() + filepath.Ext(handler.Filename) // Append the file extension

		// Create the full path for saving the file
		filePath := filepath.Join("internal/static/uploads", filename)

		// Save the file to the server
		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return "", err
		}
		defer func(dst *os.File) {
			err := dst.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}(dst)

		if _, err = io.Copy(dst, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return "", err
		}
		post.Picture = filename
	}
	return post.Picture, nil
}

func removePicture(w http.ResponseWriter, err error, existingPost *models.Posts) error {
	err = os.Remove(filepath.Join("internal/static/uploads", existingPost.Picture))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	return err
}
