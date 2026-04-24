package phgraphql

import "time"

// PostNode is the shape of one edge.node in a BackfillPosts or EnrichPosts
// response. Field names are JSON-tag-driven and match the GraphQL schema.
type PostNode struct {
	ID         string    `json:"id"`
	Slug       string    `json:"slug"`
	Name       string    `json:"name"`
	Tagline    string    `json:"tagline"`
	URL        string    `json:"url"`
	Website    string    `json:"website"`
	CreatedAt  time.Time `json:"createdAt"`
	VotesCount int       `json:"votesCount"`
	User       PostUser  `json:"user"`
}

// PostUser is the nested user field. PH may return null for some posts; the
// struct's zero value is harmless.
type PostUser struct {
	Name     string `json:"name"`
	Username string `json:"username"`
}

// PageInfo mirrors the GraphQL Relay-style pagination envelope.
type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

// PostsPage is the decoded `posts` field from a BackfillPosts response.
type PostsPage struct {
	PageInfo PageInfo     `json:"pageInfo"`
	Edges    []PostsEdge  `json:"edges"`
}

// PostsEdge is one Relay edge. We carry only Node; Cursor lives in PageInfo.EndCursor.
type PostsEdge struct {
	Node PostNode `json:"node"`
}
