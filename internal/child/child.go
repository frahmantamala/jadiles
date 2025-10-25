package child

import (
	"fmt"
	"time"
)

// Child represents the core child domain model
type Child struct {
	ID           int64
	ParentID     int64
	Name         string
	Nickname     string
	DateOfBirth  time.Time
	Gender       Gender
	SpecialNeeds string
	Photo        string
	Version      int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

// ValidateName validates child name
func (c *Child) ValidateName() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(c.Name) < 2 {
		return fmt.Errorf("name must be at least 2 characters")
	}
	if len(c.Name) > 100 {
		return fmt.Errorf("name must not exceed 100 characters")
	}
	return nil
}

// ValidateNickname validates nickname if provided
func (c *Child) ValidateNickname() error {
	if c.Nickname != "" && len(c.Nickname) > 50 {
		return fmt.Errorf("nickname must not exceed 50 characters")
	}
	return nil
}

// ValidateDateOfBirth validates date of birth
func (c *Child) ValidateDateOfBirth() error {
	if c.DateOfBirth.IsZero() {
		return fmt.Errorf("date of birth is required")
	}

	// Child must not be born in the future
	if c.DateOfBirth.After(time.Now()) {
		return fmt.Errorf("date of birth cannot be in the future")
	}

	// Child must be less than 18 years old (for kids activities)
	age := c.CalculateAge()
	if age >= 18 {
		return fmt.Errorf("child must be under 18 years old")
	}

	// Child must be at least born (0 years old)
	if age < 0 {
		return fmt.Errorf("invalid date of birth")
	}

	return nil
}

// ValidateGender validates gender
func (c *Child) ValidateGender() error {
	if c.Gender != GenderMale && c.Gender != GenderFemale {
		return fmt.Errorf("gender must be either 'male' or 'female'")
	}
	return nil
}

// ValidateSpecialNeeds validates special needs if provided
func (c *Child) ValidateSpecialNeeds() error {
	if c.SpecialNeeds != "" && len(c.SpecialNeeds) > 500 {
		return fmt.Errorf("special needs must not exceed 500 characters")
	}
	return nil
}

// Validate validates all child fields
func (c *Child) Validate() error {
	if c.ParentID == 0 {
		return fmt.Errorf("parent_id is required")
	}

	if err := c.ValidateName(); err != nil {
		return err
	}

	if err := c.ValidateNickname(); err != nil {
		return err
	}

	if err := c.ValidateDateOfBirth(); err != nil {
		return err
	}

	if err := c.ValidateGender(); err != nil {
		return err
	}

	if err := c.ValidateSpecialNeeds(); err != nil {
		return err
	}

	return nil
}

// CalculateAge calculates the child's age in years
func (c *Child) CalculateAge() int {
	now := time.Now()
	age := now.Year() - c.DateOfBirth.Year()

	// Adjust if birthday hasn't occurred this year
	if now.YearDay() < c.DateOfBirth.YearDay() {
		age--
	}

	return age
}

// GetDisplayName returns nickname if set, otherwise name
func (c *Child) GetDisplayName() string {
	if c.Nickname != "" {
		return c.Nickname
	}
	return c.Name
}
