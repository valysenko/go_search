package client

type ArticleSummary struct {
	ID      int      `json:"id"`
	Title   string   `json:"title"`
	TagList []string `json:"tag_list"`
	Url     string   `json:"url"`
}

type Article struct {
	ID           int      `json:"id"`
	Title        string   `json:"title"`
	BodyHTML     string   `json:"body_html"`
	BodyMarkdown string   `json:"body_markdown"`
	TagList      []string `json:"tag_list"`
	Url          string   `json:"url"`
}
