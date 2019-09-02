package main

type User struct {
	ID        int    `json:"id,omitempty"`
	Email     string `json:"email" validate:"nonzero"`
	FirstName string `json:"first_name" validate:"nonzero"`
	LastName  string `json:"last_name" validate:"nonzero"`
	Gender    string `json:"gender" validate:"nonzero"`
	BirthDate int    `json:"birth_date" validate:"nonzero"`
}

type Location struct {
	ID       int    `json:"id,omitempty"`
	Place    string `json:"place"`
	Country  string `json:"country"`
	City     string `json:"city"`
	Distance int    `json:"distance"`
}

type Visit struct {
	ID        int    `json:"id,omitempty"`
	Location  int    `json:"location" validate:"nonzero"`
	User      int    `json:"user" validate:"nonzero"`
	VisitedAt string `json:"visited_at" validate:"nonzero"`
	Mark      int    `json:"mark" validate:"nonzero"`
}

func (User) TableName() string {
	return "users"
}

func (Visit) TableName() string {
	return "visits"
}

func (Location) TableName() string {
	return "locations"
}
