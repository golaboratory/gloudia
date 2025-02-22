package security

import (
	"testing"
)

// TestHashString_Deterministic は、同じ入力に対して常に同じハッシュ値が返されることを検証します。
func TestHashString_Deterministic(t *testing.T) {
	crypt := &Cryptography{}
	input := "同一テスト文字列"
	hash1 := crypt.HashString(input)
	hash2 := crypt.HashString(input)
	if hash1 != hash2 {
		t.Errorf("同一入力で異なるハッシュ値が生成されました: %s と %s", hash1, hash2)
	}
}

// TestHashString_Unique は、異なる入力に対して異なるハッシュ値が返されることを検証します。
func TestHashString_Unique(t *testing.T) {
	crypt := &Cryptography{}
	input1 := "入力１"
	input2 := "入力２"
	hash1 := crypt.HashString(input1)
	hash2 := crypt.HashString(input2)
	if hash1 == hash2 {
		t.Errorf("異なる入力にもかかわらず同一のハッシュ値が生成されました: %s", hash1)
	}
}

// TestHashString_Length は、生成されるハッシュ値の長さが期待通り（SHA-256 ハッシュ：32バイト → 64文字の16進数）であることを検証します。
func TestHashString_Length(t *testing.T) {
	crypt := &Cryptography{}
	input := "任意の入力"
	hash := crypt.HashString(input)
	const expectedLength = 64
	if len(hash) != expectedLength {
		t.Errorf("ハッシュ値の長さが期待値と異なります: 期待 %d, 実際 %d", expectedLength, len(hash))
	}
}
