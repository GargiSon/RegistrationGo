package handler

type User struct {
	ID          int
	Username    string
	Email       string
	Mobile      string
	Address     string
	Gender      string
	Sports      string
	DOB         string
	Country     string
	image       []byte
	ImageBase64 string
}

type EditPageData struct {
	User       User
	Countries  []string
	SportsMap  map[string]bool
	Error      string
	Title      string
	Users      []User
	Page       int
	TotalPages int
	Info       string
	Email      string
	Ts         string
	Token      string
	Sort       string
}
