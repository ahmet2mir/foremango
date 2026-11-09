package foreman

// ForemanObject holds the fields Foreman returns for every object. Every
// generated Foreman<X> type embeds it.
type ForemanObject struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// List is one page of a Foreman index (list or search) response. Every
// generated QueryX method returns one.
type List[T any] struct {
	// Total is how many objects of this kind exist.
	Total int `json:"total"`
	// Subtotal is how many of them match the search.
	Subtotal int `json:"subtotal"`
	// Page and PerPage locate this page within the matches.
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	// Search is the search expression Foreman applied.
	Search string `json:"search"`
	// Results are the matching objects on this page.
	Results []T `json:"results"`
}
