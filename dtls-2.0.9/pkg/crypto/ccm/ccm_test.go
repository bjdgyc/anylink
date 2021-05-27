package ccm

// Refer to RFC 3610 section 8 for the vectors.

import (
	"bytes"
	"crypto/aes"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
)

func mustHexDecode(s string) []byte {
	r, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return r
}

var (
	aesKey1to12  = mustHexDecode("c0c1c2c3c4c5c6c7c8c9cacbcccdcecf") //nolint:gochecknoglobals
	aesKey13to24 = mustHexDecode("d7828d13b2b0bdc325a76236df93cc6b") //nolint:gochecknoglobals
)

// AESKey: AES Key
// CipherText: Authenticated and encrypted output
// ClearHeaderOctets:  Input with X cleartext header octets
// Data: Input with X cleartext header octets
// M: length(CBC-MAC)
// Nonce: Nonce
type vector struct {
	AESKey            []byte
	CipherText        []byte
	ClearHeaderOctets int
	Data              []byte
	M                 int
	Nonce             []byte
}

func TestRFC3610Vectors(t *testing.T) {
	cases := []vector{
		// Vectors 1-12
		{
			AESKey:            aesKey1to12,
			CipherText:        mustHexDecode("0001020304050607588c979a61c663d2f066d0c2c0f989806d5f6b61dac38417e8d12cfdf926e0"),
			ClearHeaderOctets: 8,
			Data:              mustHexDecode("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e"),
			M:                 8,
			Nonce:             mustHexDecode("00000003020100a0a1a2a3a4a5"),
		},
		{
			AESKey:            aesKey1to12,
			CipherText:        mustHexDecode("000102030405060772c91a36e135f8cf291ca894085c87e3cc15c439c9e43a3ba091d56e10400916"),
			ClearHeaderOctets: 8,
			Data:              mustHexDecode("000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F"),
			M:                 8,
			Nonce:             mustHexDecode("00000004030201a0a1a2a3a4a5"),
		},
		{
			AESKey:            aesKey1to12,
			CipherText:        mustHexDecode("000102030405060751b1e5f44a197d1da46b0f8e2d282ae871e838bb64da8596574adaa76fbd9fb0c5"),
			ClearHeaderOctets: 8,
			Data:              mustHexDecode("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			M:                 8,
			Nonce:             mustHexDecode("00000005040302a0a1a2a3a4a5"),
		},
		{
			AESKey:            aesKey1to12,
			CipherText:        mustHexDecode("000102030405060708090a0ba28c6865939a9a79faaa5c4c2a9d4a91cdac8c96c861b9c9e61ef1"),
			ClearHeaderOctets: 12,
			Data:              mustHexDecode("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e"),
			M:                 8,
			Nonce:             mustHexDecode("00000006050403a0a1a2a3a4a5"),
		},
		{
			AESKey:            aesKey1to12,
			CipherText:        mustHexDecode("000102030405060708090a0bdcf1fb7b5d9e23fb9d4e131253658ad86ebdca3e51e83f077d9c2d93"),
			ClearHeaderOctets: 12,
			Data:              mustHexDecode("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"),
			M:                 8,
			Nonce:             mustHexDecode("00000007060504a0a1a2a3a4a5"),
		},
		{
			AESKey:            aesKey1to12,
			CipherText:        mustHexDecode("000102030405060708090a0b6fc1b011f006568b5171a42d953d469b2570a4bd87405a0443ac91cb94"),
			ClearHeaderOctets: 12,
			Data:              mustHexDecode("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			M:                 8,
			Nonce:             mustHexDecode("00000008070605a0a1a2a3a4a5"),
		},
		{
			AESKey:            aesKey1to12,
			CipherText:        mustHexDecode("00010203040506070135d1b2c95f41d5d1d4fec185d166b8094e999dfed96c048c56602c97acbb7490"),
			ClearHeaderOctets: 8,
			Data:              mustHexDecode("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e"),
			M:                 10,
			Nonce:             mustHexDecode("00000009080706a0a1a2a3a4a5"),
		},
		{
			AESKey:            aesKey1to12,
			CipherText:        mustHexDecode("00010203040506077b75399ac0831dd2f0bbd75879a2fd8f6cae6b6cd9b7db24c17b4433f434963f34b4"),
			ClearHeaderOctets: 8,
			Data:              mustHexDecode("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"),
			M:                 10,
			Nonce:             mustHexDecode("0000000a090807a0a1a2a3a4a5"),
		},
		{
			AESKey:            aesKey1to12,
			CipherText:        mustHexDecode("000102030405060782531a60cc24945a4b8279181ab5c84df21ce7f9b73f42e197ea9c07e56b5eb17e5f4e"),
			ClearHeaderOctets: 8,
			Data:              mustHexDecode("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			M:                 10,
			Nonce:             mustHexDecode("0000000b0a0908a0a1a2a3a4a5"),
		},
		{
			AESKey:            aesKey1to12,
			CipherText:        mustHexDecode("000102030405060708090a0b07342594157785152b074098330abb141b947b566aa9406b4d999988dd"),
			ClearHeaderOctets: 12,
			Data:              mustHexDecode("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e"),
			M:                 10,
			Nonce:             mustHexDecode("0000000c0b0a09a0a1a2a3a4a5"),
		},
		{
			AESKey:            aesKey1to12,
			CipherText:        mustHexDecode("000102030405060708090a0b676bb20380b0e301e8ab79590a396da78b834934f53aa2e9107a8b6c022c"),
			ClearHeaderOctets: 12,
			Data:              mustHexDecode("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"),
			M:                 10,
			Nonce:             mustHexDecode("0000000d0c0b0aa0a1a2a3a4a5"),
		},
		{
			AESKey:            aesKey1to12,
			CipherText:        mustHexDecode("000102030405060708090a0bc0ffa0d6f05bdb67f24d43a4338d2aa4bed7b20e43cd1aa31662e7ad65d6db"),
			ClearHeaderOctets: 12,
			Data:              mustHexDecode("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			M:                 10,
			Nonce:             mustHexDecode("0000000e0d0c0ba0a1a2a3a4a5"),
		},
		// Vectors 13-24
		{
			AESKey:            aesKey13to24,
			CipherText:        mustHexDecode("0be1a88bace018b14cb97f86a2a4689a877947ab8091ef5386a6ffbdd080f8e78cf7cb0cddd7b3"),
			ClearHeaderOctets: 8,
			Data:              mustHexDecode("0be1a88bace018b108e8cf97d820ea258460e96ad9cf5289054d895ceac47c"),
			M:                 8,
			Nonce:             mustHexDecode("00412b4ea9cdbe3c9696766cfa"),
		},
		{
			AESKey:            aesKey13to24,
			CipherText:        mustHexDecode("63018f76dc8a1bcb4ccb1e7ca981befaa0726c55d378061298c85c92814abc33c52ee81d7d77c08a"),
			ClearHeaderOctets: 8,
			Data:              mustHexDecode("63018f76dc8a1bcb9020ea6f91bdd85afa0039ba4baff9bfb79c7028949cd0ec"),
			M:                 8,
			Nonce:             mustHexDecode("0033568ef7b2633c9696766cfa"),
		},
		{
			AESKey:            aesKey13to24,
			CipherText:        mustHexDecode("aa6cfa36cae86b40b1d23a2220ddc0ac900d9aa03c61fcf4a559a4417767089708a776796edb723506"),
			ClearHeaderOctets: 8,
			Data:              mustHexDecode("aa6cfa36cae86b40b916e0eacc1c00d7dcec68ec0b3bbb1a02de8a2d1aa346132e"),
			M:                 8,
			Nonce:             mustHexDecode("00103fe41336713c9696766cfa"),
		},
		{
			AESKey:            aesKey13to24,
			CipherText:        mustHexDecode("d0d0735c531e1becf049c24414d253c3967b70609b7cbb7c499160283245269a6f49975bcadeaf"),
			ClearHeaderOctets: 12,
			Data:              mustHexDecode("d0d0735c531e1becf049c24412daac5630efa5396f770ce1a66b21f7b2101c"),
			M:                 8,
			Nonce:             mustHexDecode("00764c63b8058e3c9696766cfa"),
		},
		{
			AESKey:            aesKey13to24,
			CipherText:        mustHexDecode("77b60f011c03e1525899bcae5545ff1a085ee2efbf52b2e04bee1e2336c73e3f762c0c7744fe7e3c"),
			ClearHeaderOctets: 12,
			Data:              mustHexDecode("77b60f011c03e1525899bcaee88b6a46c78d63e52eb8c546efb5de6f75e9cc0d"),
			M:                 8,
			Nonce:             mustHexDecode("00f8b678094e3b3c9696766cfa"),
		},
		{
			AESKey:            aesKey13to24,
			CipherText:        mustHexDecode("cd9044d2b71fdb8120ea60c0009769ecabdf48625594c59251e6035722675e04c847099e5ae0704551"),
			ClearHeaderOctets: 12,
			Data:              mustHexDecode("cd9044d2b71fdb8120ea60c06435acbafb11a82e2f071d7ca4a5ebd93a803ba87f"),
			M:                 8,
			Nonce:             mustHexDecode("00d560912d3f703c9696766cfa"),
		},
		{
			AESKey:            aesKey13to24,
			CipherText:        mustHexDecode("d85bc7e69f944fb8bc218daa947427b6db386a99ac1aef23ade0b52939cb6a637cf9bec2408897c6ba"),
			ClearHeaderOctets: 8,
			Data:              mustHexDecode("d85bc7e69f944fb88a19b950bcf71a018e5e6701c91787659809d67dbedd18"),
			M:                 10,
			Nonce:             mustHexDecode("0042fff8f1951c3c9696766cfa"),
		},
		{
			AESKey:            aesKey13to24,
			CipherText:        mustHexDecode("74a0ebc9069f5b375810e6fd25874022e80361a478e3e9cf484ab04f447efff6f0a477cc2fc9bf548944"),
			ClearHeaderOctets: 8,
			Data:              mustHexDecode("74a0ebc9069f5b371761433c37c5a35fc1f39f406302eb907c6163be38c98437"),
			M:                 10,
			Nonce:             mustHexDecode("00920f40e56cdc3c9696766cfa"),
		},
		{
			AESKey:            aesKey13to24,
			CipherText:        mustHexDecode("44a3aa3aae6475caf2beed7bc5098e83feb5b31608f8e29c38819a89c8e776f1544d4151a4ed3a8b87b9ce"),
			ClearHeaderOctets: 8,
			Data:              mustHexDecode("44a3aa3aae6475caa434a8e58500c6e41530538862d686ea9e81301b5ae4226bfa"),
			M:                 10,
			Nonce:             mustHexDecode("0027ca0c7120bc3c9696766cfa"),
		},
		{
			AESKey:            aesKey13to24,
			CipherText:        mustHexDecode("ec46bb63b02520c33c49fd7031d750a09da3ed7fddd49a2032aabf17ec8ebf7d22c8088c666be5c197"),
			ClearHeaderOctets: 12,
			Data:              mustHexDecode("ec46bb63b02520c33c49fd70b96b49e21d621741632875db7f6c9243d2d7c2"),
			M:                 10,
			Nonce:             mustHexDecode("005b8ccbcd9af83c9696766cfa"),
		},
		{
			AESKey:            aesKey13to24,
			CipherText:        mustHexDecode("47a65ac78b3d594227e85e71e882f1dbd38ce3eda7c23f04dd65071eb41342acdf7e00dccec7ae52987d"),
			ClearHeaderOctets: 12,
			Data:              mustHexDecode("47a65ac78b3d594227e85e71e2fcfbb880442c731bf95167c8ffd7895e337076"),
			M:                 10,
			Nonce:             mustHexDecode("003ebe94044b9a3c9696766cfa"),
		},
		{
			AESKey:            aesKey13to24,
			CipherText:        mustHexDecode("6e37a6ef546d955d34ab6059f32905b88a641b04b9c9ffb58cc390900f3da12ab16dce9e82efa16da62059"),
			ClearHeaderOctets: 12,
			Data:              mustHexDecode("6e37a6ef546d955d34ab6059abf21c0b02feb88f856df4a37381bce3cc128517d4"),
			M:                 10,
			Nonce:             mustHexDecode("008d493b30ae8b3c9696766cfa"),
		},
	}

	if len(cases) != 24 {
		t.Fatalf("Expected %d test cases, got: %d", 24, len(cases))
		t.FailNow()
	}

	for idx, c := range cases {
		c := c
		t.Run(fmt.Sprintf("packet vector #%d", idx+1), func(t *testing.T) {
			t.Parallel()
			blk, err := aes.NewCipher(c.AESKey)
			if err != nil {
				t.Fatalf("could not initialize AES block cipher from key: %v", err)
			}

			lccm, err := NewCCM(blk, c.M, len(c.Nonce))
			if err != nil {
				t.Fatalf("could not create CCM: %v", err)
			}

			t.Run("seal", func(t *testing.T) {
				var dst []byte
				dst = lccm.Seal(dst, c.Nonce, c.Data[c.ClearHeaderOctets:], c.Data[:c.ClearHeaderOctets])
				if !bytes.Equal(c.CipherText[c.ClearHeaderOctets:], dst) {
					t.Fatalf("ciphertext does not match, wanted %v, got %v",
						c.CipherText[c.ClearHeaderOctets:], dst)
				}
			})

			t.Run("open", func(t *testing.T) {
				var dst []byte
				dst, err = lccm.Open(dst, c.Nonce, c.CipherText[c.ClearHeaderOctets:], c.CipherText[:c.ClearHeaderOctets])
				if err != nil {
					t.Fatalf("failed to unseal: %v", err)
				}
				if !bytes.Equal(c.Data[c.ClearHeaderOctets:], dst) {
					t.Fatalf("plaintext does not match, wanted %v, got %v",
						c.Data[c.ClearHeaderOctets:], dst)
				}
			})
		})
	}
}

func TestNewCCMError(t *testing.T) {
	cases := map[string]struct {
		vector
		err error
	}{
		"ShortNonceLength": {
			vector{
				AESKey: aesKey1to12,
				M:      8,
				Nonce:  mustHexDecode("a0a1a2a3a4a5"),
			}, errInvalidNonceSize,
		},
		"LongNonceLength": {
			vector{
				AESKey: aesKey1to12,
				M:      8,
				Nonce:  mustHexDecode("0001020304050607080910111213"),
			}, errInvalidNonceSize,
		},
		"ShortTag": {
			vector{
				AESKey: aesKey1to12,
				M:      3,
				Nonce:  mustHexDecode("00010203040506070809101112"),
			}, errInvalidTagSize,
		},
		"LongTag": {
			vector{
				AESKey: aesKey1to12,
				M:      17,
				Nonce:  mustHexDecode("00010203040506070809101112"),
			}, errInvalidTagSize,
		},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			blk, err := aes.NewCipher(c.AESKey)
			if err != nil {
				t.Fatalf("could not initialize AES block cipher from key: %v", err)
			}

			if _, err := NewCCM(blk, c.M, len(c.Nonce)); !errors.Is(err, c.err) {
				t.Fatalf("expected error '%v', got '%v'", c.err, err)
			}
		})
	}
}

func TestSealError(t *testing.T) {
	cases := map[string]struct {
		vector
		err error
	}{
		"InvalidNonceLength": {
			vector{
				Data:  mustHexDecode("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e"),
				M:     8,
				Nonce: mustHexDecode("00000003020100a0a1a2a3a4"), // short
			}, errInvalidNonceSize,
		},
		"PlaintextTooLong": {
			vector{
				Data:  make([]byte, 100000),
				M:     8,
				Nonce: mustHexDecode("00000003020100a0a1a2a3a4a5"),
			}, errPlaintextTooLong,
		},
	}

	blk, err := aes.NewCipher(aesKey1to12)
	if err != nil {
		t.Fatalf("could not initialize AES block cipher from key: %v", err)
	}

	lccm, err := NewCCM(blk, 8, 13)
	if err != nil {
		t.Fatalf("could not create CCM: %v", err)
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			defer func() {
				if err := recover(); !errors.Is(err.(error), c.err) {
					t.Errorf("expected panic '%v', got '%v'", c.err, err)
				}
			}()
			var dst []byte
			_ = lccm.Seal(dst, c.Nonce, c.Data[c.ClearHeaderOctets:], c.Data[:c.ClearHeaderOctets])
		})
	}
}

func TestOpenError(t *testing.T) {
	cases := map[string]struct {
		vector
		err error
	}{
		"CiphertextTooShort": {
			vector{
				CipherText:        make([]byte, 10),
				ClearHeaderOctets: 8,
				Nonce:             mustHexDecode("00000003020100a0a1a2a3a4a5"),
			}, errCiphertextTooShort,
		},
		"CiphertextTooLong": {
			vector{
				CipherText:        make([]byte, 100000),
				ClearHeaderOctets: 8,
				Nonce:             mustHexDecode("00000003020100a0a1a2a3a4a5"),
			}, errCiphertextTooLong,
		},
	}

	blk, err := aes.NewCipher(aesKey1to12)
	if err != nil {
		t.Fatalf("could not initialize AES block cipher from key: %v", err)
	}

	lccm, err := NewCCM(blk, 8, 13)
	if err != nil {
		t.Fatalf("could not create CCM: %v", err)
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			var dst []byte
			_, err = lccm.Open(dst, c.Nonce, c.CipherText[c.ClearHeaderOctets:], c.CipherText[:c.ClearHeaderOctets])
			if !errors.Is(err, c.err) {
				t.Errorf("expected error '%v', got '%v'", c.err, err)
			}
		})
	}
}
