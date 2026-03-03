package response

type Pagination struct {
	Items any   `json:"items"`
	Total int64 `json:"total"`
}
