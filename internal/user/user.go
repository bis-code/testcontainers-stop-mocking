package user

import "time"

type User struct {
	ID        int
	Email     string
	Name      string
	CreatedAt time.Time
}
