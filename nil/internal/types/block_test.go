package types

import (
	"testing"

	"github.com/NilFoundation/nil/nil/internal/crypto/bls"
	"github.com/stretchr/testify/require"
)

func TestInvalidBlockNumber(t *testing.T) {
	t.Parallel()

	// The compiler doesn't allow (InvalidBlockNumber+1) because it's an overflow.
	// So, here's a trick not using constants.
	for i := range 3 {
		require.Equal(t, BlockNumber(i-1), InvalidBlockNumber+BlockNumber(i))
	}
}

func TestBlock_SignAndVerifySignature(t *testing.T) {
	t.Parallel()

	block := &Block{}

	// Generate keys
	privKey := bls.NewRandomKey()
	pubKey := privKey.PublicKey()
	pubKeys := []bls.PublicKey{pubKey}
	mask, err := bls.NewMask(pubKeys)
	require.NoError(t, err)
	require.NoError(t, mask.SetParticipants([]uint32{0}))

	blockHash := block.Hash(BaseShardId)

	// Sign the block
	sig, err := privKey.Sign(blockHash[:])
	require.NoError(t, err)
	sig, err = bls.AggregateSignatures([]bls.Signature{sig}, mask)
	require.NoError(t, err)
	sigBytes, err := sig.Marshal()
	require.NoError(t, err)

	block.Signature = &BlsAggregateSignature{
		Sig:  sigBytes,
		Mask: []byte{1},
	}

	// Check signature
	err = block.VerifySignature([]bls.PublicKey{pubKey}, BaseShardId)
	require.NoError(t, err)

	// Invalid public key
	invalidPrivKey := bls.NewRandomKey()
	invalidPubKey := invalidPrivKey.PublicKey()

	err = block.VerifySignature([]bls.PublicKey{invalidPubKey}, BaseShardId)
	require.ErrorContains(t, err, "invalid signature")

	// Empty public key
	err = block.VerifySignature(nil, BaseShardId)
	require.ErrorContains(t, err, "mismatching mask lengths")

	// Verify with empty signature
	block.Signature = &BlsAggregateSignature{}
	err = block.VerifySignature(nil, BaseShardId)
	require.ErrorContains(t, err, "not enough data")
}
