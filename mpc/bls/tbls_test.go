package bls

import (
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"

	"github.com/stretchr/testify/assert"
)

func TestLocalSignVerify(t *testing.T) {
	randomBytes := make([]byte, fr.Modulus().BitLen())
	rand.Read(randomBytes)

	sk := big.NewInt(0)
	sk.SetBytes(randomBytes)
	sk.Mod(sk, fr.Modulus())

	h := sha256.New()
	h.Write([]byte("the little fox jumps over the lazy dog"))
	digest := h.Sum(nil)

	sig := localSign(sk, digest)
	pk := makePublicKey(sk)
	assert.NoError(t, localVerify(pk, digest, sig))

	h = sha256.New()
	h.Write([]byte("the little fox hops over the lazy dog"))
	digest2 := h.Sum(nil)

	assert.EqualError(t, localVerify(pk, digest2, sig), "signature mismatch")

	sig2 := localSign(sk, digest2)

	assert.EqualError(t, localVerify(pk, digest, sig2), "signature mismatch")
}

func TestLocalThresholdBLS(t *testing.T) {
	shares := localGen(3, 2)
	pks := localCreatePublicKeys(shares)

	digest := sha256.Sum256([]byte("the little fox jumps over the lazy dog"))

	var signatures []bn254.G1Affine
	for i := 0; i < len(shares); i++ {
		signatures = append(signatures, *localSign(shares[i], digest[:]))
	}

	for i := 0; i < len(shares); i++ {
		assert.NoError(t, localVerify(&pks[i], digest[:], &signatures[i]))
	}

	thresholdSignature := localAggregateSignatures(signatures[:2], 1, 2)
	thresholdPK := localAggregatePublicKeys(pks, 1, 2)

	assert.NoError(t, localVerify(thresholdPK, digest[:], thresholdSignature))
}
