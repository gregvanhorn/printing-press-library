package instacart

// Persisted GraphQL operations captured from live Instacart web traffic on
// 2026-04-11. Hashes are tied to a specific web bundle and may become
// stale when Instacart ships a new client; use `instacart capture` to refresh.
//
// Each entry ships with a minimum reconstructed query text the server can
// fall back to if the hash is rejected (PersistedQueryNotFound). The query
// texts below are hand-derived from the Apollo client AST captured during
// the sniff and cover the fields we actually read.
var DefaultOps = map[string]OpSeed{
	"CurrentUserFields": {
		Hash:  "d7d1050d8a8efb9a24d2fd0d9c39f58d852ab84ea709370bcbedbca790112952",
		Query: `query CurrentUserFields { currentUser { id email firstName lastName __typename } }`,
	},
	"ShopCollectionScoped": {
		Hash:  "c6a0fcb3d1a4a14e5800cc6c38e736e85177f80f0c01a5535646f83238e65bcb",
		Query: `query ShopCollectionScoped($retailerSlug: String!, $postalCode: String!, $coordinates: ShopCollectionCoordinatesInput, $addressId: ID, $allowCanonicalFallback: Boolean) { shopCollection(retailerSlug: $retailerSlug, postalCode: $postalCode, coordinates: $coordinates, addressId: $addressId, allowCanonicalFallback: $allowCanonicalFallback) { __typename } }`,
	},
	"ShopCollectionUnscoped": {
		Hash:  "814aa179ab4aaf604c50f65150a589ce17e048747de18a1e67c6ad8af626f7a8",
		Query: `query ShopCollectionUnscoped($postalCode: String!, $coordinates: ShopCollectionCoordinatesInput, $addressId: ID) { shopCollection(postalCode: $postalCode, coordinates: $coordinates, addressId: $addressId) { __typename } }`,
	},
	"DefaultShop": {
		Hash:  "607e5aea2e2f7b3d8bf89bb7f657ce5c13bc12f89f70a7f7e6b2f1c13ded18a8",
		Query: `query DefaultShop($postalCode: String!, $coordinates: ShopCollectionCoordinatesInput, $addressId: ID) { defaultShop(postalCode: $postalCode, coordinates: $coordinates, addressId: $addressId) { __typename } }`,
	},
	"Autosuggestions": {
		Hash:  "c1342d16e457bb75a345baf43866531f8b7543e7a3199c3b56d4911f5d23f79a",
		Query: `query Autosuggestions($retailerInventorySessionToken: String, $query: String!, $autosuggestionSessionId: String) { autosuggestions(query: $query, retailerInventorySessionToken: $retailerInventorySessionToken, autosuggestionSessionId: $autosuggestionSessionId) { __typename } }`,
	},
	"ActiveCartId": {
		Hash:  "6803f97683d706ab6faa3c658a0d6766299dbe1ff55f78b720ca2ef77de7c5c7",
		Query: `query ActiveCartId($addressId: ID!, $shopId: ID!) { activeCartId(addressId: $addressId, shopId: $shopId) { id __typename } }`,
	},
	"PersonalActiveCarts": {
		Hash:  "eac9d17bd45b099fbbdabca2e111acaf2a4fa486f2ce5bc4e8acbab2f31fd8c0",
		Query: `query PersonalActiveCarts { personalActiveCarts { id cartType retailerId itemCount updatedAt __typename } }`,
	},
	"CartItemCount": {
		Hash:  "2a89f7495cfb8c1ffcb61158b561ee503fed5d0ad40e076720763c2eb806b8f6",
		Query: `query CartItemCount($id: ID!) { cart(id: $id) { id itemCount __typename } }`,
	},
	"ShopBaskets": {
		Hash:  "868584829e143dc4eb31b360c895c63e11e75af1b6d033ec8a9518d39079e937",
		Query: `query ShopBaskets($shopId: ID!, $addressId: ID!) { shopBaskets(shopId: $shopId, addressId: $addressId) { __typename } }`,
	},
	"Items": {
		Hash:  "5116339819ff07f207fd38f949a8a7f58e52cc62223b535405b087e3076ebf2f",
		Query: `query Items($ids: [ID!]!, $shopId: ID, $zoneId: ID, $postalCode: String) { items(ids: $ids, shopId: $shopId, zoneId: $zoneId, postalCode: $postalCode) { id name __typename } }`,
	},
	"ItemAttributePreferences": {
		Hash:  "927a3d148c6393b86b69e510f3e725d7fd44b19b492b8ed39f21665f76a903ed",
		Query: `query ItemAttributePreferences($productIds: [ID!]!, $shopId: ID!) { itemAttributePreferences(productIds: $productIds, shopId: $shopId) { __typename } }`,
	},
	"GetAddressById": {
		Hash:  "fa58dbba5eef5e9aa65c316621e6e39e0130b7a03cffa99cd13ded5912779a9c",
		Query: `query GetAddressById($id: ID!) { address(id: $id) { id postalCode latitude longitude streetAddress __typename } }`,
	},
	"UpdateCartItemsMutation": {
		// Real Apollo-computed hash, extracted from a live browser session by
		// invoking persistedQueryLink.request() with the captured document.
		// Instacart's server requires this hash for mutations (rejects plain
		// query text with PersistedQueryNotSupported). Refresh via `instacart
		// capture --live` when the bundle rotates.
		Hash:  "a33745461a4b19f7ae3d65e38d31f96412a352c64a4dbf4ea1c7302de1b85572",
		Query: "", // intentionally empty: mutations only go through hash-based APQ
	},

	// CartData: full cart contents including cartItemCollection.cartItems[].
	// Used by `cart show` to resolve real item names and quantities.
	"CartData": {
		Hash:  "05e3d7448576ff7d464c9244fb6687fabd1bdeee85fdf817e85487003cdb6d44",
		Query: `query CartData($id: ID!) { userCart: cart(id: $id) { id itemCount cartType retailerId updatedAt cartItemCollection { cartItems { id quantity quantityType basketProduct { id __typename } __typename } __typename } __typename } }`,
	},

	// DEPRECATED: BuyItAgainPage and CustomerOrderHistory were speculative in
	// the original PR #78 plan. Instacart does NOT expose these operations.
	// The order history is server-rendered HTML, and per-order detail assembles
	// from multiple non-aggregating ops in Apollo's client cache. See
	// docs/solutions/best-practices/instacart-orders-no-clean-graphql-op.md
	// for the finding + docs/patterns/authenticated-session-scraping.md for
	// the working data-load path (Chrome MCP dump -> `instacart history import`).
	//
	// These entries are kept with empty hashes so `instacart capture` skips
	// them cleanly. They will be removed in a future cleanup PR.
	"BuyItAgainPage":       {Hash: "", Query: ""},
	"CustomerOrderHistory": {Hash: "", Query: ""},
}

type OpSeed struct {
	Hash  string
	Query string
}

// OpNames returns the list of known operation names in a stable order.
func OpNames() []string {
	return []string{
		"CurrentUserFields",
		"ShopCollectionScoped",
		"ShopCollectionUnscoped",
		"DefaultShop",
		"Autosuggestions",
		"ActiveCartId",
		"PersonalActiveCarts",
		"CartItemCount",
		"ShopBaskets",
		"Items",
		"ItemAttributePreferences",
		"GetAddressById",
		"UpdateCartItemsMutation",
		"CartData",
		"BuyItAgainPage",
		"CustomerOrderHistory",
	}
}

// HistoryOpNames returns the subset of operations backing the history
// feature. Used by commands like `history sync` to check whether all
// required hashes are populated before firing a request.
func HistoryOpNames() []string {
	return []string{"BuyItAgainPage", "CustomerOrderHistory"}
}
