package client

// Category Members Response

type CategoryMembersResponse struct {
	Continue struct {
		CmContinue string `json:"cmcontinue"`
	} `json:"continue"`
	Query struct {
		CategoryMembers []struct {
			PageID int    `json:"pageid"`
			Title  string `json:"title"`
		} `json:"categorymembers"`
	} `json:"query"`
}

type CatrgoryMember struct {
	PageID int    `json:"pageid"`
	Title  string `json:"title"`
}

// Atricle Content Response

type ArticleResponse struct {
	Query struct {
		Pages map[string]struct {
			Title   string `json:"title"`
			Extract string `json:"extract"`
		} `json:"pages"`
	} `json:"query"`
}

// Category Members With Article Content Generator Response

type CategoryMembersWithExtractResponse struct {
	BatchComplete bool `json:"batchcomplete"`
	Query         struct {
		Pages map[string]WikiPage `json:"pages"`
	} `json:"query"`
	Continue struct {
		GcmContinue string `json:"gcmcontinue"`
	} `json:"continue"`
}

type WikiPage struct {
	PageID  int    `json:"pageid"`
	Title   string `json:"title"`
	Extract string `json:"extract"`
}
