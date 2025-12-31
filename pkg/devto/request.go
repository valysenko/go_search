package devto

type GetArticlesByTagRequest struct {
	Tag     string
	Page    int
	PerPage int
}

func NewGetArticlesByTagRequest(tag string, page int, perPage int) *GetArticlesByTagRequest {
	return &GetArticlesByTagRequest{
		Tag:     tag,
		Page:    page,
		PerPage: perPage,
	}
}

type GetArticlesByIdRequest struct {
	ID int
}

func NewGetArticlesByIdRequest(id int) *GetArticlesByIdRequest {
	return &GetArticlesByIdRequest{
		ID: id,
	}
}

type GetLatestArticlesRequest struct {
	Page    int
	PerPage int
}

func NewGetLatestArticlesRequest(page int, perPage int) *GetLatestArticlesRequest {
	return &GetLatestArticlesRequest{
		Page:    page,
		PerPage: perPage,
	}
}
