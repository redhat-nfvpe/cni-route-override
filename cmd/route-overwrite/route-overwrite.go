// Copyright 2019 CNI authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This is a "meta-plugin". It reads in its own netconf, it does not create
// any network interface but just changes route information given from
// previous cni plugins

package main

import (
	"bytes"
	"fmt"
	//	"os" //XXX
	"encoding/json"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
)

// Todo:
// + only checko route/dst
//go build ./cmd/route-overwrite/

// RouteOverwrite represents the network route-overwrite configuration
type RouteOverwriteConfig struct {
	types.NetConf

	//RawPrevResult map[string]interface{} `json:"prevResult,omitempty"`
	PrevResult *current.Result `json:"-"`

	FlushRoutes bool           `json:"flushroutes,omitempty"`
	DelRoutes   []*types.Route `json:"delroutes"`
	AddRoutes   []*types.Route `json:"addroutes"`

	Args *struct {
		A *IPAMArgs `json:"cni"`
	} `json:"args"`
}

type IPAMArgs struct {
	FlushRoutes *bool          `json:"flushroutes,omitempty"`
	DelRoutes   []*types.Route `json:"delroutes,omitempty"`
	AddRoutes   []*types.Route `json:"addroutes,omitempty"`
}

/*
type RouteOverwriteArgs struct {
	types.CommonArgs
}
*/
func parseConf(data []byte, envArgs string) (*RouteOverwriteConfig, error) {
	conf := RouteOverwriteConfig{FlushRoutes: false}

	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("failed to load netconf: %v", err)
	}

	// overwrite values by args
	if conf.Args != nil {
		if conf.Args.A.FlushRoutes != nil {
			conf.FlushRoutes = *conf.Args.A.FlushRoutes
		}

		if conf.Args.A.DelRoutes != nil {
			conf.DelRoutes = conf.Args.A.DelRoutes
		}

		if conf.Args.A.AddRoutes != nil {
			conf.AddRoutes = conf.Args.A.AddRoutes
		}

	}

	// Parse previous result
	if conf.RawPrevResult != nil {
		resultBytes, err := json.Marshal(conf.RawPrevResult)
		if err != nil {
			return nil, fmt.Errorf("could not serialize prevResult: %v", err)
		}

		res, err := version.NewResult(conf.CNIVersion, resultBytes)

		if err != nil {
			return nil, fmt.Errorf("could not parse prevResult: %v", err)
		}

		conf.RawPrevResult = nil
		conf.PrevResult, err = current.NewResultFromResult(res)
		if err != nil {
			return nil, fmt.Errorf("could not convert result to current version: %v", err)
		}
	}
	//fmt.Fprintf(os.Stderr, "XXX1: %v\n", conf.RouteFlush)

	return &conf, nil
}

func processRoutes(conf *RouteOverwriteConfig) (*current.Result, error) {
	res, err := current.NewResultFromResult(conf.PrevResult)

	if err != nil {
		return nil, fmt.Errorf("could not convert result to current version: %v", err)
	}

	newRoutes := []*types.Route{}
	// Flush route if required
	if !conf.FlushRoutes {
		//fmt.Fprintf(os.Stderr, "XXX!\n")
	NEXT:
		for _, route := range res.Routes {
			for _, delroute := range conf.DelRoutes {
				if route.Dst.IP.Equal(delroute.Dst.IP) &&
					bytes.Equal(route.Dst.Mask, delroute.Dst.Mask) {
					continue NEXT
				}
			}
			newRoutes = append(newRoutes, route)
		}
	}

	for _, route := range conf.AddRoutes {
		newRoutes = append(newRoutes, route)
	}

	res.Routes = newRoutes

	return res, nil
}

func cmdAdd(args *skel.CmdArgs) error {
	overwriteConf, err := parseConf(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	newResult, err := processRoutes(overwriteConf)
	if err != nil {
		return fmt.Errorf("failed to overwrite routes: %v", err)
	}

	return types.PrintResult(newResult, overwriteConf.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	// TODO: the settings are not reverted to the previous values. Reverting the
	// settings is not useful when the whole container goes away but it could be
	// useful in scenarios where plugins are added and removed at runtime.
	return nil
}

func cmdGet(args *skel.CmdArgs) error {
	// TODO: implement
	return fmt.Errorf("not implemented")
}

func main() {
	// TODO: implement plugin version
	skel.PluginMain(cmdAdd, cmdGet, cmdDel, version.All, "TODO")
}
