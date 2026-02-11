package request

type Pagination struct {
	Page int `json:"page" form:"page"`
	Size int `json:"size" form:"size"`
}

func (p *Pagination) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}

	if p.Size <= 0 {
		p.Size = 10
	}

	if p.Size > 100 {
		p.Size = 100
	}
}

func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.Size
}

func (p *Pagination) Limit() int {
	return p.Size
}
