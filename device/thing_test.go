package device

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var thingName = ThingName("test")
var region = Region("us-east-1")
var keyPair = KeyPair{
	CertificatePath:   "./certificates/cert.pem",
	PrivateKeyPath:    "./certificates/private.key",
	CACertificatePath: "./certificates/ca.pem",
}

type shadowStruct struct {
	State struct {
		Reported struct {
			Value int64 `json:"value"`
		} `json:"reported"`
	} `json:"state"`
}

func TestNewThing(t *testing.T) {
	thing, err := NewThing(keyPair, thingName, region)
	assert.NoError(t, err, "thing instance created without error")
	assert.NotNil(t, thing, "thing instance is not nil")
}

func TestThingShadow(t *testing.T) {
	thing, err := NewThing(keyPair, thingName, region)
	assert.NoError(t, err, "thing instance created without error")
	assert.NotNil(t, thing, "thing instance is not nil")

	thingShadowChann, err := thing.SubscribeForThingShadowChanges()
	assert.NoError(t, err, "received thing shadow subscription channel without error")

	data := time.Now().Unix()

	shadow := fmt.Sprintf(`{"state": {"reported": {"value": %d}}}`, data)

	err = thing.UpdateThingShadow(Shadow(shadow))
	assert.NoError(t, err, "thing shadow updated without error")

	updatedShadow, ok := <-thingShadowChann
	assert.True(t, ok, "the reading updated shadow channel was successful")

	unmarshaledUpdatedShadow := &shadowStruct{}

	err = json.Unmarshal(updatedShadow, unmarshaledUpdatedShadow)
	assert.NoError(t, err, "thing shadow payload unmarshaled without error")

	assert.Equal(t, data, unmarshaledUpdatedShadow.State.Reported.Value, "thing shadow update has consistent data")

	gottenShadow, err := thing.GetThingShadow()
	assert.NoError(t, err, "retrieved thing shadow without error")

	unmarshalledGottenShadow := &shadowStruct{}

	err = json.Unmarshal(gottenShadow, unmarshalledGottenShadow)
	assert.NoError(t, err, "retrieved thing shadow unmarshaling  without error")

	assert.Equal(t, data, unmarshalledGottenShadow.State.Reported.Value, "retrieved thing shadow has consistent data")
}
