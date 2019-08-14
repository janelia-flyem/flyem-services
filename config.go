package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	ClientID            string            `json:"oauthclient-id"`              // google oauth client id
	ClientSecret        string            `json:"oauthclient-secret"`          // google oauth client secret
	Secret              string            `json:"appsecret"`                   // password for token and cookie generation
	Hostname            string            `json:"hostname"`                    // name of server
	SwaggerFile         string            `json:"swagger-docs"`                // location of swagger file
	ApplicationsSecrets map[string]string `json:"applications-secrets"`        // toekn password for each supported application
	ApplicationsAuth    map[string]string `json:"applications-auth,omitempty"` // optional authorization file names for each application when creating a token
	CertPEM             string            `json:"ssl-cert,omitempty"`          // https certificate
	KeyPEM              string            `json:"ssl-key,omitempty"`           // https private key
	LoggerFile          string            `json:"log-file,omitempty"`          // location for log file
}

// LoadConfig parses json configuration and loads options
func LoadConfig(configFile string) (config Config, err error) {
	// open json file
	jsonFile, err := os.Open(configFile)
	if err != nil {
		err = fmt.Errorf("%s cannot be read", configFile)
		return
	}
	byteData, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteData, &config)
	return
}
