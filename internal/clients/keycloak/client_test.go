//go:build integration

package keycloakclient_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	keycloakclient "github.com/dndev-xx/go-ninja-chat/internal/clients/keycloak"
	"github.com/dndev-xx/go-ninja-chat/internal/testingh"
)

type KeycloakSuite struct {
	testingh.ContextSuite
	kc *keycloakclient.Client
}

func TestKeycloakSuite(t *testing.T) {
	suite.Run(t, new(KeycloakSuite))
}

func (s *KeycloakSuite) SetupSuite() {
	s.ContextSuite.SetupSuite()

	var err error
	s.kc, err = keycloakclient.New(keycloakclient.NewOptions(
		keycloakclient.WithBasePath(testingh.Config.KeycloakBasePath),
		keycloakclient.WithRealm(testingh.Config.KeycloakRealm),
		keycloakclient.WithClientID(testingh.Config.KeycloakClientID),
		keycloakclient.WithClientSecret(testingh.Config.KeycloakClientSecret),
		keycloakclient.WithDebugMode(true),
	))
	s.Require().NoError(err)
}

func (s *KeycloakSuite) TestIntrospectTokenAfterAuth() {
	token, err := s.kc.Auth(s.Ctx, testingh.Config.KeycloakTestUser, testingh.Config.KeycloakTestPassword)
	s.Require().NoError(err)
	s.T().Log(token.AccessToken)

	result, err := s.kc.IntrospectToken(s.Ctx, token.AccessToken)
	s.Require().NoError(err)
	s.True(result.Active)

	result, err = s.kc.IntrospectToken(s.Ctx, "abracadabra")
	s.Require().NoError(err)
	s.False(result.Active)
}

func (s *KeycloakSuite) TestIntrospectTokenWithInvalidSignature() {
	const token = `eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJUYVRrYVo2dHdSaG1fMnF6OGNwdDVBMjNmVkE2RkNXR29qd1hIaGhwYmJZIn0.eyJleHAiOjE5NjIwNTkwNjcsImlhdCI6MTY2MjA1ODc2NywianRpIjoiMzZhOThkMTItYTk0Ni00NmQ3LWJiZmMtYzE4ZWExNzA3Zjk2IiwiaXNzIjoiaHR0cDovL2xvY2FsaG9zdDozMDEwL3JlYWxtcy9CYW5rIiwiYXVkIjpbImNoYXQtdWktY2xpZW50IiwiYWNjb3VudCJdLCJzdWIiOiI3ZDY3YjE0ZC0yMjFlLTQ0OTktOWJlMi02NzA3ZDdkZjFhZGMiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJpbnRlZ3JhdGlvbi10ZXN0aW5nIiwic2Vzc2lvbl9zdGF0ZSI6ImE2YTg5MTNkLWQ2YmEtNDM2MC05ODJmLWUwNjBkYTVkOWIzNyIsImFjciI6IjEiLCJyZWFsbV9hY2Nlc3MiOnsicm9sZXMiOlsib2ZmbGluZV9hY2Nlc3MiLCJ1bWFfYXV0aG9yaXphdGlvbiIsImRlZmF1bHQtcm9sZXMtYmFuayJdfSwicmVzb3VyY2VfYWNjZXNzIjp7ImNoYXQtdWktY2xpZW50Ijp7InJvbGVzIjpbInN1cHBvcnQtY2hhdC1jbGllbnQiXX0sImFjY291bnQiOnsicm9sZXMiOlsibWFuYWdlLWFjY291bnQiLCJtYW5hZ2UtYWNjb3VudC1saW5rcyIsInZpZXctcHJvZmlsZSJdfX0sInNjb3BlIjoicHJvZmlsZSBlbWFpbCIsInNpZCI6ImE2YTg5MTNkLWQ2YmEtNDM2MC05ODJmLWUwNjBkYTVkOWIzNyIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJib25kMDA3IiwiZW1haWwiOiJib25kMDA3QHVrLmNvbSJ9.MvccWqlBsERq7CYuTY52AlIFi6RUqlvQok3IMVJCbnMdYLS4uDfuwNKc9mdP5WQd3HiwM78F09ibB9vGJEF8LGV-mYTUbNj1FlLBCfQAny96yhRXnADkyy2tzhfSyVaUoi3CXw65XeSgveip5XSgtyzJIFqZAGGkkAKeIB3Y4YLgr0wVowgQodrCh3mrEqnVrNwHZ533CQaaedVrZD3yyVLY1tVwa7un3sAaNBpn5d_7yqUNIA0iT7bW90U6gdzPVvdtZmwdgoDYifgNxDRzI-14-lfrrPwzkP71MOEocNr9iB2ecIEL3vUad6a61VrITJl_wux4Bxg9bEc80LU-mg` //nolint:lll,gosec // not real token
	result, err := s.kc.IntrospectToken(s.Ctx, token)
	s.Require().NoError(err)
	s.False(result.Active)
}
