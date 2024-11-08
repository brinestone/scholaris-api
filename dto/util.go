package dto

type CursorBasedPaginationParams struct {
	After uint64 `query:"after"`
	Size  uint   `query:"size"`
}

type PageBasedPaginationParams struct {
	Page uint `query:"page" json:"page" encore:"optional"`
	Size uint `query:"size" json:"size" encore:"optional"`
}

func (p *PageBasedPaginationParams) Validate() error {
	if p.Size == 0 {
		p.Size = 10
	}
	return nil
}

type PaginatedResponseMeta struct {
	Total uint `json:"total"`
}

type PaginatedResponse[T any] struct {
	Data []*T                  `json:"data"`
	Meta PaginatedResponseMeta `json:"meta"`
}
