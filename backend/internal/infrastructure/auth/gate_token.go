package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// GateTokenGenerator はゲート(Androidアプリ)用の認証トークンを発行する。
// トークンは推測不能な32バイトの乱数で、DBにはSHA-256のハッシュだけを保存する。
// 乱数が十分に長いため、ハッシュにソルトやストレッチングは使わない(照合のため決定的である必要がある)。
type GateTokenGenerator struct{}

func NewGateTokenGenerator() *GateTokenGenerator { return &GateTokenGenerator{} }

// Generate は平文トークンとそのハッシュを返す。平文はユーザーに一度だけ渡し、保存しない。
func (g *GateTokenGenerator) Generate() (plain, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	plain = base64.RawURLEncoding.EncodeToString(buf)
	return plain, HashGateToken(plain), nil
}

// HashGateToken は平文トークンを保存・照合用のハッシュに変換する。
func HashGateToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
