package credentials

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Service is dedicated to get the AWS credentials based on the device X509 certificates. The retrieved credentials
// can be used to access any AWS Service.
//
// More info here: https://aws.amazon.com/blogs/security/how-to-eliminate-the-need-for-hardcoded-aws-credentials-in-devices-by-using-the-aws-iot-credentials-provider/
type Service struct {
	url       string
	thingName string
	tlsCert   tls.Certificate
}

// Output the AWS credentials output data structure
type Output struct {
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
	Expiration      string `json:"expiration"`
}

// NewService initializes the device certificates based on the provided paths and returns a new instance of the Service.
//
// The iotCredentialsURL parameter should satisfy this pattern:
// https://<your_credentials_provider_endpoint>/role-aliases/<your-role-alias>/credentials
//
// More info here: https://aws.amazon.com/blogs/security/how-to-eliminate-the-need-for-hardcoded-aws-credentials-in-devices-by-using-the-aws-iot-credentials-provider/
func NewService(iotCredentialsURL, certPath, privateKeyPath, thingName string) (Service, error) {
	tlsCert, err := tls.LoadX509KeyPair(certPath, privateKeyPath)
	if err != nil {
		return Service{}, fmt.Errorf("failed to load the certificates: %v", err)
	}

	return Service{
		url:       iotCredentialsURL,
		thingName: thingName,
		tlsCert:   tlsCert,
	}, nil
}

// GetCredentials performs the HTTPS request authorized by the device TLS certificates in order to get the AWS credentials.
// Returns the Output object with the AWS credentials
func (s Service) GetCredentials() (Output, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{s.tlsCert},
			},
		},
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("GET", s.url, nil)
	if err != nil {
		return Output{}, fmt.Errorf("failed to create the credentials request: %v", err)
	}

	req.Header.Add("x-amzn-iot-thingname", s.thingName)

	resp, err := client.Do(req)
	if err != nil {
		return Output{}, fmt.Errorf("failed to perform the GET credentials request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return Output{}, fmt.Errorf("failed to parse the response body: %v", err)
		}

		return Output{}, fmt.Errorf("the request has failed with the status code: %d; message: %s", resp.StatusCode, string(body))
	}

	result := struct {
		Credentials Output `json:"credentials"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Output{}, fmt.Errorf("failed to parse credentials response body: %v", err)
	}

	return result.Credentials, nil
}
