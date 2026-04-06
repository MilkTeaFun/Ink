package password

import "golang.org/x/crypto/bcrypt"

// BcryptHasher hashes and verifies passwords with bcrypt.
type BcryptHasher struct{}

// Hash converts a plaintext password into a bcrypt digest.
func (BcryptHasher) Hash(password string) (string, error) {
	payload, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(payload), nil
}

// Compare checks whether the plaintext password matches the stored digest.
func (BcryptHasher) Compare(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
