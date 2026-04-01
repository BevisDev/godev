package request

const (
	defaultPaginationPage = 1
	defaultPaginationSize = 10
	maxPaginationSize     = 100
)

type Pagination struct {
	Page int `json:"page" form:"page"`
	Size int `json:"size" form:"size"`
}

func (p *Pagination) Normalize() {
	if p.Page < 1 {
		p.Page = defaultPaginationPage
	}

	if p.Size <= 0 {
		p.Size = defaultPaginationSize
	}

	if p.Size > maxPaginationSize {
		p.Size = maxPaginationSize
	}
}

func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.Size
}

func (p *Pagination) Limit() int {
	return p.Size
}
