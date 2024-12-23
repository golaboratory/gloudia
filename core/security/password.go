package security

import (
	"math/rand/v2"
	"strings"
)

var (
	// passwordNumbers は、パスワード生成に使用される数字の文字を含みます。
	passwordNumbers = []rune("0123456789")

	// passwordUpperAlphabets は、パスワード生成に使用される大文字のアルファベットを含みます。
	passwordUpperAlphabets = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	// passwordLowerAlphabets は、パスワード生成に使用される小文字のアルファベットを含みます。
	passwordLowerAlphabets = []rune("abcdefghijklmnopqrstuvwxyz")

	// passwordSymbols は、パスワード生成に使用される記号の文字を含みます。
	passwordSymbols = []rune("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~")
)

// ValidateStrength は、指定された条件を満たすかどうかパスワードの強度をチェックします。
// パスワードが条件を満たす場合は true を返し、そうでない場合は false を返します。
func ValidateStrength(
	password string,
	includeUpper, includeLower, includeNumber, includeSymbol bool,
	minLength int) (bool, error) {

	if len(password) < minLength {
		return false, nil
	}

	var hasUpper = containsRunes(password, passwordUpperAlphabets)
	if includeUpper && !hasUpper {
		return false, nil
	}

	var hasLower = containsRunes(password, passwordLowerAlphabets)
	if includeLower && !hasLower {
		return false, nil
	}

	var hasNumber = containsRunes(password, passwordNumbers)
	if includeNumber && !hasNumber {
		return false, nil
	}

	var hasSymbol = containsRunes(password, passwordSymbols)
	if includeSymbol && !hasSymbol {
		return false, nil
	}

	return true, nil
}

// GeneratePassword は、指定された条件に基づいてランダムなパスワードを生成します。
// 大文字、小文字、数字、記号の使用を指定し、生成するパスワードの長さを指定します。
// 生成されたパスワードを文字列として返します。
func GeneratePassword(
	includeUpper, includeLower, includeNumber, includeSymbol bool,
	length int) string {

	var runes []rune
	if includeUpper {
		runes = append(runes, passwordUpperAlphabets...)
	}
	if includeLower {
		runes = append(runes, passwordLowerAlphabets...)
	}
	if includeNumber {
		runes = append(runes, passwordNumbers...)
	}
	if includeSymbol {
		runes = append(runes, passwordSymbols...)
	}

	var result []rune
	rnd := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	for _ = range length {
		result = append(result, runes[rnd.IntN(len(runes))])
	}

	return string(result)
}

// containsRunes は、文字列に指定されたルーンが含まれているかどうかをチェックします。
// ルーンが文字列に含まれている場合は true を返し、そうでない場合は false を返します。
func containsRunes(s string, runes []rune) bool {
	for _, r := range runes {
		if strings.ContainsRune(s, r) {
			return true
		}
	}
	return false
}
