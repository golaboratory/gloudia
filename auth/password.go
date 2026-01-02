package auth

import (
	"log/slog"
	"strings"

	"github.com/golaboratory/gloudia/environment"
	"github.com/newmo-oss/ergo"
	"golang.org/x/crypto/bcrypt"
)

var (
	// passwordNumbers はパスワード生成に使用される数字の文字を含みます。
	passwordNumbers = []rune("0123456789")
	// passwordUpperAlphabets はパスワード生成に使用される大文字のアルファベットを含みます。
	passwordUpperAlphabets = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	// passwordLowerAlphabets はパスワード生成に使用される小文字のアルファベットを含みます。
	passwordLowerAlphabets = []rune("abcdefghijklmnopqrstuvwxyz")
	// passwordSymbols はパスワード生成に使用される記号の文字を含みます。
	passwordSymbols = []rune("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~")
)

// ValidateStrength は指定された条件を満たすかどうかパスワードの強度をチェックします。
// パスワードが条件を満たす場合はtrueを返し、そうでない場合はfalseを返します。
// 引数:
//   - password: チェック対象のパスワード
//   - includeUpper: 大文字を含める必要があるか
//   - includeLower: 小文字を含める必要があるか
//   - includeNumber: 数字を含める必要があるか
//   - includeSymbol: 記号を含める必要があるか
//   - minLength: 最小長
//
// 戻り値:
//   - bool: 条件を満たす場合はtrue
//   - error: エラー情報
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

// HashPassword は平文パスワードを bcrypt でハッシュ化して返します。
func HashPassword(password string) (string, error) {

	var cryptCost = bcrypt.DefaultCost
	env, err := environment.NewEnvValue[environment.GloudiaEnv]()
	if err == nil {
		cryptCost = env.CryptCost
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cryptCost)
	if err != nil {
		return "", ergo.New("failed to hash password", slog.String("error", err.Error()))
	}
	return string(hashedPassword), nil
}

// CheckPassword はハッシュ化されたパスワードと平文パスワードが一致するか検証します。
// 一致する場合は nil を、不一致の場合は error を返します。
func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// containsRunes は文字列に指定されたルーンが含まれているかどうかをチェックします。
// ルーンが文字列に含まれている場合はtrueを返し、そうでない場合はfalseを返します。
// 引数:
//   - s: チェック対象の文字列
//   - runes: 含めるべきルーンのスライス
//
// 戻り値:
//   - bool: 含まれていればtrue
func containsRunes(s string, runes []rune) bool {
	for _, r := range runes {
		if strings.ContainsRune(s, r) {
			return true
		}
	}
	return false
}
