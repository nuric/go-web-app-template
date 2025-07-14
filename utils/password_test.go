package utils_test

import (
	"testing"

	"github.com/nuric/go-api-template/utils"
	"github.com/stretchr/testify/require"
)

func TestVerifyPassword(t *testing.T) {
	hashed := utils.HashPassword("testpassword")
	require.NotEmpty(t, hashed, "hashed password should not be empty")
	require.True(t, utils.VerifyPassword(hashed, "testpassword"), "should verify correct password")
	require.False(t, utils.VerifyPassword(hashed, "wrongpassword"), "should not verify incorrect password")
}
