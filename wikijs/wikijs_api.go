package wikijs

import (
	"encoding/json"
	"fmt"
	"time"
)

type FinalizeResultStruct struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type ResponseStatus struct {
	Succeeded bool   `json:"succeeded"`
	ErrorCode int    `json:"errorCode"`
	Slug      string `json:"slug"`
	Message   string `json:"message"`
}

type Finalize struct {
	AdminEmail           string `json:"adminEmail"`
	AdminPassword        string `json:"adminPassword"`
	AdminPasswordConfirm string `json:"adminPasswordConfirm"`
	SiteUrl              string `json:"siteUrl"`
	Telemetry            bool   `json:"telemetry"`
}

type LoginVariables struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Strategy string `json:"strategy"`
}

type GraphQl struct {
	OperationName interface{} `json:"operationName,omitempty"`
	Variables     interface{} `json:"variables,omitempty"`
	Extensions    struct {
	} `json:"extensions,omitempty"`
	Query string `json:"query"`
}

type ApiVariables struct {
	Enabled bool `json:"enabled"`
}

type CreateApiKeyVariables struct {
	Name       string `json:"name"`
	Expiration string `json:"expiration"`
	FullAccess bool   `json:"fullAccess"`
	Group      string `json:"group,omitempty"`
}

type ApiKeyVariables struct {
	Id int `json:"id"`
}

type ResponseResultStruct struct {
	Succeeded bool   `json:"succeeded"`
	ErrorCode int    `json:"errorCode"`
	Slug      string `json:"slug"`
	Message   string `json:"message"`
	Typename  string `json:"__typename"`
}

type LoginCredentials struct {
	Data struct {
		Authentication struct {
			Login struct {
				ResponseResult    ResponseResultStruct `json:"responseResult"`
				Jwt               string               `json:"jwt"`
				MustChangePwd     interface{}          `json:"mustChangePwd"`
				MustProvideTFA    interface{}          `json:"mustProvideTFA"`
				MustSetupTFA      interface{}          `json:"mustSetupTFA"`
				ContinuationToken interface{}          `json:"continuationToken"`
				Redirect          string               `json:"redirect"`
				TfaQRImage        interface{}          `json:"tfaQRImage"`
				Typename          string               `json:"__typename"`
			} `json:"login"`
			Typename string `json:"__typename"`
		} `json:"authentication"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type ApiCredentials struct {
	Data struct {
		Authentication struct {
			CreateAPIKey struct {
				Key            string               `json:"key"`
				ResponseResult ResponseResultStruct `json:"responseResult"`
				Typename       string               `json:"__typename"`
			} `json:"createApiKey"`
			Typename string `json:"__typename"`
		} `json:"authentication"`
	} `json:"data"`
}

type GetApiKeys struct {
	Data struct {
		Authentication struct {
			APIKeys []struct {
				ID         int       `json:"id"`
				Name       string    `json:"name"`
				KeyShort   string    `json:"keyShort"`
				Expiration time.Time `json:"expiration"`
				IsRevoked  bool      `json:"isRevoked"`
				CreatedAt  time.Time `json:"createdAt"`
				UpdatedAt  time.Time `json:"updatedAt"`
				Typename   string    `json:"__typename"`
			} `json:"apiKeys"`
			Typename string `json:"__typename"`
		} `json:"authentication"`
	} `json:"data"`
}

type ApiState struct {
	Data struct {
		Authentication struct {
			APIState bool `json:"apiState"`
		} `json:"authentication"`
	} `json:"data"`
}

type KeyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type AuthenticationStrategy struct {
	Key          string         `json:"key"`
	Props        []KeyValuePair `json:"props"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	IsAvailable  bool           `json:"isAvailable"`
	UseForm      bool           `json:"useForm"`
	UsernameType string         `json:"usernameType"`
	Logo         string         `json:"logo"`
	Color        string         `json:"color"`
	Website      string         `json:"website"`
	Icon         string         `json:"icon"`
	Typename     string         `json:"__typename"`
}

type AuthenticationStrategies struct {
	Data struct {
		Authentication struct {
			Strategies []AuthenticationStrategy `json:"strategies"`
			Typename   string                   `json:"__typename"`
		} `json:"authentication"`
	} `json:"data"`
}

type ActiveAuthenticationStrategies struct {
	Data struct {
		Authentication struct {
			ActiveStrategies []struct {
				Key              string                 `json:"key"`
				Strategy         AuthenticationStrategy `json:"strategies"`
				DisplayName      string                 `json:"displayName"`
				Order            int                    `json:"order"`
				IsEnabled        bool                   `json:"isEnabled"`
				Config           []KeyValuePair         `json:"config"`
				SelfRegistration bool                   `json:"selfRegistration"`
				DomainWhitelist  []string               `json:"domainWhitelist"`
				AutoEnrollGroups []int                  `json:"autoEntrollGroups"`
				Typename         string                 `json:"__typename"`
			} `json:"activeStrategies"`
			Typename string `json:"__typename"`
		} `json:"authentication"`
	} `json:"data"`
}

func (wikijsClient *WikijsClient) apiEnabled() (bool, error) {

	getApiData := GraphQl{
		Query: `
{
	authentication {
		apiState
	}
}`,
	}

	response, _, err := wikijsClient.post("/graphql", getApiData)
	if err != nil {
		// if api is disabled then we wont be able to get a result :)
		// Check if site is up. IF so then return no error and false for apiEnabled?
		_, err = wikijsClient.retryablehttpClient.Get(wikijsClient.host + "/")
		if err != nil {
			return false, nil
		}
		return false, err
	}

	var apiState ApiState
	err = json.Unmarshal(response, &apiState)

	return apiState.Data.Authentication.APIState, nil
}

func (wikijsClient *WikijsClient) setApi(enable bool) error {

	apiEnabled, err := wikijsClient.apiEnabled()
	if err != nil {
		return err
	}

	if apiEnabled == enable {
		// no need to toggle
		return nil
	}

	setApiSateData := GraphQl{
		Variables: ApiVariables{
			Enabled: enable,
		},
		Query: `
mutation ($enabled: Boolean!) {
	authentication {
	    setApiState(enabled: $enabled) {
			responseResult {
			    succeeded
				errorCode
		        slug
		        message
		        __typename
	        }
	        __typename
	    }
	    __typename
	}
}`,
	}

	_, _, err = wikijsClient.post("/graphql", setApiSateData)
	return err
}

func (wikijsClient *WikijsClient) createApiKey(apiKeyName string, expiration string, fullAccess bool) (string, error) {
	createApiKeyData := GraphQl{
		Variables: CreateApiKeyVariables{
			Name:       apiKeyName,
			Expiration: expiration,
			FullAccess: fullAccess,
		},
		Query: `
mutation ($name: String!, $expiration: String!, $fullAccess: Boolean!, $group: Int) {
	authentication {
	    createApiKey(name: $name, expiration: $expiration, fullAccess: $fullAccess, group: $group) {
	        key
			responseResult {
		        succeeded
		        errorCode
		        slug
		        message
		        __typename
	        }
	        __typename
	    }
	    __typename
    }
}`,
	}

	response, _, err := wikijsClient.post("/graphql", createApiKeyData)
	if err != nil {
		return "", err
	}

	var apiCredentials ApiCredentials
	err = json.Unmarshal(response, &apiCredentials)
	if err != nil {
		return "", err
	}

	return apiCredentials.Data.Authentication.CreateAPIKey.Key, nil
}

func (wikijsClient *WikijsClient) getApiKeys() (*GetApiKeys, error) {

	getApiKeyData := GraphQl{
		Query: `
{
	authentication {
		apiKeys {
			id
			name
			keyShort
			expiration
			isRevoked
			createdAt
			updatedAt
			__typename
		}
		__typename
	}
}`,
	}

	response, _, err := wikijsClient.post("/graphql", getApiKeyData)
	if err != nil {
		return nil, err
	}

	var getApiKeys GetApiKeys
	err = json.Unmarshal(response, &getApiKeys)
	if err != nil {
		return nil, err
	}

	return &getApiKeys, nil
}

func (wikijsClient *WikijsClient) getApiKeyId(name string) (int, error) {
	getApiKeys, err := wikijsClient.getApiKeys()
	if err != nil {
		return -1, err
	}

	for i := range getApiKeys.Data.Authentication.APIKeys {
		if getApiKeys.Data.Authentication.APIKeys[i].Name == name {
			return getApiKeys.Data.Authentication.APIKeys[i].ID, nil
		}
	}

	return -1, fmt.Errorf("Did not find API key with name: %s", name)
}

func (wikijsClient *WikijsClient) isApiKeyRevoked(name string) (bool, error) {
	getApiKeys, err := wikijsClient.getApiKeys()
	if err != nil {
		return false, err
	}

	for i := range getApiKeys.Data.Authentication.APIKeys {
		if getApiKeys.Data.Authentication.APIKeys[i].Name == name {
			return getApiKeys.Data.Authentication.APIKeys[i].IsRevoked, nil
		}
	}

	return false, fmt.Errorf("Did not find API key with name: %s", name)
}
func (wikijsClient *WikijsClient) revokeApiKey(name string) error {
	id, err := wikijsClient.getApiKeyId(name)

	revokeApiKeyData := GraphQl{
		Variables: ApiKeyVariables{
			Id: id,
		},
		Query: `
mutation ($id: Int!) {
	authentication {
        revokeApiKey(id: $id) {
	        responseResult {
				succeeded
				errorCode
				slug
				message
				__typename
			}
			__typename
		}
		__typename
    }
}`,
	}

	_, _, err = wikijsClient.post("/graphql", revokeApiKeyData)
	if err != nil {
		return err
	}
	return nil
}

func (wikijsClient *WikijsClient) GetAuthenticationStrategies() (*AuthenticationStrategies, error) {

	getAuthenticationStrategiesData := GraphQl{
		Query: `
{
	authentication {
		strategies {
			key
			props {
				key
				value
			}
			title
			description
			isAvailable
			useForm
			usernameType
			logo
			color
			website
			icon
			__typename
		}
		__typename
	}
}`,
	}

	response, _, err := wikijsClient.post("/graphql", getAuthenticationStrategiesData)
	if err != nil {
		return nil, err
	}

	var authenticationStrategies AuthenticationStrategies
	err = json.Unmarshal(response, &authenticationStrategies)
	if err != nil {
		return nil, err
	}

	return &authenticationStrategies, nil
}

func (wikijsClient *WikijsClient) GetActiveAuthenticationStrategies() (*ActiveAuthenticationStrategies, error) {

	getActiveAuthenticationStrategiesData := GraphQl{
		Query: `
{
	authentication {
		activeStrategies {
			key
			strategy {
				key
				props {
					key
					value
				}
				title
				description
				isAvailable
				useForm
				usernameType
				logo
				color
				website
				icon
			}
			displayName
			order
			isEnabled
			config{
				key
				value
			}
			selfRegistration
			domainWhitelist
			autoEnrollGroups			
			__typename
		}
		__typename
	}
}`,
	}

	response, _, err := wikijsClient.post("/graphql", getActiveAuthenticationStrategiesData)
	if err != nil {
		return nil, err
	}

	var activeAuthenticationStrategies ActiveAuthenticationStrategies
	err = json.Unmarshal(response, &activeAuthenticationStrategies)
	if err != nil {
		return nil, err
	}

	return &activeAuthenticationStrategies, nil
}

// {
//     "operationName": null,
//     "variables": {
//         "strategies": [
//             {
//                 "key": "local",
//                 "strategyKey": "local",
//                 "displayName": "Local",
//                 "order": 0,
//                 "isEnabled": true,
//                 "config": [],
//                 "selfRegistration": false,
//                 "domainWhitelist": [],
//                 "autoEnrollGroups": []
//             },
//             {
//                 "key": "09c4f24c-7c92-49d5-937c-06356062fb4c",
//                 "strategyKey": "keycloak",
//                 "displayName": "Keycloak",
//                 "order": 1,
//                 "isEnabled": true,
//                 "config": [
//                     {
//                         "key": "host",
//                         "value": "{\"v\":\"https://your.keycloak-host.com\"}"
//                     },
//                     {
//                         "key": "realm",
//                         "value": "{\"v\":\"master\"}"
//                     },
//                     {
//                         "key": "clientId",
//                         "value": "{\"v\":\"clientid\"}"
//                     },
//                     {
//                         "key": "clientSecret",
//                         "value": "{\"v\":\"secrete\"}"
//                     },
//                     {
//                         "key": "authorizationURL",
//                         "value": "{\"v\":\"https://your.keycloak-host.com/auth/realms/master/protocol/openid-connect/auth\"}"
//                     },
//                     {
//                         "key": "tokenURL",
//                         "value": "{\"v\":\"https://your.keycloak-host.com/auth/realms/master/protocol/openid-connect/token\"}"
//                     },
//                     {
//                         "key": "userInfoURL",
//                         "value": "{\"v\":\"https://your.keycloak-host.com/auth/realms/master/protocol/openid-connect/userinfo\"}"
//                     },
//                     {
//                         "key": "logoutUpstream",
//                         "value": "{\"v\":true}"
//                     },
//                     {
//                         "key": "logoutURL",
//                         "value": "{\"v\":\"https://your.keycloak-host.com/auth/realms/master/protocol/openid-connect/logout\"}"
//                     }
//                 ],
//                 "selfRegistration": true,
//                 "domainWhitelist": [],
//                 "autoEnrollGroups": []
//             }
//         ]
//     },
//     "extensions": {},
//     "query": "mutation ($strategies: [AuthenticationStrategyInput]!) {\n  authentication {\n    updateStrategies(strategies: $strategies) {\n      responseResult {\n        succeeded\n        errorCode\n        slug\n        message\n        __typename\n      }\n      __typename\n    }\n    __typename\n  }\n}\n"
// }
