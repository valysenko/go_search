package wiki

import "time"

// Category Members Response

type CategoryMembersResponse struct {
	Continue struct {
		CmContinue string `json:"cmcontinue"`
	} `json:"continue"`
	Query struct {
		CategoryMembers []CategoryMember `json:"categorymembers"`
	} `json:"query"`
}

type CategoryMember struct {
	PageID    int       `json:"pageid"`
	Title     string    `json:"title"`
	Timestamp time.Time `json:"timestamp"`
}

// Atricle Content Response

type ArticleResponse struct {
	Query struct {
		Pages map[string]Page `json:"pages"`
	} `json:"query"`
}

type Page struct {
	PageID  int    `json:"pageid"`
	Title   string `json:"title"`
	Extract string `json:"extract"`
}

// Category Members With Article Content Generator Response

type CategoryMembersWithExtractResponse struct {
	BatchComplete bool `json:"batchcomplete"`
	Query         struct {
		Pages map[string]WikiPage `json:"pages"`
	} `json:"query"`
	Continue struct {
		CmContinue string `json:"cmcontinue"`
	} `json:"continue"`
}

type WikiPage struct {
	PageID    int        `json:"pageid"`
	Title     string     `json:"title"`
	Extract   string     `json:"extract"`
	Revisions []Revision `json:"revisions"`
}

type Revision struct {
	Timestamp time.Time `json:"timestamp"`
}
