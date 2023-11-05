package main

import (
	"fmt"
	"math/big"
	"math/rand"
)

type DHGroup struct {
	p *big.Int
	g *big.Int
}

type DHPrivateKey struct {
	a  *big.Int
	ga *big.Int
}

func GetGroup() DHGroup {
	p, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B576625E7EC6F44C42E9A63A3620FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7EDEE386BFB5A899FA5AE9F24117C4B1FE649286651ECE65381FFFFFFFFFFFFFFFF", 16)
	//p, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B576625E7EC6F44C42E9A63A3620FFFFFFFFFFFFFFFF", 16)
	g := big.NewInt(2)

	return DHGroup{p, g}
}

func GetNewDHPrivateKey(rng *rand.Rand, gr DHGroup) (pk DHPrivateKey) {
	n := GetRandomOddNumber(rng)
	pk.a = big.NewInt(0).Mod(n, gr.p)
	pk.ga = new(big.Int).Exp(gr.g, pk.a, gr.p)

	return
}

func GetDHSharedSecret(gr DHGroup, pk DHPrivateKey, gb *big.Int) (secret *big.Int) {
	secret = big.NewInt(1)
	secret.Exp(gb, pk.a, gr.p)
	return
}

func BurmesterDesmedtR2(gr DHGroup, index int, pk DHPrivateKey, r1 []*big.Int) (zi *big.Int) {
	partyCount := len(r1)

	y_i_next := r1[(index+1)%partyCount]
	y_i_previous_inv := new(big.Int).Set(r1[(partyCount+index-1)%partyCount])
	y_i_previous_inv.ModInverse(y_i_previous_inv, gr.p)

	zi = big.NewInt(1).Mul(y_i_next, y_i_previous_inv)
	zi.Mod(zi, gr.p)

	zi.Exp(zi, pk.a, gr.p)

	return zi
}

func BurmesterDesmedtSecret(gr DHGroup, index int, pk DHPrivateKey, r1 []*big.Int, r2 []*big.Int) (secret *big.Int) {
	partyCount := len(r1)
	n := big.NewInt(int64(partyCount))
	one := big.NewInt(1)

	y_i_previous := r1[(partyCount+index-1)%partyCount]

	K := big.NewInt(1)
	K.Exp(y_i_previous, n, gr.p)
	K.Exp(K, pk.a, gr.p)

	for i := 0; i < partyCount-1; i++ {
		n.Sub(n, one)

		t := big.NewInt(1)
		t.Exp(r2[(index+i)%partyCount], n, gr.p)

		K.Mul(K, t)
		K.Mod(K, gr.p)
	}

	return K
}

func TestDH(rng *rand.Rand) {
	SIZE = 1024
	gr := GetGroup()

	pk1 := GetNewDHPrivateKey(rng, gr)
	pk2 := GetNewDHPrivateKey(rng, gr)

	ss1 := GetDHSharedSecret(gr, pk1, pk2.ga)
	ss2 := GetDHSharedSecret(gr, pk2, pk1.ga)

	fmt.Println(BigIntToB64(ss1), BigIntToB64(ss2))
}

func TestBurmesterDesmedt(rng *rand.Rand) {
	n := 5
	gr := GetGroup()

	keys := []DHPrivateKey{}
	r1 := []*big.Int{}

	for i := 0; i < n; i++ {
		pk := GetNewDHPrivateKey(rng, gr)
		keys = append(keys, pk)
		r1 = append(r1, pk.ga)
	}

	r2 := []*big.Int{}

	for i := 0; i < n; i++ {
		r2 = append(r2, BurmesterDesmedtR2(gr, i, keys[i], r1))
	}

	for i := 0; i < n; i++ {
		secret := BurmesterDesmedtSecret(gr, i, keys[i], r1, r2)
		fmt.Println(BigIntToB64(secret))
	}

}
