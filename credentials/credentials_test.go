package credentials

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var thingName = ""
var url = ""
var certPath = "./certificates/cert.pem"
var privateKeyPath = "./certificates/private.key"

func TestMain(m *testing.M) {
	var ok bool

	thingName, ok = os.LookupEnv("AWS_IOT_THING_NAME")
	if !ok {
		panic("AWS_IOT_THING_NAME environment variable must be defined")
	}

	url, ok = os.LookupEnv("AWS_IOT_CREDENTIALS_URL")
	if !ok {
		panic("AWS_MQTT_ENDPOINT environment variable must be defined")
	}

	code := m.Run()
	os.Exit(code)
}

func TestService_GetCredentials(t *testing.T) {
	s, err := NewService(url, certPath, privateKeyPath, thingName)
	assert.NoError(t, err, "credentials service created without error")

	out, err := s.GetCredentials()
	assert.NoError(t, err, "credentials retrieved without error")

	assert.NotEmpty(t, out.AccessKeyId, "the retrieved accessKeyId is not empty")
	assert.NotEmpty(t, out.SecretAccessKey, "the retrieved secretAccessKey is not empty")
	assert.NotEmpty(t, out.SessionToken, "the retrieved sessionToken is not empty")
	assert.NotEmpty(t, out.Expiration, "the retrieved expiration is not empty")

	fmt.Println(out)
}
