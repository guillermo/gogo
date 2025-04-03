package user

type User struct {
	ID   int
	Name string
}

func (u *User) GetName() string {
	return u.Name
}

func (u *User) SetName(name string) {
	u.Name = name
}
