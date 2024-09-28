package dto

type PaginationParams struct {
	After uint64 `query:"after"`
	Size  uint   `query:"size"`
}
