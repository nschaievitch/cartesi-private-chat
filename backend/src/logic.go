package main

import (
	"math/big"
	"strings"

	"github.com/google/uuid"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

var groups map[string]Group = make(map[string]Group)

var stateTransitions map[string][]StateTransition = make(map[string][]StateTransition)

type SignatureError struct {
	message string
}

type UnauthorizedError struct {
	message string
}

type GroupNotFoundError struct {
	message string
}

func (e *SignatureError) Error() string {
	return e.message
}

func (e *UnauthorizedError) Error() string {
	return e.message
}

func (e *GroupNotFoundError) Error() string {
	return e.message
}

type Address string

type Group struct {
	Members      []Address
	R1           []*string
	R2           []*string
	GroupAddress Address
}

type StateTransition struct {
	Action    string
	Author    Address
	Timestamp int64
}

// func VerifySignature(content string, address Address, signature string) bool {
// 	return VerifyRSASignatureFromB64(content, signature, UnmarshalPubKey(string(address)))
// }

func VerifySignature(content string, address Address, signature string) bool {
	return true
	sig := hexutil.MustDecode(signature)

	hash := crypto.Keccak256Hash([]byte(content)).Bytes()

	if sig[crypto.RecoveryIDOffset] == 27 || sig[crypto.RecoveryIDOffset] == 28 {
		sig[crypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1
	}

	recovered, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return false
	}

	recoveredAddr := crypto.PubkeyToAddress(*recovered)

	return string(address) == recoveredAddr.Hex()
}

func findAddress(addresses []Address, address Address) int {
	for i, a := range addresses {
		if strings.ToLower(string(a)) == strings.ToLower(string(address)) {
			return i
		}
	}

	return -1
}

func all(list []*string) bool {
	for _, v := range list {
		if v == nil {
			return false
		}
	}

	return true
}

func CreateGroup(members []Address) (Group, string) {
	var group Group
	group.Members = members
	groupSize := len(members)

	group.R1 = make([]*string, groupSize)
	group.R2 = make([]*string, groupSize)

	id := uuid.NewString()
	return group, id
}

func SubmitR1(id string, r1Value *big.Int, address Address, signature string) error {
	if !VerifySignature(BigIntToB64(r1Value), address, signature) {
		return &SignatureError{"invalid signature"}
	}

	group, exists := groups[id]

	if !exists {
		return &GroupNotFoundError{"group not found"}
	}

	ind := findAddress(group.Members, address)

	if ind == -1 {
		return &UnauthorizedError{"address not in group"}
	}

	s := BigIntToB64(r1Value)
	group.R1[ind] = &s
	return nil
}

func SubmitR2(id string, r2Value *big.Int, address Address, signature string) error {
	if !VerifySignature(BigIntToB64(r2Value), address, signature) {
		return &SignatureError{"invalid signature"}
	}

	group, exists := groups[id]

	if !exists {
		return &GroupNotFoundError{"group not found"}
	}

	ind := findAddress(group.Members, address)

	if ind == -1 {
		return &UnauthorizedError{"address not in group"}
	}

	s := BigIntToB64(r2Value)
	group.R2[ind] = &s

	return nil
}

func SubmitGroupAddress(id string, groupAddress string, address Address, signature string) error {
	if !VerifySignature(groupAddress, address, signature) {
		return &SignatureError{"invalid signature"}
	}

	group, exists := groups[id]

	if !exists {
		return &GroupNotFoundError{"group not found"}
	}

	ind := findAddress(group.Members, address)

	if ind == -1 {
		return &UnauthorizedError{"address not in group"}
	}

	group.GroupAddress = Address(groupAddress)

	return nil
}

func R1Complete(id string) bool {
	group, exists := groups[id]

	if !exists {
		return false
	}

	return all(group.R1)
}

func R2Complete(id string) bool {
	group, exists := groups[id]

	if !exists {
		return false
	}

	return all(group.R2)
}

func SubmitTransition(id string, action string, address Address, timestamp int64) error {
	group, exists := groups[id]

	if !exists {
		return &GroupNotFoundError{"group not found"}
	}

	ind := findAddress(group.Members, address)

	if ind == -1 {
		return &UnauthorizedError{"address not in group"}
	}

	st := StateTransition{
		action,
		address,
		timestamp,
	}

	groupTransitions, _ := stateTransitions[id]

	stateTransitions[id] = append(groupTransitions, st)

	return nil
}
