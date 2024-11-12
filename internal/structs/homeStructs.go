package structs

import "vosskamp-reisen-3/internal/models"

type PaginatedData struct {
	CurrentPage      int
	TotalPages       int
	Limit            int
	PreviousPage     int
	NextPage         int
	PageButtonsRange []int
}

type HomePostsData struct {
	Posts *[]models.Posts
	PaginatedData
}
