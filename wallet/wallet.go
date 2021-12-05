package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"

	"github.com/monostylegc/BabyDoge/utils"
)

type wallet struct {
	privateKey *ecdsa.PrivateKey
	Address    string
}

var w *wallet

const (
	fileName string = "babydog.wallet"
)

func hasWalletFile() bool {
	_, err := os.Stat(fileName)

	return !os.IsNotExist(err)
}

func createPrivateKey() *ecdsa.PrivateKey {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	utils.HandleErr(err)

	return privKey
}

func persistKey(privateKey *ecdsa.PrivateKey) {
	bytes, err := x509.MarshalECPrivateKey(privateKey)
	utils.HandleErr(err)

	err = os.WriteFile(fileName, bytes, 0644)
	utils.HandleErr(err)
}

func restoreKey() *ecdsa.PrivateKey {
	keyAsByte, err := os.ReadFile(fileName)
	utils.HandleErr(err)
	key, err := x509.ParseECPrivateKey(keyAsByte)
	utils.HandleErr(err)
	return key
}

func addressFromKey(key *ecdsa.PrivateKey) string {
	z := append(key.X.Bytes(), key.Y.Bytes()...)

	return fmt.Sprintf("%x", z)
}

func Sign(payload string, w *wallet) string {
	signAsByte, err := hex.DecodeString(payload)
	utils.HandleErr(err)
	r, s, err := ecdsa.Sign(rand.Reader, w.privateKey, signAsByte)
	utils.HandleErr(err)
	signature := append(r.Bytes(), s.Bytes()...)
	return fmt.Sprintf("%x", signature)
}

func restoreBigInts(payload string) (*big.Int, *big.Int, error) {
	bytes, err := hex.DecodeString(payload)
	if err != nil {
		return nil, nil, err
	}
	firstHalfByte := bytes[:len(bytes)/2]
	secondHalfByte := bytes[len(bytes)/2:]

	bigA, bigB := big.Int{}, big.Int{}

	bigA.SetBytes(firstHalfByte)
	bigB.SetBytes(secondHalfByte)

	return &bigA, &bigB, nil
}

func Verify(signature, payload, address string) bool {
	r, s, err := restoreBigInts(signature)
	utils.HandleErr(err)
	x, y, err := restoreBigInts(address)
	utils.HandleErr(err)

	publicKey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	payloadBytes, err := hex.DecodeString(payload)
	utils.HandleErr(err)
	ok := ecdsa.Verify(&publicKey, payloadBytes, r, s)

	return ok
}

//Singleton pattern
func Wallet() *wallet {
	if w == nil {
		w = &wallet{}
		//지갑이 있는가?
		if hasWalletFile() {
			//지갑이 있으면 파일에서 불러옴
			w.privateKey = restoreKey()
		} else {
			//지갑이 없으면 하나 생성한다.
			key := createPrivateKey()
			persistKey(key)
			w.privateKey = key
		}
		w.Address = addressFromKey(w.privateKey)
	}
	return w
}
