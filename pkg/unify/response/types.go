package response

//ListResult ... data format in listing requst
type ListResult struct {
	Items       []interface{} `json:"items"`
	Total       int           `json:"totalItems"`
	PageSize    int           `json:"pageSize"`
	TotalPages  int           `json:"totalPages"`
	CurrentPage int           `json:"currentPage"`
}
