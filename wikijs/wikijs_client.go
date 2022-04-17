package wikijs

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/thanhpk/randstr"
	"golang.org/x/net/publicsuffix"
)

type ApiError struct {
	Code    int
	Message string
}

func (e *ApiError) Error() string {
	return e.Message
}

type WikijsClient struct {
	host                string
	clientCredentials   *ClientCredentials
	retryablehttpClient *retryablehttp.Client
	configured          bool
	debug               bool
}

type ClientCredentials struct {
	AdminEmail string
	Password   string
	JwtToken   string
	ApiToken   string
	ApiKeyName string
}

func wikiJsClient(host string, clientTimeout int64, caCert string) (*WikijsClient, error) {
	clientCredentials := &ClientCredentials{}

	retryablehttpClient, err := newHttpClient(clientTimeout, caCert)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %v", err)
	}

	wikijsClient := WikijsClient{
		host:                host,
		clientCredentials:   clientCredentials,
		retryablehttpClient: retryablehttpClient,
		configured:          false,
	}

	response, err := wikijsClient.retryablehttpClient.Get(wikijsClient.host + "/")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to wikijs: %v", err)
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Wikijs returned HTTP status code: %v", response.StatusCode)
	}

	return &wikijsClient, nil
}

func NewWikijsClient(host, adminEmail, password string, initialSetup bool, clientTimeout int64, caCert string) (*WikijsClient, error) {
	wikijsClient, err := wikiJsClient(host, clientTimeout, caCert)
	if err != nil {
		return nil, err
	}

	if initialSetup {
		err = wikijsClient.setup(adminEmail, password)
		if err != nil {
			return nil, fmt.Errorf("failed to perform initial seup of wikijs: %v", err)
		}
	}

	err = wikijsClient.login(adminEmail, password)
	if err != nil {
		return nil, fmt.Errorf("failed to login to wikijs: %v", err)
	}

	err = wikijsClient.setApi(true)
	if err != nil {
		return nil, fmt.Errorf("failed to enable API to wikijs: %v", err)
	}

	apiKeyName := "terraform_" + randstr.String(16)
	key, err := wikijsClient.createApiKey(apiKeyName, "1y", true)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %v", err)
	}

	wikijsClient.clientCredentials.ApiToken = key
	wikijsClient.clientCredentials.ApiKeyName = apiKeyName
	wikijsClient.configured = true

	if tfLog, ok := os.LookupEnv("TF_LOG"); ok {
		if tfLog == "DEBUG" {
			wikijsClient.debug = true
		}
	}

	return wikijsClient, nil
}

func (wikijsClient *WikijsClient) setup(adminEmail, adminPassword string) error {
	setupData := Finalize{
		AdminEmail:           adminEmail,
		AdminPassword:        adminPassword,
		AdminPasswordConfirm: adminPassword,
		SiteUrl:              wikijsClient.host,
		Telemetry:            false,
	}
	setupDone, err := wikijsClient.SetupDone()
	if err != nil {
		return err
	}
	wikijsClient.clientCredentials.AdminEmail = adminEmail
	wikijsClient.clientCredentials.Password = adminPassword

	if !setupDone {
		response, _, err := wikijsClient.post("/finalize", setupData)
		if err != nil {
			return fmt.Errorf("Error POSTing to /finalize: %v", err)
		}
		var finalizeResultStruct FinalizeResultStruct
		err = json.Unmarshal(response, &finalizeResultStruct)
		if err != nil {
			return err
		}
		// race condition hitting this from multiple clients, especially during testing.
		// if it failed, wait and test the setup.
		if !finalizeResultStruct.Ok {
			time.Sleep(1 * time.Second)
			setupDone, err = wikijsClient.SetupDone()
		}
		if err != nil {
			return err
		}
	}

	setupDone, err = wikijsClient.SetupDone()
	if err != nil {
		return fmt.Errorf("Error confirming setup completed: %v", err)
	}
	return nil
}

func (wikijsClient *WikijsClient) postLogin(adminEmail, adminPassword string) (*LoginCredentials, error) {
	loginData := GraphQl{
		Variables: LoginVariables{
			Username: adminEmail,
			Password: adminPassword,
			Strategy: "local",
		},
		Query: "mutation ($username: String!, $password: String!, $strategy: String!) {\n  authentication {\n    login(username: $username, password: $password, strategy: $strategy) {\n      responseResult {\n        succeeded\n        errorCode\n        slug\n        message\n        __typename\n      }\n      jwt\n      mustChangePwd\n      mustProvideTFA\n      mustSetupTFA\n      continuationToken\n      redirect\n      tfaQRImage\n      __typename\n    }\n    __typename\n  }\n}\n",
	}

	response, _, err := wikijsClient.post("/graphql", loginData)
	if err != nil {
		return nil, err
	}

	var loginCredentials LoginCredentials
	err = json.Unmarshal(response, &loginCredentials)
	if err != nil {
		return nil, err
	}
	return &loginCredentials, nil
}

func (wikijsClient *WikijsClient) login(adminEmail, adminPassword string) error {
	loginCredentials, err := wikijsClient.postLogin(adminEmail, adminPassword)
	if err != nil {
		return err
	}

	if !loginCredentials.Data.Authentication.Login.ResponseResult.Succeeded {
		// We can hit a rate limit on login. If so, read the response and wait the required time before trying again.
		if len(loginCredentials.Errors) > 0 {
			var i int
			_, err := fmt.Sscanf(loginCredentials.Errors[0].Message, "Too many requests, please try again in %d seconds.", &i)
			if err == nil {
				time.Sleep(time.Duration(i) * time.Second)
				loginCredentials, err = wikijsClient.postLogin(adminEmail, adminPassword)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("Error logging in: %s", loginCredentials.Data.Authentication.Login.ResponseResult.Message)
			}
		} else {
			return fmt.Errorf("Error logging in: %s", loginCredentials.Data.Authentication.Login.ResponseResult.Message)
		}
	}

	wikijsClient.clientCredentials.JwtToken = loginCredentials.Data.Authentication.Login.Jwt

	cookie := &http.Cookie{
		Name:   "jwt",
		Value:  wikijsClient.clientCredentials.JwtToken,
		MaxAge: 300,
	}
	urlObj, err := url.Parse(wikijsClient.host)
	if err != nil {
		return err
	}
	wikijsClient.retryablehttpClient.HTTPClient.Jar.SetCookies(urlObj, []*http.Cookie{cookie})

	return nil
}

func (wikijsClient *WikijsClient) Cleanup() error {

	err := wikijsClient.revokeApiKey(wikijsClient.clientCredentials.ApiKeyName)
	if err == nil {
		wikijsClient.clientCredentials.ApiKeyName = ""
	}
	return err
}

func newHttpClient(clientTimeout int64, caCert string) (*retryablehttp.Client, error) {
	cookieJar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{},
		Proxy:           http.ProxyFromEnvironment,
	}

	if caCert != "" {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(caCert))
		transport.TLSClientConfig.RootCAs = caCertPool
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	retryClient.RetryWaitMin = time.Second * 1
	retryClient.RetryWaitMax = time.Second * 3

	// httpClient := retryClient.StandardClient()
	// httpClient.Timeout = time.Second * time.Duration(clientTimeout)
	// httpClient.Transport = transport
	// httpClient.Jar = cookieJar
	retryClient.HTTPClient.Timeout = time.Second * time.Duration(clientTimeout)
	retryClient.HTTPClient.Transport = transport
	retryClient.HTTPClient.Jar = cookieJar

	return retryClient, nil
}
func (wikijsClient *WikijsClient) RequiresSetup() (bool, error) {
	response, err := wikijsClient.retryablehttpClient.Get(wikijsClient.host + "/")
	if err != nil {
		return true, fmt.Errorf("failed to connect to wikijs: %v", err)
	}

	responseBody, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return true, fmt.Errorf("failed to read response body: %v", readErr)
	}

	if !strings.Contains(string(responseBody), "setup") {
		return false, nil
	}
	return true, nil
}

func (wikijsClient *WikijsClient) SetupDone() (bool, error) {
	requiresSetup, err := wikijsClient.RequiresSetup()
	if err != nil {
		return false, fmt.Errorf("failed to connect to wikijs: %v", err)
	}
	log.Printf("SetupDone, RequiresSetup %v", requiresSetup)
	if !requiresSetup {
		return true, nil
	}
	return false, nil
}

func (wikijsClient *WikijsClient) get(path string, params map[string]string) ([]byte, error) {
	resourceUrl := wikijsClient.host + path

	request, err := retryablehttp.NewRequest(http.MethodGet, resourceUrl, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		query := url.Values{}
		for k, v := range params {
			query.Add(k, v)
		}
		request.URL.RawQuery = query.Encode()
	}

	body, _, err := wikijsClient.sendRequest(request, nil)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (wikijsClient *WikijsClient) post(path string, requestBody interface{}) ([]byte, string, error) {
	resourceUrl := wikijsClient.host + path

	request, err := retryablehttp.NewRequest(http.MethodPost, resourceUrl, nil)
	if err != nil {
		return nil, "", err
	}

	payload, err := wikijsClient.marshal(requestBody)
	if err != nil {
		return nil, "", err
	}

	body, location, err := wikijsClient.sendRequest(request, payload)
	if err != nil {
		return nil, "", err
	}
	return body, location, nil
}

func (wikijsClient *WikijsClient) marshal(requestBody interface{}) ([]byte, error) {

	return json.Marshal(requestBody)
}

func (wikijsClient *WikijsClient) addRequestHeaders(request *retryablehttp.Request) {
	request.Header.Set("Accept", "application/json")

	if wikijsClient.clientCredentials.ApiToken != "" {
		request.Header.Set("Authorization", "Bearer "+wikijsClient.clientCredentials.ApiToken)
	}

	// if wikijsClient.userAgent != "" {
	// 	request.Header.Set("User-Agent", wikijsClient.userAgent)
	// }

	if request.Method == http.MethodPost || request.Method == http.MethodPut || request.Method == http.MethodDelete {
		request.Header.Set("Content-type", "application/json")
	}
}

func (wikijsClient *WikijsClient) sendRequest(request *retryablehttp.Request, body []byte) ([]byte, string, error) {

	requestMethod := request.Method
	requestPath := request.URL.Path

	log.Printf("[DEBUG] Sending %s to %s", requestMethod, requestPath)
	if body != nil {
		request.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	wikijsClient.addRequestHeaders(request)
	response, err := wikijsClient.retryablehttpClient.Do(request)
	if err != nil {
		log.Printf("[DEBUG] failed doing Do: %s", err)
		return nil, "", fmt.Errorf("error sending request: %v", err)
	}

	defer response.Body.Close()

	responseBody, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		return nil, "", err2
	}

	if response.StatusCode >= 400 {
		errorMessage := fmt.Sprintf("error sending %s request to %s: %s.", request.Method, request.URL.Path, response.Status)

		if len(responseBody) != 0 {
			errorMessage = fmt.Sprintf("%s Response body: %s", errorMessage, responseBody)
		}

		return nil, "", &ApiError{
			Code:    response.StatusCode,
			Message: errorMessage,
		}
	}

	return responseBody, response.Header.Get("Location"), nil
}
