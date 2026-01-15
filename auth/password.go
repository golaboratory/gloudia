package auth

import (
	"crypto/rand"
	"log/slog"
	"math/big"
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
	env, err := environment.NewEnvValue[environment.GloudiaEnv]("")
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

// GenerateRandomCode は指定された要件に基づいてランダムなパスコードを生成します。
// 各文字種（大文字、小文字、数字、記号）の使用有無を指定できます。
// 指定された各文字種から、少なくとも1文字が含まれることが保証されます。
// 残りは、指定された文字種の和集合からランダムに選択されます。
// 最終的な結果はシャッフルされます。
//
// 引数:
//   - length: 生成するパスコードの長さ
//   - includeUpper: 大文字を含めるか
//   - includeLower: 小文字を含めるか
//   - includeNumber: 数字を含めるか
//   - includeSymbol: 記号を含めるか
//
// 戻り値:
//   - string: 生成されたパスコード
//   - error: 長さが要件を満たさない場合（0以下、または必須文字種数より短い）、または文字種が一つも選択されていない場合のエラー
func GenerateRandomCode(length int, includeUpper, includeLower, includeNumber, includeSymbol bool) (string, error) {
	if length <= 0 {
		return "", ergo.New("length must be greater than 0")
	}

	var combinedChars []rune
	var requiredChars []rune

	if includeUpper {
		combinedChars = append(combinedChars, passwordUpperAlphabets...)
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordUpperAlphabets))))
		if err != nil {
			return "", ergo.New("failed to generate random number", slog.String("error", err.Error()))
		}
		requiredChars = append(requiredChars, passwordUpperAlphabets[idx.Int64()])
	}

	if includeLower {
		combinedChars = append(combinedChars, passwordLowerAlphabets...)
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordLowerAlphabets))))
		if err != nil {
			return "", ergo.New("failed to generate random number", slog.String("error", err.Error()))
		}
		requiredChars = append(requiredChars, passwordLowerAlphabets[idx.Int64()])
	}

	if includeNumber {
		combinedChars = append(combinedChars, passwordNumbers...)
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordNumbers))))
		if err != nil {
			return "", ergo.New("failed to generate random number", slog.String("error", err.Error()))
		}
		requiredChars = append(requiredChars, passwordNumbers[idx.Int64()])
	}

	if includeSymbol {
		combinedChars = append(combinedChars, passwordSymbols...)
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordSymbols))))
		if err != nil {
			return "", ergo.New("failed to generate random number", slog.String("error", err.Error()))
		}
		requiredChars = append(requiredChars, passwordSymbols[idx.Int64()])
	}

	if len(combinedChars) == 0 {
		return "", ergo.New("at least one character type must be selected")
	}

	if length < len(requiredChars) {
		return "", ergo.New("length is too short to include all required character types")
	}

	// 必須文字以外をランダムに埋める
	resultRunes := make([]rune, 0, length)
	resultRunes = append(resultRunes, requiredChars...)
	remaining := length - len(requiredChars)

	for i := 0; i < remaining; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(combinedChars))))
		if err != nil {
			return "", ergo.New("failed to generate random number", slog.String("error", err.Error()))
		}
		resultRunes = append(resultRunes, combinedChars[idx.Int64()])
	}

	// シャッフル (Fisher-Yates)
	for i := len(resultRunes) - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", ergo.New("failed to shuffle", slog.String("error", err.Error()))
		}
		j := int(jBig.Int64())
		resultRunes[i], resultRunes[j] = resultRunes[j], resultRunes[i]
	}

	return string(resultRunes), nil
}
