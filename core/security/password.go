package security

import (
	"math/rand/v2"
	"strings"
	"time"
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

// GeneratePassword は指定された条件に基づいてランダムなパスワードを生成します。
// 大文字、小文字、数字、記号の使用を指定し、生成するパスワードの長さを指定します。
// 生成されたパスワードを文字列として返します。
// 引数:
//   - includeUpper: 大文字を含めるか
//   - includeLower: 小文字を含めるか
//   - includeNumber: 数字を含めるか
//   - includeSymbol: 記号を含めるか
//   - length: パスワードの長さ
//
// 戻り値:
//   - string: 生成されたパスワード
func GeneratePassword(
	includeUpper, includeLower, includeNumber, includeSymbol bool,
	length int) string {

	var required []rune
	var runes []rune
	if includeUpper {
		runes = append(runes, passwordUpperAlphabets...)
		// 必要な大文字を１文字追加
		rnd := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 1))
		required = append(required, passwordUpperAlphabets[rnd.IntN(len(passwordUpperAlphabets))])
	}
	if includeLower {
		runes = append(runes, passwordLowerAlphabets...)
		rnd := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 2))
		required = append(required, passwordLowerAlphabets[rnd.IntN(len(passwordLowerAlphabets))])
	}
	if includeNumber {
		runes = append(runes, passwordNumbers...)
		rnd := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 3))
		required = append(required, passwordNumbers[rnd.IntN(len(passwordNumbers))])
	}
	if includeSymbol {
		runes = append(runes, passwordSymbols...)
		rnd := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 4))
		required = append(required, passwordSymbols[rnd.IntN(len(passwordSymbols))])
	}

	// 指定長より必要文字数が多い場合は、必要文字数だけで生成
	if length < len(required) {
		length = len(required)
	}

	// 残りの桁をランダムに埋める
	rnd := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 5))
	result := make([]rune, 0, length)
	result = append(result, required...)
	for i := len(required); i < length; i++ {
		result = append(result, runes[rnd.IntN(len(runes))])
	}

	// シャッフル処理
	for i := range result {
		j := rnd.IntN(len(result))
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
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
