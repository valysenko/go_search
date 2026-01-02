package wiki

import (
	"context"
	"go_search/pkg/httpclient"
)

// https://wikitech.wikimedia.org/wiki/Robot_policy
const userAgentHeader = "User-Agent"
const userAgent = "MyGoSearchTestIndexer/1.0 (https://example.com; contact@example.com)"
const baseUrl = "https://en.wikipedia.org/w/api.php"

type HTTPClient interface {
	Get(ctx context.Context, path string, headers httpclient.Headers, out any) error
}

type WikiClient struct {
	client HTTPClient
}

// https://www.mediawiki.org/wiki/API:Main_page
func NewWikiClient(timeoutSeconds int) *WikiClient {
	return &WikiClient{
		client: httpclient.NewHttpClient(timeoutSeconds, baseUrl),
	}
}

// https://www.mediawiki.org/wiki/API:Lists
// https://www.mediawiki.org/wiki/API:Categorymembers
// https://www.mediawiki.org/wiki/API:Continue
func (wc *WikiClient) GetCategoryMembers(ctx context.Context, request *GetCategoryMembersRequest) (*CategoryMembersResponse, error) {
	params := request.UrlValues()
	url := "?" + params.Encode()
	headers := httpclient.Headers{
		userAgentHeader: userAgent,
	}

	var result CategoryMembersResponse
	err := wc.client.Get(ctx, url, headers, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// https://www.mediawiki.org/wiki/API:Extracts
// https://www.mediawiki.org/wiki/API:Pageids
// https://www.mediawiki.org/wiki/Extension:TextExtracts#API
func (wc *WikiClient) GetArticleContent(ctx context.Context, request *GetArticleContentRequest) (*ArticleResponse, error) {
	params := request.UrlValues()
	url := "?" + params.Encode()
	headers := httpclient.Headers{
		userAgentHeader: userAgent,
	}

	var result ArticleResponse
	err := wc.client.Get(ctx, url, headers, &result)
	if err != nil {
		return nil, err
	}

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
func (wc *WikiClient) GetAllCategoryMembersWithPageContent(ctx context.Context, request *GetCategoryMembersWithGeneratorRequest) (*CategoryMembersWithExtractResponse, error) {
	params := request.UrlValues()
	url := "?" + params.Encode()
	headers := httpclient.Headers{
		userAgentHeader: userAgent,
	}

	var result CategoryMembersWithExtractResponse
	err := wc.client.Get(ctx, url, headers, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
