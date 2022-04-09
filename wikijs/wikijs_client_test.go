package wikijs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type WikijsClientTestSuite struct {
	suite.Suite
	Host     string
	Username string
	Password string
	Client   *WikijsClient
}

func TestWikijsClientTestSuite(t *testing.T) {
	suite.Run(t, new(WikijsClientTestSuite))
}

func (suite *WikijsClientTestSuite) SetupSuite() {
	suite.Host = os.Getenv("WIKIJS_HOST")
	suite.Username = os.Getenv("WIKIJS_USERNAME")
	suite.Password = os.Getenv("WIKIJS_PASSWORD")

	client, err := NewWikijsClient(suite.Host, suite.Username, suite.Password, true, 10, "")
	assert.Nil(suite.T(), err)
	if assert.NotNil(suite.T(), client) {
		suite.Client = client
		setupDone, err := suite.Client.SetupDone()
		assert.Equal(suite.T(), true, setupDone)
		assert.Nil(suite.T(), err)
	}
}

func (suite *WikijsClientTestSuite) TesWikiJsClient() {
	client, err := wikiJsClient(suite.Host, 10, "")
	assert.Nil(suite.T(), err)

	client, err = wikiJsClient("http://does.not.exist", 1, "")
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), client)
}

func (suite *WikijsClientTestSuite) TestLogin() {
	client, err := wikiJsClient(suite.Host, 10, "")
	assert.NotNil(suite.T(), client)

	// client, err := NewWikijsClient(suite.Host, suite.Username, suite.Password, true, 10, "")
	// assert.Nil(suite.T(), err)
	// if assert.NotNil(suite.T(), client) {
	// 	enabled, err := client.apiEnabled()
	// 	assert.Nil(suite.T(), err)
	// 	assert.Equal(suite.T(), true, enabled, "API should be enabled")
	// }

	err = client.login(suite.Username, suite.Password)
	assert.Nil(suite.T(), err)

	err = client.login("invalid_user", suite.Password)
	assert.NotNil(suite.T(), err)

	err = client.login(suite.Username, "incorrect_password")
	assert.NotNil(suite.T(), err)
}

func (suite *WikijsClientTestSuite) TestMarshal() {
	apiError := &ApiError{
		Code:    1025,
		Message: "Here is a message",
	}
	marshalledMessage, err := suite.Client.marshal(apiError)
	assert.Nil(suite.T(), err)
	if assert.NotNil(suite.T(), marshalledMessage) {
		expectedString := "{\"Code\":1025,\"Message\":\"Here is a message\"}"
		assert.Equal(suite.T(), []byte(expectedString), marshalledMessage)
	}
}

func (suite *WikijsClientTestSuite) TestCleanup() {
	if suite.Client.configured {
		apiKeyName := suite.Client.clientCredentials.ApiKeyName
		id, err := suite.Client.getApiKeyId(apiKeyName)
		assert.Nil(suite.T(), err)
		assert.NotNil(suite.T(), id)

		err = suite.Client.Cleanup()
		assert.Nil(suite.T(), err)

		id, err = suite.Client.getApiKeyId(apiKeyName)
		assert.NotNil(suite.T(), err)
		assert.Equal(suite.T(), -1, id)
	}
}
