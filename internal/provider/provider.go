package provider

type Source int

const (
	SourceUnknown Source = iota
	SourceDevTo
	SourceHashnode
	SourceWiki
)

type Query struct {
	Tags     []string
	Category string
	TagSlug  string
}
