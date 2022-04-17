package wikijs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/thanhpk/randstr"
)

type WikijsApiTestSuite struct {
	suite.Suite
	Host     string
	Username string
	Password string
	Client   *WikijsClient
}

func TestWikijsApiTestSuite(t *testing.T) {
	suite.Run(t, new(WikijsApiTestSuite))
}

func (suite *WikijsApiTestSuite) SetupSuite() {
	suite.Host = os.Getenv("WIKIJS_HOST")
	suite.Username = os.Getenv("WIKIJS_USERNAME")
	suite.Password = os.Getenv("WIKIJS_PASSWORD")
	suite.Client, _ = NewWikijsClient(suite.Host, suite.Username, suite.Password, true, 10, "")
	if assert.NotNil(suite.T(), suite.Client) {
		setupDone, err := suite.Client.SetupDone()
		assert.Equal(suite.T(), true, setupDone)
		assert.Nil(suite.T(), err)
	}
}

func (suite *WikijsApiTestSuite) TestSetApi() {

	enabled, err := suite.Client.apiEnabled()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), true, enabled, "API should be enabled in setup")

	err = suite.Client.setApi(true)
	assert.Nil(suite.T(), err)
	enabled, err = suite.Client.apiEnabled()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), true, enabled, "API should be still be enabled")

	err = suite.Client.setApi(false)
	assert.Nil(suite.T(), err)
	enabled, err = suite.Client.apiEnabled()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), false, enabled, "API should be disabled")

	// login again with new client to clear cookies and auth with creds.
	suite.Client, _ = NewWikijsClient(suite.Host, suite.Username, suite.Password, true, 10, "")

	err = suite.Client.setApi(true)
	assert.Nil(suite.T(), err)
	enabled, err = suite.Client.apiEnabled()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), true, enabled, "API should be enabled")
}

func (suite *WikijsApiTestSuite) TestCreateApiKey() {

	apiKeyName := "terraform_" + randstr.String(16)
	keyId, err := suite.Client.getApiKeyId(apiKeyName)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), -1, keyId, "API key should not exist")

	key, err := suite.Client.createApiKey(apiKeyName, "1y", true)
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), key)
	keyId, err = suite.Client.getApiKeyId(apiKeyName)
	assert.Nil(suite.T(), err)
	assert.NotEqual(suite.T(), -1, keyId, "API key should exist")

}

func (suite *WikijsApiTestSuite) TestRevokeApiKey() {

	apiKeyName := "terraform_" + randstr.String(16)
	key, err := suite.Client.createApiKey(apiKeyName, "1y", true)
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), key)

	revoked, err := suite.Client.isApiKeyRevoked(apiKeyName)
	assert.Equal(suite.T(), false, revoked, "API key should not be revoked")

	err = suite.Client.revokeApiKey(apiKeyName)
	assert.Nil(suite.T(), err)
	revoked, err = suite.Client.isApiKeyRevoked(apiKeyName)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), true, revoked, "API key should be revoked")

	err = suite.Client.revokeApiKey("does-not-exist")
	// odd that it succeedes for a keyID that doesn't exist
	assert.Nil(suite.T(), err)
}

func (suite *WikijsApiTestSuite) TestGetAuthenticationStrategies() {

	strategies, err := suite.Client.GetAuthenticationStrategies()
	assert.Nil(suite.T(), err)
	if assert.NotNil(suite.T(), strategies) {
		for i := range strategies.Data.Authentication.Strategies {
			assert.NotEmpty(suite.T(), strategies.Data.Authentication.Strategies[i].Key)
		}
	}
}

func (suite *WikijsApiTestSuite) TestGetActiveAuthenticationStrategies() {

	activeStrategies, err := suite.Client.GetActiveAuthenticationStrategies()
	assert.Nil(suite.T(), err)
	if assert.NotNil(suite.T(), activeStrategies) {
		for i := range activeStrategies.Data.Authentication.ActiveStrategies {
			assert.NotEmpty(suite.T(), activeStrategies.Data.Authentication.ActiveStrategies[i].Key)
		}
	}
}
