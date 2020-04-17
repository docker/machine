package azureutil

import (
	"math/rand"
	"time"
)

/* Utilities */

// randomAzureStorageAccountName generates a valid storage account name prefixed
// with a predefined string. Availability of the name is not checked. Uses maximum
// length to maximise randomness.
func randomAzureStorageAccountName() string {
	const (
		maxLen = 24
		chars  = "0123456789abcdefghijklmnopqrstuvwxyz"
	)
	return storageAccountPrefix + randomString(maxLen-len(storageAccountPrefix), chars)
}

// randomString generates a random string of given length using specified alphabet.
func randomString(n int, alphabet string) string {
	r := timeSeed()
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[r.Intn(len(alphabet))]
	}
	return string(b)
}

func timeSeed() *rand.Rand { return rand.New(rand.NewSource(time.Now().UTC().UnixNano())) }
