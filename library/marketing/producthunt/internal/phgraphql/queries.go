package phgraphql

// GraphQL documents used by the CLI. Kept as Go consts rather than embedded
// .graphql files so `go build` alone suffices — no codegen, no embed tag.
//
// Per the self-warming CLI plan (Key Technical Decisions):
//   * Minimal field set. Each field in the response costs complexity points,
//     so we query only what the local store schema uses.
//   * Order NEWEST so cursor-based pagination is stable across restarts.
const (
	// BackfillPostsQuery pages PH posts within a date window. Used by the
	// `backfill` command. Variables:
	//   first: page size (<= 20 is polite on the budget)
	//   after: cursor string (empty on first call)
	//   postedAfter: ISO8601 DateTime
	//   postedBefore: ISO8601 DateTime
	BackfillPostsQuery = `
query BackfillPosts($first: Int!, $after: String, $postedAfter: DateTime, $postedBefore: DateTime) {
  posts(first: $first, after: $after, postedAfter: $postedAfter, postedBefore: $postedBefore, order: NEWEST) {
    pageInfo { hasNextPage endCursor }
    edges {
      node {
        id
        slug
        name
        tagline
        url
        website
        createdAt
        votesCount
        user { name username }
      }
    }
  }
}`

	// EnrichPostsQuery fetches posts whose name/slug/tagline plausibly match
	// a topic, within the last 30 days. Used by `search --enrich`. PH's
	// GraphQL doesn't support full-text in the posts field; we use NEWEST
	// order + a tight time window and filter client-side.
	//
	// Variables:
	//   first: page size (<= 20 — single page is the whole enrichment)
	//   postedAfter: ISO8601 DateTime (30 days ago)
	//   postedBefore: ISO8601 DateTime (now)
	EnrichPostsQuery = `
query EnrichPosts($first: Int!, $postedAfter: DateTime, $postedBefore: DateTime) {
  posts(first: $first, postedAfter: $postedAfter, postedBefore: $postedBefore, order: NEWEST) {
    pageInfo { endCursor }
    edges {
      node {
        id
        slug
        name
        tagline
        url
        website
        createdAt
        votesCount
        user { name username }
      }
    }
  }
}`
)
