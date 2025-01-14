/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bls

import (
	"encoding/asn1"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254"
)

type PublicParams struct {
	Parties     []int
	PublicKeys  [][]byte
	ThresholdPK []byte
}

type Verifier struct {
	pks                []bn254.G2Affine
	tPK                bn254.G2Affine
	parties2EvalPoints map[uint16]int64
}

func (v *Verifier) Init(rawPP []byte) error {
	pp := &PublicParams{}
	if _, err := asn1.Unmarshal(rawPP, pp); err != nil {
		return err
	}

	v.pks = nil
	_, err := v.tPK.SetBytes(pp.ThresholdPK)
	if err != nil {
		return err
	}

	for _, rawPK := range pp.PublicKeys {
		var pk bn254.G2Affine
		if _, err := pk.SetBytes(rawPK); err != nil {
			return err
		} else {
			v.pks = append(v.pks, pk)
		}
	}

	v.parties2EvalPoints = make(map[uint16]int64)
	for i, p := range pp.Parties {
		v.parties2EvalPoints[uint16(p)] = int64(i + 1)
	}

	return nil
}

func (v *Verifier) AggregateSignatures(signatures [][]byte, signers []uint16) ([]byte, error) {
	if len(signers) == 0 {
		panic("no signers")
	}

	if len(signatures) != len(signers) {
		panic(fmt.Sprintf("%d signers but %d signatures", len(signers), len(signatures)))
	}

	sigs := make([]bn254.G1Affine, len(signatures))
	for i := 0; i < len(signatures); i++ {
		var sig bn254.G1Affine
		_, err := sig.SetBytes(signatures[i])
		if err != nil {
			return nil, err
		}
		sigs[i] = sig
	}

	evalPoints := make([]int64, len(signers))
	for i, signer := range signers {
		evalPoint, exists := v.parties2EvalPoints[signer]
		if !exists {
			panic(fmt.Sprintf("signature %d was signed by an unknown party %d", i, signer))
		}
		evalPoints[i] = evalPoint
	}

	res := localAggregateSignatures(sigs, evalPoints...).Bytes()
	return res[:], nil
}

func (v *Verifier) Verify(digest []byte, signature []byte) error {
	var sig bn254.G1Affine
	_, err := sig.SetBytes(signature)
	if err != nil {
		return err
	}

	return localVerify(&v.tPK, digest, &sig)
}
