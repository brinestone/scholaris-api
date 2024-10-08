package dto

type PaginationParams struct {
	After uint64 `query:"after"`
	Size  uint   `query:"size"`
}

type PaginatedResponseMeta struct {
	Total uint `json:"total"`
}

type PaginatedResponse[T any] struct {
	Data []*T                  `json:"data"`
	Meta PaginatedResponseMeta `json:"meta"`
}
