package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
)

var SIZE int = 512

const PRIME_LIMIT_CHECK = 1000
const EXP int = 65537

type RSAKeys struct {
	P *big.Int
	Q *big.Int
	E *big.Int
	D *big.Int
	N *big.Int
}

type PubKey struct {
	N *big.Int
	E *big.Int
}

func SievePrimes(limit int) (list []int) {
	list = append(list, 2, 3)

	for i := 5; i < limit; i += 2 {
		isPrime := true

		for _, v := range list {
			if i%v == 0 {
				isPrime = false
				break
			}
		}

		if isPrime {
			list = append(list, i)
		}
	}

	return
}

func FLTPrimeTest(base int, n *big.Int) bool {
	baseBig := big.NewInt(int64(base))
	one := big.NewInt(1)
	minusOne := big.NewInt(0)
	minusOne = minusOne.Abs(n)
	minusOne = minusOne.Sub(n, one)

	mod := big.NewInt(1)
	mod.Mod(n, baseBig)

	if mod.Sign() == 0 {
		fmt.Println("FAILED DIV BY", base)
		return false
	}

	r := baseBig.Exp(baseBig, minusOne, n)

	return r.Cmp(one) == 0
}

func GetRandomOddNumber(rng *rand.Rand) *big.Int {
	n := big.NewInt(1)
	m := big.NewInt(0)
	one := big.NewInt(1)
	two := big.NewInt(2)
	size := big.NewInt(int64(SIZE))

	for i := big.NewInt(1); i.Cmp(size) == -1; i.Add(i, one) {
		dig := rng.Intn(2)
		if dig == 1 {
			n.Add(n, m.Exp(two, i, nil))
		}
	}

	return n
}

func TestFLTBases(baseList []int, n *big.Int) (isPrime bool) {
	isPrime = true

	for _, v := range baseList {
		if !FLTPrimeTest(v, n) {
			isPrime = false
			if v > 2 {
				fmt.Println("FAILED BASE", v)
			}
			return
		}
	}

	return
}

func GetPseudoPrime(baseList []int, rng *rand.Rand) *big.Int {
	n := GetRandomOddNumber(rng)
	one := big.NewInt(1)

	for n.Cmp(one) < 1 || !TestFLTBases(baseList, n) {
		n = GetRandomOddNumber(rng)
	}

	return n
}

func GenerateRSAKeys(rng *rand.Rand) (keys RSAKeys) {
	primes := SievePrimes(PRIME_LIMIT_CHECK)
	keys.P = GetPseudoPrime(primes, rng)
	keys.Q = GetPseudoPrime(primes, rng)
	keys.N = big.NewInt(1)
	keys.N.Mul(keys.P, keys.Q)
	keys.E = big.NewInt(int64(EXP))

	phiN := big.NewInt(-1)
	phiN.Add(phiN, keys.P)

	qMinusOne := big.NewInt(-1)
	qMinusOne.Add(qMinusOne, keys.Q)

	phiN.Mul(phiN, qMinusOne)

	keys.D = big.NewInt(1)
	keys.D.ModInverse(keys.E, phiN)

	return
}

func GetPublicKey(keys RSAKeys) (pub PubKey) {
	pub.E = keys.E
	pub.N = keys.N

	return
}

func EncryptChunk(chunk []byte, pub PubKey) []byte {
	byteNum := big.NewInt(0)
	byteNum.SetBytes(chunk)

	encryptedNum := big.NewInt(1)

	encryptedNum.Exp(byteNum, pub.E, pub.N)

	return encryptedNum.Bytes()
}

func SignChunk(chunk []byte, keys RSAKeys) []byte {
	pk := PubKey{
		keys.N,
		keys.D,
	}

	return EncryptChunk(chunk, pk)
}

func DecryptChunk(chunk []byte, keys RSAKeys) []byte {
	byteNum := big.NewInt(0)
	byteNum.SetBytes(chunk)

	decryptedNum := big.NewInt(1)
	decryptedNum.Exp(byteNum, keys.D, keys.N)

	return decryptedNum.Bytes()
}

func VerifyRSASignature(chunk []byte, signature []byte, pub PubKey) bool {
	dec := EncryptChunk(signature, pub)

	if len(dec) != len(chunk) {
		return false
	}

	for i, v := range chunk {
		if dec[i] != v {
			return false
		}
	}

	return true
}

func EncryptChunkToB64(chunk []byte, pub PubKey) string {
	enc := EncryptChunk(chunk, pub)
	return base64.StdEncoding.EncodeToString(enc)
}

func DecryptChunkFromB64(s string, keys RSAKeys) []byte {
	bytes, _ := base64.StdEncoding.DecodeString(s)
	return DecryptChunk(bytes, keys)
}

func SignChunkFromB64(s string, keys RSAKeys) []byte {
	bytes, _ := base64.StdEncoding.DecodeString(s)
	return SignChunk(bytes, keys)
}

func VerifyRSASignatureFromB64(cont string, sig string, pub PubKey) bool {
	contBytes, _ := base64.StdEncoding.DecodeString(cont)
	sigBytes, _ := base64.StdEncoding.DecodeString(sig)

	return VerifyRSASignature(contBytes, sigBytes, pub)
}

func KeysToJson(keys RSAKeys) (j string) {
	bytes, _ := json.Marshal(keys)

	return string(bytes)
}

func LoadKeysFromJson(j string) (keys RSAKeys) {
	json.Unmarshal([]byte(j), &keys)

	return
}

func BigIntToB64(n *big.Int) string {
	bytes := n.Bytes()
	return base64.StdEncoding.EncodeToString(bytes)
}

func B64ToBigInt(b64 string) *big.Int {
	bytes, _ := base64.StdEncoding.DecodeString(b64)
	n := big.NewInt(0)
	n.SetBytes(bytes)
	return n
}

func MarshalKeys(keys RSAKeys) string {
	p := BigIntToB64(keys.P)
	q := BigIntToB64(keys.Q)
	e := BigIntToB64(keys.E)
	d := BigIntToB64(keys.D)
	n := BigIntToB64(keys.N)

	return fmt.Sprintf("%s.%s.%s.%s.%s", p, q, e, d, n)
}

func MarshalPubKey(pub PubKey) string {
	e := BigIntToB64(pub.E)
	n := BigIntToB64(pub.N)

	return fmt.Sprintf("%s.%s", e, n)
}

func UnmarshalKeys(s string) RSAKeys {
	parts := strings.Split(s, ".")
	p := parts[0]
	q := parts[1]
	e := parts[2]
	d := parts[3]
	n := parts[4]

	return RSAKeys{
		P: B64ToBigInt(p),
		Q: B64ToBigInt(q),
		E: B64ToBigInt(e),
		D: B64ToBigInt(d),
		N: B64ToBigInt(n),
	}
}

func UnmarshalPubKey(s string) PubKey {
	parts := strings.Split(s, ".")
	e := parts[0]
	n := parts[1]

	return PubKey{
		E: B64ToBigInt(e),
		N: B64ToBigInt(n),
	}
}
