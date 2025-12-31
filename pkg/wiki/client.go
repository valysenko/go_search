package wiki

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// https://wikitech.wikimedia.org/wiki/Robot_policy
const UserAgentHeader = "User-Agent"
const UserAgent = "MyGoSearchTestIndexer/1.0 (https://example.com; contact@example.com)"

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
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
func (wc *WikiClient) GetCategoryMembers(request *GetCategoryMembersRequest) (*CategoryMembersResponse, error) {
	params := request.UrlValues()
	req, err := http.NewRequest("GET", wc.apiUrl+"?"+params.Encode(), nil)
	fmt.Println(wc.apiUrl + "?" + params.Encode())
	req.Header.Set(UserAgentHeader, UserAgent)
	if err != nil {
		return nil, err
	}

	resp, err := wc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CategoryMembersResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// https://www.mediawiki.org/wiki/API:Extracts
// https://www.mediawiki.org/wiki/API:Pageids
// https://www.mediawiki.org/wiki/Extension:TextExtracts#API
func (wc *WikiClient) GetArticleContent(request *GetArticleContentRequest) (*ArticleResponse, error) {
	params := request.UrlValues()
	req, err := http.NewRequest("GET", wc.apiUrl+"?"+params.Encode(), nil)
	req.Header.Set(UserAgentHeader, UserAgent)
	if err != nil {
		return nil, err
	}

	resp, err := wc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ArticleResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)

	return &result, nil
}

// method is not appropriate
// - cmprop parameter from categorymembers is not supported when using generator
// - possible to include revisions array and use first revision instead
// - BUT not possible to fetch multiple pages with 'rvdir', 'rwlimit' params - last revision timestamp returned
// - not possible to fetch multiple pages. warning is: "extracts": {"*": "\"exlimit\" was too large for a whole article extracts request, lowered to 1."}
// - there is also problem with "continue" - rvcontinue in response instead of cmcontinue
// https://www.mediawiki.org/wiki/API:Query#Generators
// https://www.mediawiki.org/wiki/Extension:TextExtracts
// https://www.mediawiki.org/wiki/API:Revisions
func (wc *WikiClient) GetAllCategoryMembersWithPageContent(request *GetCategoryMembersWithGeneratorRequest) (*CategoryMembersWithExtractResponse, error) {
	params := request.UrlValues()

	req, err := http.NewRequest("GET", wc.apiUrl+"?"+params.Encode(), nil)
	fmt.Println(wc.apiUrl + "?" + params.Encode())
	req.Header.Set(UserAgentHeader, UserAgent)
	if err != nil {
		return nil, err
	}

	resp, err := wc.client.Do(req)
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

	return &result, nil
}
