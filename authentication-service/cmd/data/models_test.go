package data_test

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestFunc(t *testing.T) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("verysecret"), 12)
	if err != nil {
		log.Println(err)
		return
	}
	//fmt.Println("Hashed password:", string(hashedPassword))
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte("verysecret"))
	assert.Equal(t, err, nil)

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte("verysecrets"))
	assert.NotEqual(t, err, nil)
}
