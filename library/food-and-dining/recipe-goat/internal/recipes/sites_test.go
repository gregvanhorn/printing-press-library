package recipes

import "testing"

func TestFindSiteRecognizesNewRecipeSources(t *testing.T) {
	tests := []struct {
		host      string
		wantName  string
		wantTier  int
		wantTrust float64
	}{
		{host: "www.themediterraneandish.com", wantName: "The Mediterranean Dish", wantTier: 1, wantTrust: 0.9},
		{host: "kitchensanctuary.com", wantName: "Kitchen Sanctuary", wantTier: 1, wantTrust: 0.9},
		{host: "grandbaby-cakes.com", wantName: "Grandbaby Cakes", wantTier: 1, wantTrust: 0.9},
		{host: "mykoreankitchen.com", wantName: "My Korean Kitchen", wantTier: 1, wantTrust: 0.9},
		{host: "www.oliviascuisine.com", wantName: "Olivia's Cuisine", wantTier: 1, wantTrust: 0.9},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := FindSite(tt.host)
			if got.Name != tt.wantName {
				t.Fatalf("FindSite(%q) name = %q, want %q", tt.host, got.Name, tt.wantName)
			}
			if got.Tier != tt.wantTier {
				t.Fatalf("FindSite(%q) tier = %d, want %d", tt.host, got.Tier, tt.wantTier)
			}
			if got.Trust != tt.wantTrust {
				t.Fatalf("FindSite(%q) trust = %v, want %v", tt.host, got.Trust, tt.wantTrust)
			}
			if got.RecipeURLPattern == nil {
				t.Fatalf("FindSite(%q) missing recipe URL pattern", tt.host)
			}
		})
	}
}

func TestNewRecipeSourcePatternsMatchRealRecipePermalinks(t *testing.T) {
	tests := []struct {
		host string
		url  string
	}{
		{host: "themediterraneandish.com", url: "https://www.themediterraneandish.com/zaatar-garlic-salmon-recipe/"},
		{host: "kitchensanctuary.com", url: "https://www.kitchensanctuary.com/crispy-chicken-tenders-garlic-chilli-dip/"},
		{host: "grandbaby-cakes.com", url: "https://grandbaby-cakes.com/salmon-croquettes/"},
		{host: "mykoreankitchen.com", url: "https://mykoreankitchen.com/korean-style-stir-fried-udon-noodles-with-chicken-and-veggies/"},
		{host: "oliviascuisine.com", url: "https://www.oliviascuisine.com/sheet-pan-chicken-fajitas/"},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			site := FindSite(tt.host)
			if !looksLikeRecipeLink(tt.url, site) {
				t.Fatalf("looksLikeRecipeLink(%q, %q) = false, want true", tt.url, tt.host)
			}
		})
	}
}

func TestNewRecipeSourcePatternsRejectObviousNonRecipePaths(t *testing.T) {
	tests := []struct {
		host string
		url  string
	}{
		{host: "themediterraneandish.com", url: "https://www.themediterraneandish.com/search/chicken/"},
		{host: "kitchensanctuary.com", url: "https://www.kitchensanctuary.com/category/chicken/"},
		{host: "grandbaby-cakes.com", url: "https://grandbaby-cakes.com/blog/"},
		{host: "mykoreankitchen.com", url: "https://mykoreankitchen.com/recommends/"},
		{host: "oliviascuisine.com", url: "https://www.oliviascuisine.com/category/recipes/"},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			site := FindSite(tt.host)
			if looksLikeRecipeLink(tt.url, site) {
				t.Fatalf("looksLikeRecipeLink(%q, %q) = true, want false", tt.url, tt.host)
			}
		})
	}
}
