package vo

type Address struct {
	Street       string
	Number       string
	Complement   string
	Neighborhood string
	City         string
	State        string
	ZipCode      string
}

func (a Address) IsEmpty() bool {
	return a.Street == "" &&
		a.Number == "" &&
		a.Complement == "" &&
		a.Neighborhood == "" &&
		a.City == "" &&
		a.State == "" &&
		a.ZipCode == ""
}
