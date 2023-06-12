package validator

import (
    "testing"
)

type User struct {
    Username string `validate:"required,regex=^[a-zA-Z0-9]+$"`
    Email    string `validate:"required,email"`
}

func TestValidation(t *testing.T) {
    validate := New()

    // Valid user
    user := User{
        Username: "JohnDoe123",
        Email:    "johndoe@example.com",
    }

    err := validate.Struct(user)
    if err != nil {
        t.Errorf("Validation failed unexpectedly: %s", err.Error())
    }

    // Invalid username (contains special characters)
    user = User{
        Username: "John@Doe",
        Email:    "johndoe@example.com",
    }

    err = validate.Struct(user)
    if err == nil {
        t.Error("Validation should have failed for invalid username")
    }

    // Invalid email (missing domain)
    user = User{
        Username: "JohnDoe123",
        Email:    "johndoe",
    }

    err = validate.Struct(user)
    if err == nil {
        t.Error("Validation should have failed for invalid email")
    }
}

func TestCustomValidation(t *testing.T) {
    validate := New()

    // Valid UUID
    err := validate.ValidateUUID("e0a4fa0a-14ea-485f-9a16-eb68425cf9c2")
    if err != nil {
        t.Errorf("Validation failed unexpectedly: %s", err.Error())
    }

    // Invalid UUID
    err = validate.ValidateUUID("invalid-uuid")
    if err == nil {
        t.Error("Validation should have failed for invalid UUID")
    }
}
