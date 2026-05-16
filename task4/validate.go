package main

import (
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	reName  = regexp.MustCompile(`^[\p{L} ]+$`)
	rePhone = regexp.MustCompile(`^\+?[0-9()\- ]{7,32}$`)
	reEmail = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
)

var validGenders = map[string]bool{
	"male": true, "female": true,
}

var validLanguageIDs = map[string]bool{
	"1": true, "2": true, "3": true, "4": true,
	"5": true, "6": true, "7": true, "8": true,
	"9": true, "10": true, "11": true, "12": true,
}

func validate(r *http.Request) (FormData, FormErrors) {
	var data FormData
	errors := make(FormErrors)

	name := strings.TrimSpace(r.FormValue("name"))
	switch {
	case name == "":
		errors["name"] = "Name is required"
	case utf8.RuneCountInString(name) > 150:
		errors["name"] = "Name must be at most 150 characters"
	case !reName.MatchString(name):
		errors["name"] = "Name contains invalid characters"
	default:
		data.Name = name
	}

	phone := strings.TrimSpace(r.FormValue("phone"))
	switch {
	case phone == "":
		errors["phone"] = "Phone is required"
	case !rePhone.MatchString(phone):
		errors["phone"] = "Phone contains invalid characters"
	default:
		data.Phone = phone
	}

	email := strings.TrimSpace(r.FormValue("email"))
	switch {
	case email == "":
		errors["email"] = "Email is required"
	case len(email) > 255:
		errors["email"] = "Email must be at most 255 characters"
	case !reEmail.MatchString(email):
		errors["email"] = "Email format is invalid, try name@domain.com"
	default:
		data.Email = email
	}

	birthdate := strings.TrimSpace(r.FormValue("birthdate"))
	switch {
	case birthdate == "":
		errors["birthdate"] = "Birthdate is required"
	default:
		parsed, err := time.Parse("2006-01-02", birthdate)
		if err != nil {
			errors["birthdate"] = "Birthdate format is invalid (expected YYYY-MM-DD)"
		} else if parsed.After(time.Now()) {
			errors["birthdate"] = "Birthdate cannot be in the future"
		} else {
			data.Birthdate = birthdate
		}
	}

	gender := r.FormValue("gender")
	if !validGenders[gender] {
		errors["gender"] = "Gender must be 'male' or '	female'"
	} else {
		data.Gender = gender
	}

	languages := r.Form["languages[]"]
	switch {
	case len(languages) == 0:
		errors["languages"] = "At least one language must be selected"
	default:
		allValid := true
		for _, id := range languages {
			if !validLanguageIDs[id] {
				errors["languages"] = "Invalid language selection"
				allValid = false
				break
			}
		}
		if allValid {
			data.Languages = languages
		}
	}

	bio := strings.TrimSpace(r.FormValue("bio"))
	if bio == "" {
		errors["bio"] = "Bio is required"
	} else {
		data.Bio = bio
	}

	contract := r.FormValue("contract")
	if contract == "" {
		errors["contract"] = "You must accept the contract"
	} else if contract != "on" {
		errors["contract"] = "Invalid contract value"
	} else {
		data.Contract = true
	}

	return data, errors
}
