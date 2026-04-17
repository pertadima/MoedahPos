package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestHash(t *testing.T) {
	password := "my-secure-password"

	t.Run("Bcrypt Password Flow", func(t *testing.T) {
		hashed, err := Password(password, bcrypt.DefaultCost)
		assert.NoError(t, err)
		assert.NotEmpty(t, hashed)

		err = CheckPassword(password, hashed)
		assert.NoError(t, err)

		err = CheckPassword("wrong-password", hashed)
		assert.Error(t, err)
	})

	t.Run("SHA256", func(t *testing.T) {
		s := "secret-string"
		h1 := SHA256(s)
		h2 := SHA256(s)
		assert.Equal(t, h1, h2)
		assert.NotEmpty(t, h1)
		assert.NotEqual(t, s, h1)
	})
}
