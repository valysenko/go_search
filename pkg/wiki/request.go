package wiki

import (
	"net/url"
)

// Category Members Request

type GetCategoryMembersRequest struct {
	Action        string `json:"action"`
	List          string `json:"list"`
	CmTitle       string `json:"cmtitle"`
	CmLimit       string `json:"cmlimit"`
	Format        string `json:"format"`
	Formatversion string `json:"formatversion"`
	CmSort        string `json:"cmsort"`
	CmDir         string `json:"cmdir"`
	CmProp        string `json:"cmprop"`
	CmContinue    string `json:"cmcontinue,omitempty"`
}

func NewGetCategoryMembersRequest(category string, cmContinue string) *GetCategoryMembersRequest {
	return &GetCategoryMembersRequest{
		Action:     "query",
		List:       "categorymembers",
		CmTitle:    "Category:" + category,
		CmLimit:    "10",
		Format:     "json",
		CmSort:     "timestamp",
		CmDir:      "desc",
		CmProp:     "ids|title|timestamp",
		CmContinue: cmContinue,
	}
}

func (cmreq *GetCategoryMembersRequest) UrlValues() url.Values {
	params := url.Values{}
	params.Set("action", cmreq.Action)
	params.Set("list", cmreq.List)
	params.Set("cmtitle", cmreq.CmTitle)
	params.Set("cmlimit", cmreq.CmLimit)
	params.Set("format", cmreq.Format)
	params.Set("cmsort", cmreq.CmSort)
	params.Set("cmdir", cmreq.CmDir)
	params.Set("cmprop", cmreq.CmProp)
	if cmreq.CmContinue != "" {
		params.Set("cmcontinue", cmreq.CmContinue)
	}
	return params
}

// Atricle Content Request

type GetArticleContentRequest struct {
	Action          string `json:"action"`
	Prop            string `json:"prop"`
	PageIds         string `json:"pageids"`
	Eplaintext      string `json:"explaintext"`
	Format          string `json:"format"`
	Exsectionformat string `json:"exsectionformat"`
}

func NewGetArticleContentRequest(pageId string) *GetArticleContentRequest {
	return &GetArticleContentRequest{
		Action:          "query",
		Prop:            "extracts",
		PageIds:         pageId, // use in batch 1|2|3
		Eplaintext:      "true",
		Format:          "json",
		Exsectionformat: "plain",
	}
}

func (areq *GetArticleContentRequest) UrlValues() url.Values {
	params := url.Values{}
	params.Set("action", areq.Action)
	params.Set("prop", areq.Prop)
	params.Set("pageids", areq.PageIds)
	params.Set("explaintext", areq.Eplaintext)
	params.Set("format", areq.Format)
	params.Set("exsectionformat", areq.Exsectionformat)
	return params
}

// Category Members With Article Content Generator Request

type GetCategoryMembersWithGeneratorRequest struct {
	Action          string `json:"action"`
	Generator       string `json:"generator"`
	GcmTitle        string `json:"gcmtitle"`
	GcmType         string `json:"page"`
	GcmLimit        string `json:"gcmlimit"`
	GcmSort         string `json:"gcmsort"`
	GcmDir          string `json:"gcmdir"`
	Prop            string `json:"prop"`
	Eplaintext      string `json:"explaintext"`
	Format          string `json:"json"`
	Exsectionformat string `json:"exsectionformat,omitempty"`
	RvProp          string `json:"rvprop,omitempty"`
	RvLimit         string `json:"rvlimit,omitempty"`
	RvDir           string `json:"rvdir,omitempty"`
	GcmContinue     string `json:"gcmcontinue,omitempty"`
}

func NewGetCategoryMembersWithGeneratorRequest(category string) *GetCategoryMembersWithGeneratorRequest {
	return &GetCategoryMembersWithGeneratorRequest{
		Action:          "query",
		Generator:       "categorymembers",
		GcmTitle:        "Category:" + category,
		GcmType:         "page",
		GcmLimit:        "1",
		GcmSort:         "timestamp",
		GcmDir:          "desc",
		Prop:            "extracts|revisions",
		Eplaintext:      "true",
		Format:          "json",
		Exsectionformat: "plain",
		RvProp:          "timestamp",
		RvLimit:         "1",
		RvDir:           "older",
		GcmContinue:     "",
	}
}

func (cmreq *GetCategoryMembersWithGeneratorRequest) UrlValues() url.Values {
	params := url.Values{}
	params.Set("action", cmreq.Action)
	params.Set("generator", cmreq.Generator)
	params.Set("gcmtitle", cmreq.GcmTitle)
	params.Set("gcmtype", cmreq.GcmType)
	params.Set("gcmsort", cmreq.GcmSort)
	params.Set("gcmdir", cmreq.GcmDir)
	params.Set("gcmlimit", cmreq.GcmLimit)
	params.Set("prop", cmreq.Prop)
	params.Set("explaintext", cmreq.Eplaintext)
	params.Set("format", cmreq.Format)
	params.Set("exsectionformat", cmreq.Exsectionformat)
	params.Set("rvprop", cmreq.RvProp)
	params.Set("rvlimit", cmreq.RvLimit)
	params.Set("rvdir", cmreq.RvDir)
	if cmreq.GcmContinue != "" {
		params.Set("gcmcontinue", cmreq.GcmContinue)
	}
	return params
}
