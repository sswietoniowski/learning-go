package main

var (
	Name     = "Alice"
	Password = "top-secret"
)

func UpdateUser(newName *string, newPassword *string) {
	if newName == nil && newPassword == nil {
		return
	}
	if newName != nil {
		Name = *newName
	}
	if newPassword != nil {
		Password = *newPassword
	}
}

func main() {
	// Nothing to update
	UpdateUser(nil, nil)

	// Update password
	newPassword := "much-more-secure"
	UpdateUser(nil, &newPassword)

	// Update both
	newName := "Bob"
	UpdateUser(&newName, &newPassword)
}
