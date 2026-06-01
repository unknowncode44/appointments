package test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mousav1/ticket/internal/token"
	"github.com/mousav1/ticket/internal/util"
	"github.com/stretchr/testify/require"
)

func TestJWTMaker(t *testing.T) {
	maker, err := token.NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	username := util.RandomUsername()
	role := "tenantUser"
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, payload, err := maker.CreateToken(username, role, nil, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.Equal(t, role, payload.Role)
	require.Nil(t, payload.TenantID)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)

}

func TestExpiredJWTToken(t *testing.T) {
	maker, err := token.NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	token, payload, err := maker.CreateToken(util.RandomUsername(), "user", nil, -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.Error(t, err)
	require.Nil(t, payload)
}

func TestInvalidJWTTokenAlgNone(t *testing.T) {
	payload, err := token.NewPayload(util.RandomUsername(), "user", nil, time.Minute)
	require.NoError(t, err)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	getToken, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	maker, err := token.NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	payload, err = maker.VerifyToken(getToken)
	require.Error(t, err)
	require.Nil(t, payload)
}
