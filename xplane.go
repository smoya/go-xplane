// Copyright 2016 Ornen. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xplane

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/ornen/go-xplane/messages"
)

const (
	datagramPrefixLength = 5
	messageLength        = 36
)

type XPlane struct {
	RemoteAddress string
	LocalAddress  string
	Messages      chan Message
	connection    *net.UDPConn
	receivePeriod *time.Duration
}

type Opt func(*XPlane)

func New(remoteAddress, localAddress string, opts ...Opt) XPlane {
	x := XPlane{
		RemoteAddress: remoteAddress,
		LocalAddress:  localAddress,
		Messages:      make(chan Message),
	}

	for _, o := range opts {
		o(&x)
	}

	return x
}

func ReceiveEvery(t time.Duration) Opt {
	return func(x *XPlane) {
		x.receivePeriod = &t
	}
}

func (x *XPlane) Receive() {
	serverAddr, err := net.ResolveUDPAddr("udp", x.LocalAddress)
	serverConn, err := net.ListenUDP("udp", serverAddr)

	if err != nil {
		panic(err)
	}

	defer serverConn.Close()

	var sequence uint64
	buf := make([]byte, 1024)
	if x.receivePeriod != nil {
		t := time.NewTicker(*x.receivePeriod)
		for {
			select {
			case <-t.C:
				x.readBuf(serverConn, buf, sequence)
			}
		}
	} else {
		for {
			x.readBuf(serverConn, buf, sequence)
			sequence = +1
		}
	}
}

func (x *XPlane) readBuf(c *net.UDPConn, b []byte, sequence uint64) {
	n, _, _ := c.ReadFromUDP(b)
	m := (n - datagramPrefixLength) / messageLength

	for i := 0; i < m; i++ {
		sentence := b[datagramPrefixLength+i*messageLength : datagramPrefixLength+(i+1)*messageLength]
		x.parse(sentence, sequence)
	}
}

func (x *XPlane) parse(sentence []byte, sequence uint64) {
	messageType := sentence[0]
	messageBuffer := bytes.NewBuffer(sentence[4:])

	messageData := make([]float32, 8)
	binary.Read(messageBuffer, binary.LittleEndian, &messageData)

	switch messageType {
	case messages.SpeedMessageType:
		x.Messages <- messages.NewSpeedMessage(sequence, messageData)
	case messages.GLoadMessageType:
		x.Messages <- messages.NewGLoadMessage(sequence, messageData)
	case messages.AngularVelocitiesMessageType:
		x.Messages <- messages.NewAngularVelocitiesMessage(sequence, messageData)
	case messages.PitchRollHeadingMessageType:
		x.Messages <- messages.NewPitchRollHeadingMessage(sequence, messageData)
	case messages.FlightControlMessageType:
		x.Messages <- messages.NewFlightControlMessage(sequence, messageData)
	case messages.GearsBrakesMessageType:
		x.Messages <- messages.NewGearsBrakesMessage(sequence, messageData)
	case messages.WeatherMessageType:
		x.Messages <- messages.NewWeatherMessage(sequence, messageData)
	case messages.LatLonAltMessageType:
		x.Messages <- messages.NewLatLonAltMessage(sequence, messageData)
	case messages.LocVelDistTraveledMessageType:
		x.Messages <- messages.NewLocVelDistTraveledMessage(sequence, messageData)
	case messages.BatteryAmperageMessageType:
		x.Messages <- messages.NewBatteryAmperageMessage(sequence, messageData)
	case messages.BatteryVoltageMessageType:
		x.Messages <- messages.NewBatteryVoltageMessage(sequence, messageData)
	case messages.EngineRPMMessageType:
		x.Messages <- messages.NewEngineRPMMessage(sequence, messageData)
	case messages.PropRPMMessageType:
		x.Messages <- messages.NewPropRPMMessage(sequence, messageData)
	case messages.PropPitchMessageType:
		x.Messages <- messages.NewPropPitchMessage(sequence, messageData)
	default:
		log.Println("Unknown message type: ", messageType)
	}
}

func (x *XPlane) Connect() {
	udpAddr, err := net.ResolveUDPAddr("udp", x.RemoteAddress)

	if err != nil {
		fmt.Println("Wrong address!")
		return
	}

	x.connection, err = net.DialUDP("udp", nil, udpAddr)
}

func (x *XPlane) Send(command Command) {
	commandData := command.Data()

	buf := new(bytes.Buffer)

	buf.Write([]byte{'D', 'A', 'T', 'A', 0})
	buf.Write([]byte{byte(command.Type()), 0, 0, 0})

	if err := binary.Write(buf, binary.LittleEndian, &commandData); err != nil {
		fmt.Println(err)
		return
	}

	x.connection.Write(buf.Bytes())
}
