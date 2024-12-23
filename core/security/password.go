package security

var (
	passwordNumbers        = []rune("0123456789")
	passwordUpperAlphabets = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	passwordLowerAlphabets = []rune("abcdefghijklmnopqrstuvwxyz")
	passwordSymbols        = []rune("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~")
)

func ValidateStrength(
	password string,
	includeUpper, includeLower, includeNumber, includeSymbol bool,
	minLength int) (bool, error) {

	if len(password) < minLength {
		return false, nil
	}

	if includeUpper && !hasUpper(password) {
		return false, nil
	}

	return true, nil
}
