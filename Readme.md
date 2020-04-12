# AWS IoT SDK for Go lang
The aws-iot-device-sdk-go package allows developers to write Go lang applications which access the AWS IoT Platform via MQTT.
## Install
`go get "github.com/kuzemkon/aws-iot-device-sdk-go"`
## Example
```
package main
import (
    "github.com/kuzemkon/aws-iot-device-sdk-go/device"
    "fmt"
)

func main() {
    thing, err := device.NewThing(
        device.KeyPair{
            PrivateKeyPath: "path/to/private/key",
            CertificatePath: "path/to/certificate",
            CACertificatePath: "path/to/rootCA",
        },
        "xxxxxxxxxx.iot.us-east-1.amazonaws.com",  // AWS IoT endpoint
        device.ThingName("thing_name"),
    )
    if err != nil {
        panic(err)
    }

    s, err := thing.GetThingShadow()
    if err != nil {
        panic(err)
    }
    fmt.Println(s)

    shadowChan, err := thing.SubscribeForThingShadowChanges()
    if err != nil {
        panic(err)
    }

    for {
        select {
        case s, ok := <- shadowChan:
            if !ok {
                panic("failed to read from shadow channel")
            }
            fmt.Println(s)
        }
    }
}
```
## Reference
```
// NewThing returns a new instance of Thing
func NewThing(keyPair KeyPair, thingName ThingName, region Region) (*Thing, error)
```
```
// GetThingShadow gets the current thing shadow
func (t *Thing) GetThingShadow() (Shadow, error)
```
```
// UpdateThingShadow publish a message with new thing shadow
func (t *Thing) UpdateThingShadow(payload Shadow) error
```
```
// SubscribeForThingShadowChanges returns the channel with the shadow updates
func (t *Thing) SubscribeForThingShadowChanges() (chan Shadow, error) 
```
