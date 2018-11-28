package device

import (
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Thing a structure for working with the AWS IoT device shadows
type Thing struct {
	client    mqtt.Client
	thingName ThingName
}

// ThingName the name of the AWS IoT device representation
type ThingName string

// Region the AWS region
type Region string

// KeyPair the structure contains the path to the AWS MQTT credentials
type KeyPair struct {
	PrivateKeyPath  string
	CertificatePath string
}

// Shadow device shadow data
type Shadow []byte

// NewThing returns a new instance of Thing
func NewThing(keyPair KeyPair, thingName ThingName, region Region) (*Thing, error) {
	tlsCert, err := tls.LoadX509KeyPair(keyPair.CertificatePath, keyPair.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	awsServerURL := fmt.Sprintf("ssl://data.iot.%s.amazonaws.com:8883", region)

	mqttOpts := mqtt.NewClientOptions()
	mqttOpts.AddBroker(awsServerURL)
	mqttOpts.SetMaxReconnectInterval(1 * time.Second)
	mqttOpts.SetClientID(string(thingName))
	mqttOpts.SetTLSConfig(&tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	})

	c := mqtt.NewClient(mqttOpts)
	token := c.Connect()
	token.Wait()

	return &Thing{
		client:    c,
		thingName: thingName,
	}, token.Error()
}

// GetThingShadow gets the current thing shadow
func (t *Thing) GetThingShadow() (Shadow, error) {
	shadowChan := make(chan Shadow)
	errChan := make(chan error)

	if token := t.client.Subscribe(
		fmt.Sprintf("$aws/things/%s/shadow/get/accepted", t.thingName),
		0,
		func(client mqtt.Client, msg mqtt.Message) {
			shadowChan <- msg.Payload()
		},
	); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	if token := t.client.Subscribe(
		fmt.Sprintf("$aws/things/%s/shadow/get/rejected", t.thingName),
		0,
		func(client mqtt.Client, msg mqtt.Message) {
			errChan <- errors.New(string(msg.Payload()))
		},
	); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	if token := t.client.Publish(
		fmt.Sprintf("$aws/things/%s/shadow/get", t.thingName),
		0,
		false,
		[]byte("{}"),
	); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	for {
		select {
		case s, ok := <-shadowChan:
			if !ok {
				return nil, errors.New("failed to read from shadow channel")
			}
			return s, nil
		case err, ok := <-errChan:
			if !ok {
				return nil, errors.New("failed to read from error channel")
			}
			return nil, err
		}
	}
}

// UpdateThingShadow publish a message with new thing shadow
func (t *Thing) UpdateThingShadow(payload Shadow) error {
	token := t.client.Publish(fmt.Sprintf("$aws/things/%s/shadow/update", t.thingName), 0, false, []byte(payload))
	token.Wait()
	return token.Error()
}

// SubscribeForThingShadowChanges returns the channel with the shadow updates
func (t *Thing) SubscribeForThingShadowChanges() (chan Shadow, error) {
	shadowChan := make(chan Shadow)

	token := t.client.Subscribe(
		fmt.Sprintf("$aws/things/%s/shadow/update/accepted", t.thingName),
		0,
		func(client mqtt.Client, msg mqtt.Message) {
			shadowChan <- msg.Payload()
		},
	)
	token.Wait()

	return shadowChan, token.Error()
}
