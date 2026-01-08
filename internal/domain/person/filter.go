package person

type ListFilter struct {
	Name       string
	Email      string
	City       string
	State      string
	ActiveOnly bool
	Page       int
	Limit      int
}

const DefaultLimit = 20
const MaxLimit = 100

func (f *ListFilter) Normalize() {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = DefaultLimit
	}
	if f.Limit > MaxLimit {
		f.Limit = MaxLimit
	}
}

func (f *ListFilter) Offset() int {
	return (f.Page - 1) * f.Limit
}

type ListResult struct {
	Persons []*Person
	Total   int
	Page    int
	Limit   int
}

func (r *ListResult) TotalPages() int {
	if r.Total == 0 {
		return 0
	}
	pages := r.Total / r.Limit
	if r.Total%r.Limit > 0 {
		pages++
	}
	return pages
}

func (r *ListResult) HasNextPage() bool {
	return r.Page < r.TotalPages()
}

func (r *ListResult) HasPrevPage() bool {
	return r.Page > 1
}
