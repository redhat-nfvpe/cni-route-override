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

package main

import (
	//"fmt"
	"net"
	//"os"
	"testing"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/testutils"

	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// ginkgo -p --randomizeAllSpecs --randomizeSuites --failOnPending --progress -r ./cmd/...

func TestRouteOverwrite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RouteOverwrite")
}

// helper function

func testAddRoute(link netlink.Link, ip net.IP, mask net.IPMask, gw net.IP) error {
	dst := &net.IPNet{
		IP:   ip,
		Mask: mask,
	}
	route := netlink.Route{LinkIndex: link.Attrs().Index, Dst: dst, Gw: gw}
	err := netlink.RouteAdd(&route)
	return err
}

func testAddAddr(link netlink.Link, ip net.IP, mask net.IPMask) error {
	var address = &net.IPNet{IP: ip, Mask: mask}
	var addr = &netlink.Addr{IPNet: address}
	err := netlink.AddrAdd(link, addr)
	return err
}

func testHasRoute(routes []netlink.Route, dst *net.IPNet) bool {
	for _, route := range routes {
		// default route case
		if dst == nil {
			if route.Dst == nil {
				return true
			}
		} else if route.Dst != nil && dst != nil &&
			route.Dst.IP.Equal(dst.IP) && route.Dst.Mask.String() == dst.Mask.String() {
			return true
		}
	}

	return false
}

var _ = Describe("route-overwrite operations by conf", func() {
	const IFNAME string = "dummy0"
	var originalNS ns.NetNS
	var targetNS ns.NetNS

	BeforeEach(func() {
		// Create a new NetNS so we don't modify the host
		var err error
		originalNS, err = testutils.NewNS()
		Expect(err).NotTo(HaveOccurred())

		targetNS, err = testutils.NewNS()
		Expect(err).NotTo(HaveOccurred())

		err = targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()

			err = netlink.LinkAdd(&netlink.Dummy{
				LinkAttrs: netlink.LinkAttrs{
					Name: IFNAME,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			_, err = netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())
			return nil
		})
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		Expect(originalNS.Close()).To(Succeed())
	})

	It("passes prevResult through unchanged", func() {
		conf := []byte(`{
			"name": "test",
			"type": "route-overwrite",
			"cniVersion": "0.3.1",
			"prevResult": {
				"interfaces": [
				{
					"name": "dummy0",
					"sandbox":"netns"
				}],
				"ips": [
				{
					"version": "4",
					"address": "10.0.0.2/24",
					"gateway": "10.0.0.1",
					"interface": 0
				}],
				"routes": [
				{
					"dst": "30.0.0.0/24"
				},
				{
					"dst": "20.0.0.0/24",
					"gw": "10.0.0.254"
				}]
			}
		}`)

		args := &skel.CmdArgs{
			ContainerID: "dummy",
			Netns:       targetNS.Path(),
			IfName:      IFNAME,
			StdinData:   conf,
		}

		// set address/route as fakeCNI plugin
		err := targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())
			err = netlink.LinkSetUp(link)
			Expect(err).NotTo(HaveOccurred())

			// addr 10.0.0.2/24
			err = testAddAddr(link, net.IPv4(10, 0, 0, 2), net.CIDRMask(24, 32))
			Expect(err).NotTo(HaveOccurred())

			// add default gateway into IFNAME
			err = testAddRoute(link,
				net.IPv4(0, 0, 0, 0), net.CIDRMask(0, 0),
				net.IPv4(10, 0, 0, 1))
			Expect(err).NotTo(HaveOccurred())

			//"dst": "30.0.0.0/24"
			err = testAddRoute(link,
				net.IPv4(30, 0, 0, 0), net.CIDRMask(24, 32),
				net.IPv4(10, 0, 0, 1))
			Expect(err).NotTo(HaveOccurred())

			// "dst": "20.0.0.0/24", "gw": "10.0.0.254"
			err = testAddRoute(link,
				net.IPv4(20, 0, 0, 0), net.CIDRMask(24, 32),
				net.IPv4(10, 0, 0, 254))
			Expect(err).NotTo(HaveOccurred())

			return nil
		})
		Expect(err).NotTo(HaveOccurred())

		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			var result *current.Result

			r, _, err := testutils.CmdAddWithArgs(args, func() error {
				return cmdAdd(args)
			})
			Expect(err).NotTo(HaveOccurred())

			result, err = current.GetResult(r)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(result.Interfaces)).To(Equal(1))
			Expect(result.Interfaces[0].Name).To(Equal(IFNAME))
			Expect(len(result.IPs)).To(Equal(1))
			Expect(result.IPs[0].Address.String()).To(Equal("10.0.0.2/24"))
			Expect(result.Routes[0].Dst.String()).To(Equal("30.0.0.0/24"))
			Expect(result.Routes[0].GW).To(BeNil())
			Expect(result.Routes[1].Dst.String()).To(Equal("20.0.0.0/24"))
			Expect(result.Routes[1].GW.String()).To(Equal("10.0.0.254"))

			return nil
		})

		// check route info
		err = targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()

			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())

			// FAMILY_ALL for all, but use v4 for its simplicity
			routes, _ := netlink.RouteList(link, netlink.FAMILY_V4)
			//fmt.Fprintf(os.Stderr, "routes: %v\n", routes) // XXX
			Expect(len(routes)).To(Equal(4)) // default + add2 + interface route
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("check flushroutes clears all routes", func() {
		conf := []byte(`{
			"name": "test",
			"type": "route-overwrite",
			"cniVersion": "0.3.1",
			"flushroutes": true,
			"prevResult": {
				"interfaces": [
				{
					"name": "dummy0",
					"sandbox":"netns"
				}],
				"ips": [
				{
					"version": "4",
					"address": "10.0.0.2/24",
					"gateway": "10.0.0.1",
					"interface": 0
				}],
				"routes": [
				{
					"dst": "0.0.0.0/0",
					"gw": "10.0.0.1"
				},
				{
					"dst": "30.0.0.0/24"
				},
				{
					"dst": "20.0.0.0/24",
					"gw": "10.0.0.254"
				}]
			}
		}`)

		args := &skel.CmdArgs{
			ContainerID: "dummy",
			Netns:       targetNS.Path(),
			IfName:      IFNAME,
			StdinData:   conf,
		}

		// set address/route as fakeCNI plugin
		err := targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())
			err = netlink.LinkSetUp(link)
			Expect(err).NotTo(HaveOccurred())

			// addr 10.0.0.2/24
			err = testAddAddr(link, net.IPv4(10, 0, 0, 2), net.CIDRMask(24, 32))
			Expect(err).NotTo(HaveOccurred())

			// add default gateway into IFNAME
			err = testAddRoute(link,
				net.IPv4(0, 0, 0, 0), net.CIDRMask(0, 0),
				net.IPv4(10, 0, 0, 1))
			Expect(err).NotTo(HaveOccurred())

			//"dst": "30.0.0.0/24"
			err = testAddRoute(link,
				net.IPv4(30, 0, 0, 0), net.CIDRMask(24, 32),
				net.IPv4(10, 0, 0, 1))
			Expect(err).NotTo(HaveOccurred())

			// "dst": "20.0.0.0/24", "gw": "10.0.0.254"
			err = testAddRoute(link,
				net.IPv4(20, 0, 0, 0), net.CIDRMask(24, 32),
				net.IPv4(10, 0, 0, 254))
			Expect(err).NotTo(HaveOccurred())

			return nil
		})
		Expect(err).NotTo(HaveOccurred())

		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			var result *current.Result

			r, _, err := testutils.CmdAddWithArgs(args, func() error {
				return cmdAdd(args)
			})
			Expect(err).NotTo(HaveOccurred())

			result, err = current.GetResult(r)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(result.Interfaces)).To(Equal(1))
			Expect(result.Interfaces[0].Name).To(Equal(IFNAME))
			Expect(len(result.IPs)).To(Equal(1))
			Expect(result.IPs[0].Address.String()).To(Equal("10.0.0.2/24"))
			Expect(result.Routes).To(BeNil())

			Expect(err).NotTo(HaveOccurred())

			return nil
		})

		// check route info
		err = targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()

			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())

			// FAMILY_ALL for all, but use v4 for its simplicity
			routes, _ := netlink.RouteList(link, netlink.FAMILY_V4)
			Expect(len(routes)).To(Equal(1)) // interface route
			return nil
		})
		Expect(err).NotTo(HaveOccurred())

	})

	It("check flushgateway clears gw routes", func() {
		conf := []byte(`{
			"name": "test",
			"type": "route-overwrite",
			"cniVersion": "0.3.1",
			"flushgateway": true,
			"prevResult": {
				"interfaces": [
				{
					"name": "dummy0",
					"sandbox":"netns"
				}],
				"ips": [
				{
					"version": "4",
					"address": "10.0.0.2/24",
					"gateway": "10.0.0.1",
					"interface": 0
				}],
				"routes": [
				{
					"dst": "0.0.0.0/0",
					"gw": "10.0.0.1"
				},
				{
					"dst": "30.0.0.0/24"
				},
				{
					"dst": "20.0.0.0/24",
					"gw": "10.0.0.254"
				}
			]
		}
	}`)

		args := &skel.CmdArgs{
			ContainerID: "dummy",
			Netns:       targetNS.Path(),
			IfName:      IFNAME,
			StdinData:   conf,
		}

		// set address/route as fakeCNI plugin
		err := targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())
			err = netlink.LinkSetUp(link)
			Expect(err).NotTo(HaveOccurred())

			// addr 10.0.0.2/24
			err = testAddAddr(link, net.IPv4(10, 0, 0, 2), net.CIDRMask(24, 32))
			Expect(err).NotTo(HaveOccurred())

			// add default gateway into IFNAME
			err = testAddRoute(link,
				net.IPv4(0, 0, 0, 0), net.CIDRMask(0, 0),
				net.IPv4(10, 0, 0, 1))
			Expect(err).NotTo(HaveOccurred())

			//"dst": "30.0.0.0/24"
			err = testAddRoute(link,
				net.IPv4(30, 0, 0, 0), net.CIDRMask(24, 32),
				net.IPv4(10, 0, 0, 1))
			Expect(err).NotTo(HaveOccurred())

			// "dst": "20.0.0.0/24", "gw": "10.0.0.254"
			err = testAddRoute(link,
				net.IPv4(20, 0, 0, 0), net.CIDRMask(24, 32),
				net.IPv4(10, 0, 0, 254))
			Expect(err).NotTo(HaveOccurred())

			return nil
		})
		Expect(err).NotTo(HaveOccurred())

		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			var result *current.Result

			r, _, err := testutils.CmdAddWithArgs(args, func() error {
				return cmdAdd(args)
			})
			Expect(err).NotTo(HaveOccurred())

			result, err = current.GetResult(r)
			Expect(err).NotTo(HaveOccurred())

			// check no default gw
			Expect(len(result.IPs)).To(Equal(1))
			Expect(result.IPs[0].Gateway.String()).To(Equal("0.0.0.0"))
			Expect(result.Routes[0].Dst.String()).NotTo(Equal("0.0.0.0/0"))

			return nil
		})

		// check route info
		err = targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()

			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())

			// FAMILY_ALL for all, but use v4 for its simplicity
			routes, _ := netlink.RouteList(link, netlink.FAMILY_V4)
			//fmt.Fprintf(os.Stderr, "gw routes: %v\n", routes) // XXX
			Expect(len(routes)).To(Equal(3)) // add2 + interface route
			Expect(testHasRoute(routes, nil)).To(Equal(false))
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("check delroutes works", func() {
		conf := []byte(`{
			"name": "test",
			"type": "route-overwrite",
			"cniVersion": "0.3.1",
			"delroutes": [ { "dst": "20.0.0.0/24" } ],
			"prevResult": {
				"interfaces": [
				{
					"name": "dummy0",
					"sandbox":"netns"
				}],
				"ips": [
				{
					"version": "4",
					"address": "10.0.0.2/24",
					"gateway": "10.0.0.1",
					"interface": 0
				}],
				"routes": [
				{
					"dst": "0.0.0.0/0",
					"gw": "10.0.0.1"
				},
				{
					"dst": "30.0.0.0/24"
				},
				{
					"dst": "20.0.0.0/24",
					"gw": "10.0.0.254"
				}]
			}
		}`)

		args := &skel.CmdArgs{
			ContainerID: "dummy",
			Netns:       targetNS.Path(),
			IfName:      IFNAME,
			StdinData:   conf,
		}

		// set address/route as fakeCNI plugin
		err := targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())
			err = netlink.LinkSetUp(link)
			Expect(err).NotTo(HaveOccurred())

			// addr 10.0.0.2/24
			err = testAddAddr(link, net.IPv4(10, 0, 0, 2), net.CIDRMask(24, 32))
			Expect(err).NotTo(HaveOccurred())

			// add default gateway into IFNAME
			err = testAddRoute(link,
				net.IPv4(0, 0, 0, 0), net.CIDRMask(0, 0),
				net.IPv4(10, 0, 0, 1))
			Expect(err).NotTo(HaveOccurred())

			//"dst": "30.0.0.0/24"
			err = testAddRoute(link,
				net.IPv4(30, 0, 0, 0), net.CIDRMask(24, 32),
				net.IPv4(10, 0, 0, 1))
			Expect(err).NotTo(HaveOccurred())

			// "dst": "20.0.0.0/24", "gw": "10.0.0.254"
			err = testAddRoute(link,
				net.IPv4(20, 0, 0, 0), net.CIDRMask(24, 32),
				net.IPv4(10, 0, 0, 254))
			Expect(err).NotTo(HaveOccurred())

			return nil
		})
		Expect(err).NotTo(HaveOccurred())

		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			var result *current.Result

			r, _, err := testutils.CmdAddWithArgs(args, func() error {
				return cmdAdd(args)
			})
			Expect(err).NotTo(HaveOccurred())

			result, err = current.GetResult(r)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(result.Interfaces)).To(Equal(1))
			Expect(result.Interfaces[0].Name).To(Equal(IFNAME))
			Expect(len(result.IPs)).To(Equal(1))
			Expect(result.IPs[0].Address.String()).To(Equal("10.0.0.2/24"))
			Expect(result.Routes[0].Dst.String()).To(Equal("0.0.0.0/0"))
			Expect(result.Routes[0].GW).NotTo(BeNil())

			return nil
		})

		// check route info
		err = targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()

			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())

			// FAMILY_ALL for all, but use v4 for its simplicity
			routes, _ := netlink.RouteList(link, netlink.FAMILY_V4)
			//fmt.Fprintf(os.Stderr, "routes: %v\n", routes) // XXX
			Expect(len(routes)).To(Equal(3))
			_, delroute1, _ := net.ParseCIDR("20.0.0.0/24")
			Expect(testHasRoute(routes, delroute1)).To(Equal(false))

			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("check addroutes works", func() {
		conf := []byte(`{
			"name": "test",
			"type": "route-overwrite",
			"cniVersion": "0.3.1",
			"addroutes": [
			{
				"dst": "20.0.0.0/24"
			}],
			"prevResult": {
				"interfaces": [
				{
					"name": "dummy0", "sandbox":"netns"
				}],
				"ips": [
				{
					"version": "4",
					"address": "10.0.0.2/24",
					"gateway": "10.0.0.254",
					"interface": 0
				}],
				"routes": [
				{
					"dst": "0.0.0.0/0",
					"gw": "10.0.0.254"
				},
				{
					"dst": "30.0.0.0/24"
				}]
			}
		}`)

		args := &skel.CmdArgs{
			ContainerID: "dummy",
			Netns:       targetNS.Path(),
			IfName:      IFNAME,
			StdinData:   conf,
		}

		// set address/route as fakeCNI plugin
		err := targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())
			err = netlink.LinkSetUp(link)
			Expect(err).NotTo(HaveOccurred())

			// addr 10.0.0.2/24
			err = testAddAddr(link, net.IPv4(10, 0, 0, 2), net.CIDRMask(24, 32))
			Expect(err).NotTo(HaveOccurred())

			// add default gateway into IFNAME
			err = testAddRoute(link,
				net.IPv4(0, 0, 0, 0), net.CIDRMask(0, 0),
				net.IPv4(10, 0, 0, 1))
			Expect(err).NotTo(HaveOccurred())

			//"dst": "30.0.0.0/24"
			err = testAddRoute(link,
				net.IPv4(30, 0, 0, 0), net.CIDRMask(24, 32),
				net.IPv4(10, 0, 0, 1))
			Expect(err).NotTo(HaveOccurred())

			return nil
		})
		Expect(err).NotTo(HaveOccurred())
		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			var result *current.Result

			r, _, err := testutils.CmdAddWithArgs(args, func() error {
				return cmdAdd(args)
			})
			Expect(err).NotTo(HaveOccurred())

			result, err = current.GetResult(r)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(result.Interfaces)).To(Equal(1))
			Expect(result.Interfaces[0].Name).To(Equal(IFNAME))
			Expect(len(result.IPs)).To(Equal(1))
			Expect(result.IPs[0].Address.String()).To(Equal("10.0.0.2/24"))
			Expect(result.Routes[0].Dst.String()).To(Equal("0.0.0.0/0"))
			Expect(result.Routes[0].GW.String()).To(Equal("10.0.0.254"))
			Expect(result.Routes[1].Dst.String()).To(Equal("30.0.0.0/24"))
			Expect(result.Routes[1].GW).To(BeNil())
			Expect(result.Routes[2].Dst.String()).To(Equal("20.0.0.0/24"))
			Expect(result.Routes[2].GW).To(BeNil())

			return nil
		})

		err = targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()

			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())

			// FAMILY_ALL for all, but use v4 for its simplicity
			routes, _ := netlink.RouteList(link, netlink.FAMILY_V4)
			//fmt.Fprintf(os.Stderr, "XXX routes: %v\n", routes) // XXX
			Expect(len(routes)).To(Equal(4))
			_, delroute1, _ := net.ParseCIDR("20.0.0.0/24")
			Expect(testHasRoute(routes, delroute1)).To(Equal(true))
			_, delroute2, _ := net.ParseCIDR("10.0.0.0/24")
			Expect(testHasRoute(routes, delroute2)).To(Equal(true))
			_, delroute3, _ := net.ParseCIDR("30.0.0.0/24")
			Expect(testHasRoute(routes, delroute3)).To(Equal(true))
			Expect(testHasRoute(routes, nil)).To(Equal(true))

			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})

})

var _ = Describe("route-overwrite operations by args", func() {
	const IFNAME string = "dummy0"

	var originalNS ns.NetNS
	var targetNS ns.NetNS

	BeforeEach(func() {
		// Create a new NetNS so we don't modify the host
		var err error
		originalNS, err = testutils.NewNS()
		Expect(err).NotTo(HaveOccurred())

		targetNS, err = testutils.NewNS()
		Expect(err).NotTo(HaveOccurred())

		err = targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()

			err = netlink.LinkAdd(&netlink.Dummy{
				LinkAttrs: netlink.LinkAttrs{
					Name: IFNAME,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			_, err = netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())
			return nil
		})
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		Expect(originalNS.Close()).To(Succeed())
	})

	It("check flushroutes clears all routes", func() {
		conf := []byte(`{
			"name": "test",
			"type": "route-overwrite",
			"cniVersion": "0.3.1",
			"args": {
				"cni": {
					"flushroutes": true
				}
			},
			"prevResult": {
				"interfaces": [
				{
					"name": "dummy0",
					"sandbox":"netns"
				}],
				"ips": [
				{
					"version": "4",
					"address": "10.0.0.2/24",
					"gateway": "10.0.0.1",
					"interface": 0
				}],
				"routes": [
				{
					"dst": "0.0.0.0/0",
					"gw": "10.0.0.1"
				},
				{
					"dst": "30.0.0.0/24"
				},
				{
					"dst": "20.0.0.0/24",
					"gw": "10.0.0.254"
				}]
			}
		}`)

		args := &skel.CmdArgs{
			ContainerID: "dummy",
			Netns:       targetNS.Path(),
			IfName:      IFNAME,
			StdinData:   conf,
		}

		// set address/route as fakeCNI plugin
		err := targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())
			err = netlink.LinkSetUp(link)
			Expect(err).NotTo(HaveOccurred())

			// addr 10.0.0.2/24
			err = testAddAddr(link, net.IPv4(10, 0, 0, 2), net.CIDRMask(24, 32))
			Expect(err).NotTo(HaveOccurred())

			// add default gateway into IFNAME
			err = testAddRoute(link,
				net.IPv4(0, 0, 0, 0), net.CIDRMask(0, 0),
				net.IPv4(10, 0, 0, 1))
			Expect(err).NotTo(HaveOccurred())

			//"dst": "30.0.0.0/24"
			err = testAddRoute(link,
				net.IPv4(30, 0, 0, 0), net.CIDRMask(24, 32),
				net.IPv4(10, 0, 0, 1))
			Expect(err).NotTo(HaveOccurred())

			// "dst": "20.0.0.0/24", "gw": "10.0.0.254"
			err = testAddRoute(link,
				net.IPv4(20, 0, 0, 0), net.CIDRMask(24, 32),
				net.IPv4(10, 0, 0, 254))
			Expect(err).NotTo(HaveOccurred())

			return nil
		})
		Expect(err).NotTo(HaveOccurred())

		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			var result *current.Result

			r, _, err := testutils.CmdAddWithArgs(args, func() error {
				return cmdAdd(args)
			})
			Expect(err).NotTo(HaveOccurred())

			result, err = current.GetResult(r)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(result.Interfaces)).To(Equal(1))
			Expect(result.Interfaces[0].Name).To(Equal(IFNAME))
			Expect(len(result.IPs)).To(Equal(1))
			Expect(result.IPs[0].Address.String()).To(Equal("10.0.0.2/24"))
			Expect(result.Routes).To(BeNil())

			Expect(err).NotTo(HaveOccurred())

			return nil
		})

		// check route info
		err = targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()

			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())

			// FAMILY_ALL for all, but use v4 for its simplicity
			routes, _ := netlink.RouteList(link, netlink.FAMILY_V4)
			Expect(len(routes)).To(Equal(1)) // interface route
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("check delroutes works", func() {
		conf := []byte(`{
			"name": "test",
			"type": "route-overwrite",
			"cniVersion": "0.3.1",
			"args": {
				"cni": {
					"delroutes": [
					{
						"dst": "20.0.0.0/24"
					}]
				}
			},
			"prevResult": {
				"interfaces": [
				{
					"name": "dummy0",
					"sandbox":"netns"
				}],
				"ips": [
				{
					"version": "4",
					"address": "10.0.0.2/24",
					"gateway": "10.0.0.1",
					"interface": 0
				}],
				"routes": [
				{
					"dst": "0.0.0.0/0",
					"gw": "10.0.0.1"
				},
				{
					"dst": "30.0.0.0/24"
				},
				{
					"dst": "20.0.0.0/24",
					"gw": "10.0.0.254"
				}]
			}
		}`)

		args := &skel.CmdArgs{
			ContainerID: "dummy",
			Netns:       targetNS.Path(),
			IfName:      IFNAME,
			StdinData:   conf,
		}

		// set address/route as fakeCNI plugin
		err := targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())
			err = netlink.LinkSetUp(link)
			Expect(err).NotTo(HaveOccurred())

			// addr 10.0.0.2/24
			err = testAddAddr(link, net.IPv4(10, 0, 0, 2), net.CIDRMask(24, 32))
			Expect(err).NotTo(HaveOccurred())

			// add default gateway into IFNAME
			err = testAddRoute(link,
				net.IPv4(0, 0, 0, 0), net.CIDRMask(0, 0),
				net.IPv4(10, 0, 0, 1))
			Expect(err).NotTo(HaveOccurred())

			//"dst": "30.0.0.0/24"
			err = testAddRoute(link,
				net.IPv4(30, 0, 0, 0), net.CIDRMask(24, 32),
				net.IPv4(10, 0, 0, 1))
			Expect(err).NotTo(HaveOccurred())

			// "dst": "20.0.0.0/24", "gw": "10.0.0.254"
			err = testAddRoute(link,
				net.IPv4(20, 0, 0, 0), net.CIDRMask(24, 32),
				net.IPv4(10, 0, 0, 254))
			Expect(err).NotTo(HaveOccurred())

			return nil
		})
		Expect(err).NotTo(HaveOccurred())

		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			var result *current.Result

			r, _, err := testutils.CmdAddWithArgs(args, func() error {
				return cmdAdd(args)
			})
			Expect(err).NotTo(HaveOccurred())

			result, err = current.GetResult(r)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(result.Interfaces)).To(Equal(1))
			Expect(result.Interfaces[0].Name).To(Equal(IFNAME))
			Expect(len(result.IPs)).To(Equal(1))
			Expect(result.IPs[0].Address.String()).To(Equal("10.0.0.2/24"))
			Expect(result.Routes[0].Dst.String()).To(Equal("0.0.0.0/0"))
			Expect(result.Routes[0].GW).NotTo(BeNil())

			return nil
		})

		// check route info
		err = targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()

			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())

			// FAMILY_ALL for all, but use v4 for its simplicity
			routes, _ := netlink.RouteList(link, netlink.FAMILY_V4)
			//fmt.Fprintf(os.Stderr, "routes: %v\n", routes) // XXX
			Expect(len(routes)).To(Equal(3))
			_, delroute1, _ := net.ParseCIDR("20.0.0.0/24")
			Expect(testHasRoute(routes, delroute1)).To(Equal(false))

			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("check addroutes works", func() {
		conf := []byte(`{
			"name": "test",
			"type": "route-overwrite",
			"cniVersion": "0.3.1",
			"args": {
				"cni": {
					"addroutes": [
					{
						"dst": "20.0.0.0/24"
					}]
				}
			},
			"prevResult": {
				"interfaces": [
				{
					"name": "dummy0",
					"sandbox":"netns"
				}],
				"ips": [
				{
					"version": "4",
					"address": "10.0.0.2/24",
					"gateway": "10.0.0.254",
					"interface": 0
				}],
				"routes": [
				{
					"dst": "0.0.0.0/0",
					"gw": "10.0.0.254"
				},
				{
					"dst": "30.0.0.0/24"
				}]
			}
		}`)
		args := &skel.CmdArgs{
			ContainerID: "dummy",
			Netns:       targetNS.Path(),
			IfName:      IFNAME,
			StdinData:   conf,
		}

		// set address/route as fakeCNI plugin
		err := targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())
			err = netlink.LinkSetUp(link)
			Expect(err).NotTo(HaveOccurred())

			// addr 10.0.0.2/24
			err = testAddAddr(link, net.IPv4(10, 0, 0, 2), net.CIDRMask(24, 32))
			Expect(err).NotTo(HaveOccurred())

			// add default gateway into IFNAME
			err = testAddRoute(link,
				net.IPv4(0, 0, 0, 0), net.CIDRMask(0, 0),
				net.IPv4(10, 0, 0, 1))
			Expect(err).NotTo(HaveOccurred())

			//"dst": "30.0.0.0/24"
			err = testAddRoute(link,
				net.IPv4(30, 0, 0, 0), net.CIDRMask(24, 32),
				net.IPv4(10, 0, 0, 1))
			Expect(err).NotTo(HaveOccurred())

			// "dst": "20.0.0.0/24", "gw": "10.0.0.254"
			err = testAddRoute(link,
				net.IPv4(20, 0, 0, 0), net.CIDRMask(24, 32),
				net.IPv4(10, 0, 0, 254))
			Expect(err).NotTo(HaveOccurred())

			return nil
		})
		Expect(err).NotTo(HaveOccurred())
		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			var result *current.Result

			r, _, err := testutils.CmdAddWithArgs(args, func() error {
				return cmdAdd(args)
			})
			Expect(err).NotTo(HaveOccurred())

			result, err = current.GetResult(r)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(result.Interfaces)).To(Equal(1))
			Expect(result.Interfaces[0].Name).To(Equal(IFNAME))
			Expect(len(result.IPs)).To(Equal(1))
			Expect(result.IPs[0].Address.String()).To(Equal("10.0.0.2/24"))
			Expect(result.Routes[0].Dst.String()).To(Equal("0.0.0.0/0"))
			Expect(result.Routes[0].GW.String()).To(Equal("10.0.0.254"))
			Expect(result.Routes[1].Dst.String()).To(Equal("30.0.0.0/24"))
			Expect(result.Routes[1].GW).To(BeNil())
			Expect(result.Routes[2].Dst.String()).To(Equal("20.0.0.0/24"))
			Expect(result.Routes[2].GW).To(BeNil())

			return nil
		})

		err = targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()

			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())

			// FAMILY_ALL for all, but use v4 for its simplicity
			routes, _ := netlink.RouteList(link, netlink.FAMILY_V4)
			Expect(len(routes)).To(Equal(4))
			_, delroute1, _ := net.ParseCIDR("20.0.0.0/24")
			Expect(testHasRoute(routes, delroute1)).To(Equal(true))
			_, delroute2, _ := net.ParseCIDR("10.0.0.0/24")
			Expect(testHasRoute(routes, delroute2)).To(Equal(true))
			_, delroute3, _ := net.ParseCIDR("30.0.0.0/24")
			Expect(testHasRoute(routes, delroute3)).To(Equal(true))
			Expect(testHasRoute(routes, nil)).To(Equal(true))

			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})

})
