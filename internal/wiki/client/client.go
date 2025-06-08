package client

import (
	"encoding/json"
	"io"
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type WikiClient struct {
	client HTTPClient
	apiUrl string
}

// https://www.mediawiki.org/wiki/API:Main_page
func NewWikiClient() *WikiClient {
	return &WikiClient{
		client: &http.Client{},
		apiUrl: "https://en.wikipedia.org/w/api.php",
	}
}

// https://www.mediawiki.org/wiki/API:Lists
// https://www.mediawiki.org/wiki/API:Categorymembers
// https://www.mediawiki.org/wiki/API:Continue
func (wc *WikiClient) GetCategoryMembers(request *GetCategoryMembersRequest) ([]*CatrgoryMember, error) {
	var categoryMembers []*CatrgoryMember
	cmContinue := "" // retrieved from response

	for {
		request.CmContinue = cmContinue
		params := request.UrlValues()

		resp, err := http.Get(wc.apiUrl + "?" + params.Encode())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var result CategoryMembersResponse
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &result)

		for _, item := range result.Query.CategoryMembers {
			categoryMembers = append(categoryMembers, &CatrgoryMember{
				Title:  item.Title,
				PageID: item.PageID,
			})
		}

		if result.Continue.CmContinue == "" {
			break
		}
		cmContinue = result.Continue.CmContinue
	}

	return categoryMembers, nil
}

// https://www.mediawiki.org/wiki/API:Extracts
// https://www.mediawiki.org/wiki/API:Pageids
// https://www.mediawiki.org/wiki/Extension:TextExtracts#API
func (wc *WikiClient) GetArticleContent(request *GetArticleContentRequest) (*ArticleResponse, error) {
	params := request.UrlValues()
	resp, err := http.Get(wc.apiUrl + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result ArticleResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)

	return &result, nil
}

// https://www.mediawiki.org/wiki/API:Query#Generators
// https://www.mediawiki.org/wiki/Extension:TextExtracts
func (wc *WikiClient) GetCategoryMembersWithPageContent(request *GetCategoryMembersWithGeneratorRequest) ([]*WikiPage, error) {
	var pages []*WikiPage
	gcmContinue := "" // retrieved from response

	for {
		request.GcmContinue = gcmContinue
		params := request.UrlValues()

		resp, err := http.Get(wc.apiUrl + "?" + params.Encode())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var result CategoryMembersWithExtractResponse
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(body, &result)

		for _, page := range result.Query.Pages {
			pages = append(pages, &page)
		}

		if result.Continue.GcmContinue == "" {
			break
		}

		gcmContinue = result.Continue.GcmContinue
	}

	return pages, nil
}
