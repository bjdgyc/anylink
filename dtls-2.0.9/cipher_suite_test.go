package dtls

import (
	"context"
	"testing"
	"time"

	"github.com/pion/dtls/v2/internal/ciphersuite"
	"github.com/pion/dtls/v2/internal/net/dpipe"
	"github.com/pion/transport/test"
)

func TestCipherSuiteName(t *testing.T) {
	testCases := []struct {
		suite    CipherSuiteID
		expected string
	}{
		{TLS_ECDHE_ECDSA_WITH_AES_128_CCM, "TLS_ECDHE_ECDSA_WITH_AES_128_CCM"},
		{CipherSuiteID(0x0000), "0x0000"},
	}

	for _, testCase := range testCases {
		res := CipherSuiteName(testCase.suite)
		if res != testCase.expected {
			t.Fatalf("Expected: %s, got %s", testCase.expected, res)
		}
	}
}

func TestAllCipherSuites(t *testing.T) {
	actual := len(allCipherSuites())
	if actual == 0 {
		t.Fatal()
	}
}

// CustomCipher that is just used to assert Custom IDs work
type testCustomCipherSuite struct {
	ciphersuite.TLSEcdheEcdsaWithAes128GcmSha256
	authenticationType CipherSuiteAuthenticationType
}

func (t *testCustomCipherSuite) ID() CipherSuiteID {
	return 0xFFFF
}

func (t *testCustomCipherSuite) AuthenticationType() CipherSuiteAuthenticationType {
	return t.authenticationType
}

// Assert that two connections that pass in a CipherSuite with a CustomID works
func TestCustomCipherSuite(t *testing.T) {
	type result struct {
		c   *Conn
		err error
	}

	// Check for leaking routines
	report := test.CheckRoutines(t)
	defer report()

	runTest := func(cipherFactory func() []CipherSuite) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ca, cb := dpipe.Pipe()
		c := make(chan result)

		go func() {
			client, err := testClient(ctx, ca, &Config{
				CipherSuites:       []CipherSuiteID{},
				CustomCipherSuites: cipherFactory,
			}, true)
			c <- result{client, err}
		}()

		server, err := testServer(ctx, cb, &Config{
			CipherSuites:       []CipherSuiteID{},
			CustomCipherSuites: cipherFactory,
		}, true)

		clientResult := <-c

		if err != nil {
			t.Error(err)
		} else {
			_ = server.Close()
		}

		if clientResult.err != nil {
			t.Error(clientResult.err)
		} else {
			_ = clientResult.c.Close()
		}
	}

	t.Run("Custom ID", func(t *testing.T) {
		runTest(func() []CipherSuite {
			return []CipherSuite{&testCustomCipherSuite{authenticationType: CipherSuiteAuthenticationTypeCertificate}}
		})
	})

	t.Run("Anonymous Cipher", func(t *testing.T) {
		runTest(func() []CipherSuite {
			return []CipherSuite{&testCustomCipherSuite{authenticationType: CipherSuiteAuthenticationTypeAnonymous}}
		})
	})
}
