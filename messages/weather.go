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

package messages

type WeatherMessage struct {
	Message
	Pressure    float64
	Temperature float64
}

func NewWeatherMessage(sequence uint64, data []float32) WeatherMessage {
	return WeatherMessage{
		Message: Message{
			sequence: sequence,
		},
		Pressure:    float64(data[0]),
		Temperature: float64(data[1]),
	}
}

func (m WeatherMessage) Type() uint {
	return WeatherMessageType
}
