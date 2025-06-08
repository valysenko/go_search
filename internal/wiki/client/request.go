package client

import (
	"net/url"
)

// Category Members Request

type GetCategoryMembersRequest struct {
	Action     string `json:"action"`
	List       string `json:"list"`
	CmTitle    string `json:"cmtitle"`
	CmLimit    string `json:"cmlimit"`
	Format     string `json:"json"`
	CmContinue string `json:"cmcontinue,omitempty"`
}

func NewGetCategoryMembersRequest(category string, cmContinue string) *GetCategoryMembersRequest {
	return &GetCategoryMembersRequest{
		Action:     "query",
		List:       "categorymembers",
		CmTitle:    "Category:" + category,
		CmLimit:    "50",
		Format:     "json",
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
	Format          string `json:"json"`
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
	Generator       string `json:"list"`
	GcmTitle        string `json:"gcmtitle"`
	GcmType         string `json:"page"`
	GcmLimit        string `json:"gcmlimit"`
	Prop            string `json:"prop"`
	Eplaintext      string `json:"explaintext"`
	Format          string `json:"json"`
	Exsectionformat string `json:"exsectionformat"`
	GcmContinue     string `json:"gcmcontinue,omitempty"`
}

func NewGetCategoryMembersWithGeneratorRequest(category string) *GetCategoryMembersWithGeneratorRequest {
	return &GetCategoryMembersWithGeneratorRequest{
		Action:          "query",
		Generator:       "categorymembers",
		GcmTitle:        "Category:" + category,
		GcmType:         "page",
		GcmLimit:        "1",
		Prop:            "extracts",
		Eplaintext:      "true",
		Format:          "json",
		Exsectionformat: "plain",
		GcmContinue:     "",
	}
}

func (cmreq *GetCategoryMembersWithGeneratorRequest) UrlValues() url.Values {
	params := url.Values{}
	params.Set("action", cmreq.Action)
	params.Set("generator", cmreq.Generator)
	params.Set("gcmtitle", cmreq.GcmTitle)
	params.Set("gcmtype", cmreq.GcmType)
	params.Set("gcmlimit", cmreq.GcmLimit)
	params.Set("prop", cmreq.Prop)
	params.Set("explaintext", cmreq.Eplaintext)
	params.Set("format", cmreq.Format)
	params.Set("exsectionformat", cmreq.Exsectionformat)
	if cmreq.GcmContinue != "" {
		params.Set("gcmcontinue", cmreq.GcmContinue)
	}
	return params
}
