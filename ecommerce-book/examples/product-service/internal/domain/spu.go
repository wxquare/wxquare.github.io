package domain

// SPU 实体（Standard Product Unit）
type SPU struct {
	id          int64
	title       string
	categoryID  int64
	brandID     int64
	description string
	attributes  map[string]string
}

func NewSPU(id int64, title string, categoryID, brandID int64) *SPU {
	return &SPU{
		id:         id,
		title:      title,
		categoryID: categoryID,
		brandID:    brandID,
		attributes: make(map[string]string),
	}
}

func (s *SPU) ID() int64 {
	return s.id
}

func (s *SPU) Title() string {
	return s.title
}

func (s *SPU) CategoryID() int64 {
	return s.categoryID
}

func (s *SPU) BrandID() int64 {
	return s.brandID
}

func (s *SPU) Description() string {
	return s.description
}

func (s *SPU) SetDescription(desc string) {
	s.description = desc
}

func (s *SPU) Attributes() map[string]string {
	return s.attributes
}

func (s *SPU) SetAttribute(key, value string) {
	s.attributes[key] = value
}
