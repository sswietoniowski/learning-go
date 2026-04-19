package conditionals

var Password = "current-password"

func ResetPassword(code int) {
	if code != 2022 {
		return
	}
	Password = "new-password"
}
