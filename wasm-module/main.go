package main

import (
	"crypto/ecdsa"
	"flag"
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

type Address string

func getAddress(k *ecdsa.PrivateKey) Address {
	pubKey := k.Public()
	pubKeyECDSA, _ := pubKey.(*ecdsa.PublicKey)
	return Address(crypto.PubkeyToAddress(*pubKeyECDSA).Hex())
}

func MarshalDHKey(key DHPrivateKey) string {
	return fmt.Sprintf("%s.%s", BigIntToB64(key.a), BigIntToB64(key.ga))
}

func UnmarshalDHKey(b64 string) DHPrivateKey {
	parts := strings.Split(b64, ".")
	return DHPrivateKey{
		B64ToBigInt(parts[0]),
		B64ToBigInt(parts[1]),
	}
}

func UnmarshalBigIntList(b64 string) []*big.Int {
	parts := strings.Split(b64, ".")
	ns := []*big.Int{}

	for _, v := range parts {
		ns = append(ns, B64ToBigInt(v))
	}

	return ns
}

func main() {
	rng := rand.New(rand.NewSource(time.Now().Unix()))
	gr := GetGroup()
	// funcs: gen pk

	flag.Parse()

	switch flag.Arg(0) {
	case "generate":
		pk := GetNewDHPrivateKey(rng, gr)
		fmt.Println(MarshalDHKey(pk))
		fmt.Println(BigIntToB64(pk.ga))
	case "get-r2":
		pk := UnmarshalDHKey(flag.Arg(1))
		i, _ := strconv.Atoi(flag.Arg(2))
		r1s := UnmarshalBigIntList(flag.Arg(3))

		fmt.Println(BigIntToB64(BurmesterDesmedtR2(gr, i, pk, r1s)))
	case "get-secret":
		pk := UnmarshalDHKey(flag.Arg(1))
		i, _ := strconv.Atoi(flag.Arg(2))
		r1s := UnmarshalBigIntList(flag.Arg(3))
		r2s := UnmarshalBigIntList(flag.Arg(4))

		fmt.Println(BigIntToB64(BurmesterDesmedtSecret(gr, i, pk, r1s, r2s)))
	}
}
