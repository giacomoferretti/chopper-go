/*
 * Copyright 2021 Giacomo Ferretti
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"reflect"
	"strconv"
	"testing"
)

func TestParseChannelsString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output []int
	}{
		{
			name: "correct",
			input:  "1,2,3",
			output: []int{1, 2, 3},
		},
		{
			name: "invalid_value",
			input:  "0",
			output: []int{},
		},
		{
			name: "commas_suffix",
			input:  "1,,",
			output: []int{1},
		},
		{
			name: "commas_prefix",
			input:  ",,3",
			output: []int{3},
		},
		{
			name: "words_prefix",
			input:  "asd1,2,3",
			output: []int{1, 2, 3},
		},
		{
			name: "words_suffix",
			input:  "1asd,2,3",
			output: []int{1, 2, 3},
		},
		{
			name: "words_in_between",
			input:  "1a,sd2,3",
			output: []int{1, 2, 3},
		},
		{
			name: "commas_only",
			input:  ",,",
			output: []int{},
		},
		{
			name: "spaces",
			input:  "1 2 3",
			output: []int{123},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := parseChannelsString(tt.input)

			if want, got := tt.output, result; !reflect.DeepEqual(want, got) {
				t.Fatalf("parseChannelsStrings(%v):\n- want: %v\n-  got: %v", tt.input, want, got)
			}
		})
	}
}

func TestChannelToFrequency(t *testing.T) {
	tests := []struct {
		channel  int
		frequency int
	}{
		{
			channel: 1,
			frequency: 2412,
		},
		{
			channel: 2,
			frequency: 2417,
		},
		{
			channel: 3,
			frequency: 2422,
		},
		{
			channel: 4,
			frequency: 2427,
		},
		{
			channel: 5,
			frequency: 2432,
		},
		{
			channel: 6,
			frequency: 2437,
		},
		{
			channel: 7,
			frequency: 2442,
		},
		{
			channel: 8,
			frequency: 2447,
		},
		{
			channel: 9,
			frequency: 2452,
		},
		{
			channel: 10,
			frequency: 2457,
		},
		{
			channel: 11,
			frequency: 2462,
		},
		{
			channel: 12,
			frequency: 2467,
		},
		{
			channel: 13,
			frequency: 2472,
		},
		{
			channel: 14,
			frequency: 2484,
		},
		{
			channel: 15,
			frequency: 0,
		},
		{
			channel: -1,
			frequency: 0,
		},
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.channel), func(t *testing.T) {
			if want, got := tt.frequency, channelToFrequency(tt.channel); want != got {
				t.Fatalf("parseChannelsStrings(%v):\n- want: %v\n-  got: %v", tt.channel, want, got)
			}
		})
	}
}
