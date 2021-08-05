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
	"errors"
	"fmt"
	"github.com/mdlayher/genetlink"
	"github.com/mdlayher/netlink"
	"github.com/mdlayher/netlink/nlenc"
	"github.com/mdlayher/wifi"
	flag "github.com/spf13/pflag"
	"github.com/xlab/nl80211/nl80211"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	running = true
	showHelp bool
	showVersion bool
	interfaceName string
	channelsString string
	delay int
	timeout int
)

const (
	ProgramName = "chopper"
	Version = "1.0.0"
)

func checkMonitorInterface(iface string) (*wifi.Interface, error) {
	client, err := wifi.New()
	if err != nil {
		log.Panicln(err)
	}
	defer client.Close()

	interfaces, err := client.Interfaces()
	if err != nil {
		return nil, err
	}

	// Find interface
	var ifaceFound *wifi.Interface
	for _, wiface := range interfaces {
		if wiface.Name == iface {
			ifaceFound = wiface
			break
		}
	}

	// Check monitor mode
	if ifaceFound == nil {
		return nil, errors.New(fmt.Sprintf("cannot find %v", iface))
	} else if ifaceFound.Type != wifi.InterfaceTypeMonitor {
		return nil, errors.New(fmt.Sprintf("%v is not in monitor mode", iface))
	}

	return ifaceFound, nil
}

func channelToFrequency(channel int) int {
	// TODO: Add support for 5GHz
	if channel <= 0 {
		return 0
	}

	if channel == 14 {
		return 2484
	} else if channel < 14 {
		return 2407 + channel * 5
	}

	return 0
}

func parseChannelsString(input string) ([]int, error) {
	ret := make([]int, 0)

	// Remove all non-digits
	reg, err := regexp.Compile("[^0-9,]+")
	if err != nil {
		return nil, err
	}
	processedString := reg.ReplaceAllString(input, "")

	// Split on comma
	for _, part := range strings.Split(processedString, ",") {
		if part == "" {
			continue
		}

		value, err := strconv.ParseInt(part, 10, 32)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "WARNING: there was an error parsing: %v\n", err)
			_, _ = fmt.Fprintf(os.Stderr, "WARNING: please report it.\n")
			continue
		}

		if value == 0 {
			continue
		}

		ret = append(ret, int(value))
	}

	return ret, nil
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}


func main() {
	// Command arguments
	flag.BoolVarP(&showHelp, "help", "h", false, "show this help message")
	flag.BoolVarP(&showVersion, "version", "V", false, "show version")
	flag.StringVarP(&interfaceName, "interface", "i", "", "interface name (must be in monitor mode)")
	flag.StringVarP(&channelsString, "channels", "c", "", "comma-separated list of channels (default: 1,8,2,9,3,10,4,11,5,12,6,13,7)")
	flag.IntVarP(&delay, "delay", "d", 200, "delay between each hop")
	flag.IntVarP(&timeout, "timeout", "t", 0, "exit the program after X seconds")
	flag.Parse()

	if showHelp {
		flag.Usage()
		os.Exit(0)
	} else if showVersion {
		fmt.Printf("%s v%s\n", ProgramName, Version)
		os.Exit(0)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		running = false
	}()

	// Check arguments
	if interfaceName == "" {
		flag.Usage()
		os.Exit(1)
	}
	if isFlagPassed("delay") && delay < 10 {
		_, _ = fmt.Fprintf(os.Stderr, "WARNING: the delay is very small, why are you doing this?\n")
	}
	if isFlagPassed("timeout") {
		if timeout <= 0 {
			_, _ = fmt.Fprintf(os.Stderr, "WARNING: timeout cannot be 0, running until SIGINT.\n")
		} else {
			time.AfterFunc(time.Duration(timeout)*time.Second, func() {
				quit <- os.Interrupt
			})
		}
	}
	channels, _ := parseChannelsString(channelsString)
	if len(channels) <= 0 {
		channels = []int{1, 8, 2, 9, 3, 10, 4, 11, 5, 12, 6, 13, 7}
	}

	// Check interface
	iface, err := checkMonitorInterface(interfaceName)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	// Connect to generic Netlink socket
	nlSocket, err := genetlink.Dial(nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: cannot connect to Netlink socket: %v\n", err)
		os.Exit(1)
	}
	defer nlSocket.Close()

	// Resolve nl80211
	nl80211Family, err := nlSocket.GetFamily("nl80211")
	if err != nil {
		// TODO: Print families for debugging purposes
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: nl80211 not available\n")
		os.Exit(1)
	}

	idx := 0
	for running {
		// Prepare attributes
		data, err := netlink.MarshalAttributes(
			[]netlink.Attribute{
				{
					Type: nl80211.AttrIfindex,
					Data: nlenc.Uint32Bytes(uint32(iface.Index)),
				},
				{
					Type: nl80211.AttrWiphyFreq,
					Data: nlenc.Uint32Bytes(uint32(channelToFrequency(channels[idx]))),
				},

				// TODO: Add support for HT20, HT40+, HT40-
				{
					Type: nl80211.AttrChannelWidth,
					Data: nlenc.Uint32Bytes(uint32(nl80211.ChanWidth20Noht)),
				},
				{
					Type: nl80211.AttrWiphyChannelType,
					Data: nlenc.Uint32Bytes(uint32(nl80211.ChanHt20)),
				},
			})
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}

		// Prepare message
		nlMessage := genetlink.Message{
			Header: genetlink.Header{
				Command: nl80211.CommandSetChannel,
				Version: nl80211Family.Version,
			},
			Data: data,
		}

		_, err = nlSocket.Execute(nlMessage, nl80211Family.ID, netlink.Request | netlink.Acknowledge)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Cannot set channel %v\n", channels[idx])
			_, _ = fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}

		// Increase counter
		idx++
		if idx >= len(channels) {
			idx = 0
		}

		// Delay
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}