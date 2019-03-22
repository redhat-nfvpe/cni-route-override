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
	//	"fmt" //XXX
	"testing"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types/current"
	//	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// ginkgo -p --randomizeAllSpecs --randomizeSuites --failOnPending --progress -r .

func TestRouteOverwrite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RouteOverwrite")
}

var _ = Describe("route-overwrite operations by conf", func() {
	const IFNAME string = "dummy0"

	It("passes prevResult through unchanged", func() {
		conf := []byte(`{
	"name": "test",
	"type": "route-overwrite",
	"cniVersion": "0.3.1",
	"prevResult": {
		"interfaces": [
			{"name": "dummy0", "sandbox":"netns"}
		],
		"ips": [
			{
				"version": "4",
				"address": "10.0.0.2/24",
				"gateway": "10.0.0.1",
				"interface": 0
			}
		],
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
			Netns:       "",
			IfName:      IFNAME,
			StdinData:   conf,
		}

		defer GinkgoRecover()

		r, _, err := testutils.CmdAddWithArgs(args, func() error {
			return cmdAdd(args)
		})
		Expect(err).NotTo(HaveOccurred())

		result, err := current.GetResult(r)
		Expect(err).NotTo(HaveOccurred())

		Expect(len(result.Interfaces)).To(Equal(1))
		Expect(result.Interfaces[0].Name).To(Equal(IFNAME))
		Expect(len(result.IPs)).To(Equal(1))
		Expect(result.IPs[0].Address.String()).To(Equal("10.0.0.2/24"))
		Expect(result.Routes[0].Dst.String()).To(Equal("0.0.0.0/0"))
		Expect(result.Routes[0].GW.String()).To(Equal("10.0.0.1"))
		Expect(result.Routes[1].Dst.String()).To(Equal("30.0.0.0/24"))
		Expect(result.Routes[1].GW).To(BeNil())
		Expect(result.Routes[2].Dst.String()).To(Equal("20.0.0.0/24"))
		Expect(result.Routes[2].GW.String()).To(Equal("10.0.0.254"))

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
			{"name": "dummy0", "sandbox":"netns"}
		],
		"ips": [
			{
				"version": "4",
				"address": "10.0.0.2/24",
				"gateway": "10.0.0.1",
				"interface": 0
			}
		],
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
			Netns:       "",
			IfName:      IFNAME,
			StdinData:   conf,
		}

		defer GinkgoRecover()

		r, _, err := testutils.CmdAddWithArgs(args, func() error {
			return cmdAdd(args)
		})
		Expect(err).NotTo(HaveOccurred())

		result, err := current.GetResult(r)
		Expect(err).NotTo(HaveOccurred())

		Expect(len(result.Interfaces)).To(Equal(1))
		Expect(result.Interfaces[0].Name).To(Equal(IFNAME))
		Expect(len(result.IPs)).To(Equal(1))
		Expect(result.IPs[0].Address.String()).To(Equal("10.0.0.2/24"))
		Expect(result.Routes).To(BeNil())

		Expect(err).NotTo(HaveOccurred())
	})

	It("check delroutes works", func() {
		conf := []byte(`{
	"name": "test",
	"type": "route-overwrite",
	"cniVersion": "0.3.1",
        "delroutes": [ { "dst": "0.0.0.0/0" }, 
                       { "dst": "20.0.0.0/24" } ],
	"prevResult": {
		"interfaces": [
			{"name": "dummy0", "sandbox":"netns"}
		],
		"ips": [
			{
				"version": "4",
				"address": "10.0.0.2/24",
				"gateway": "10.0.0.1",
				"interface": 0
			}
		],
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
			Netns:       "",
			IfName:      IFNAME,
			StdinData:   conf,
		}

		defer GinkgoRecover()

		r, _, err := testutils.CmdAddWithArgs(args, func() error {
			return cmdAdd(args)
		})
		Expect(err).NotTo(HaveOccurred())

		result, err := current.GetResult(r)
		Expect(err).NotTo(HaveOccurred())

		Expect(len(result.Interfaces)).To(Equal(1))
		Expect(result.Interfaces[0].Name).To(Equal(IFNAME))
		Expect(len(result.IPs)).To(Equal(1))
		Expect(result.IPs[0].Address.String()).To(Equal("10.0.0.2/24"))
		Expect(result.Routes[0].Dst.String()).To(Equal("30.0.0.0/24"))
		Expect(result.Routes[0].GW).To(BeNil())

		Expect(err).NotTo(HaveOccurred())
	})

	It("check addroutes works", func() {
		conf := []byte(`{
	"name": "test",
	"type": "route-overwrite",
	"cniVersion": "0.3.1",
        "delroutes": [ { "dst": "0.0.0.0/0" }, 
                       { "dst": "20.0.0.0/24" } ],
        "addroutes": [ { "dst": "0.0.0.0/0", "gw": "10.0.0.254" }, 
                       { "dst": "20.0.0.0/24" } ],
	"prevResult": {
		"interfaces": [
			{"name": "dummy0", "sandbox":"netns"}
		],
		"ips": [
			{
				"version": "4",
				"address": "10.0.0.2/24",
				"gateway": "10.0.0.1",
				"interface": 0
			}
		],
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
			Netns:       "",
			IfName:      IFNAME,
			StdinData:   conf,
		}

		defer GinkgoRecover()

		r, _, err := testutils.CmdAddWithArgs(args, func() error {
			return cmdAdd(args)
		})
		Expect(err).NotTo(HaveOccurred())

		result, err := current.GetResult(r)
		Expect(err).NotTo(HaveOccurred())

		Expect(len(result.Interfaces)).To(Equal(1))
		Expect(result.Interfaces[0].Name).To(Equal(IFNAME))
		Expect(len(result.IPs)).To(Equal(1))
		Expect(result.IPs[0].Address.String()).To(Equal("10.0.0.2/24"))
		Expect(result.Routes[0].Dst.String()).To(Equal("30.0.0.0/24"))
		Expect(result.Routes[0].GW).To(BeNil())
		Expect(result.Routes[1].Dst.String()).To(Equal("0.0.0.0/0"))
		Expect(result.Routes[1].GW.String()).To(Equal("10.0.0.254"))
		Expect(result.Routes[2].Dst.String()).To(Equal("20.0.0.0/24"))
		Expect(result.Routes[2].GW).To(BeNil())

		Expect(err).NotTo(HaveOccurred())
	})

})

var _ = Describe("route-overwrite operations by args", func() {
	const IFNAME string = "dummy0"

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
			{"name": "dummy0", "sandbox":"netns"}
		],
		"ips": [
			{
				"version": "4",
				"address": "10.0.0.2/24",
				"gateway": "10.0.0.1",
				"interface": 0
			}
		],
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
			Netns:       "",
			IfName:      IFNAME,
			StdinData:   conf,
		}

		defer GinkgoRecover()

		r, _, err := testutils.CmdAddWithArgs(args, func() error {
			return cmdAdd(args)
		})
		Expect(err).NotTo(HaveOccurred())

		result, err := current.GetResult(r)
		Expect(err).NotTo(HaveOccurred())

		Expect(len(result.Interfaces)).To(Equal(1))
		Expect(result.Interfaces[0].Name).To(Equal(IFNAME))
		Expect(len(result.IPs)).To(Equal(1))
		Expect(result.IPs[0].Address.String()).To(Equal("10.0.0.2/24"))
		Expect(result.Routes).To(BeNil())

		Expect(err).NotTo(HaveOccurred())
	})

	It("check delroutes works", func() {
		conf := []byte(`{
	"name": "test",
	"type": "route-overwrite",
	"cniVersion": "0.3.1",
        "args": {
          "cni": {
            "delroutes": [ { "dst": "0.0.0.0/0" }, 
                           { "dst": "20.0.0.0/24" } ]
          }
        },
	"prevResult": {
		"interfaces": [
			{"name": "dummy0", "sandbox":"netns"}
		],
		"ips": [
			{
				"version": "4",
				"address": "10.0.0.2/24",
				"gateway": "10.0.0.1",
				"interface": 0
			}
		],
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
			Netns:       "",
			IfName:      IFNAME,
			StdinData:   conf,
		}

		defer GinkgoRecover()

		r, _, err := testutils.CmdAddWithArgs(args, func() error {
			return cmdAdd(args)
		})
		Expect(err).NotTo(HaveOccurred())

		result, err := current.GetResult(r)
		Expect(err).NotTo(HaveOccurred())

		Expect(len(result.Interfaces)).To(Equal(1))
		Expect(result.Interfaces[0].Name).To(Equal(IFNAME))
		Expect(len(result.IPs)).To(Equal(1))
		Expect(result.IPs[0].Address.String()).To(Equal("10.0.0.2/24"))
		Expect(result.Routes[0].Dst.String()).To(Equal("30.0.0.0/24"))
		Expect(result.Routes[0].GW).To(BeNil())

		Expect(err).NotTo(HaveOccurred())
	})

	It("check addroutes works", func() {
		conf := []byte(`{
	"name": "test",
	"type": "route-overwrite",
	"cniVersion": "0.3.1",
        "args": {
          "cni": {
          "delroutes": [ { "dst": "0.0.0.0/0" }, 
                         { "dst": "20.0.0.0/24" } ],
          "addroutes": [ { "dst": "0.0.0.0/0", "gw": "10.0.0.254" }, 
                         { "dst": "20.0.0.0/24" } ]
          }
        },
	"prevResult": {
		"interfaces": [
			{"name": "dummy0", "sandbox":"netns"}
		],
		"ips": [
			{
				"version": "4",
				"address": "10.0.0.2/24",
				"gateway": "10.0.0.1",
				"interface": 0
			}
		],
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
			Netns:       "",
			IfName:      IFNAME,
			StdinData:   conf,
		}

		defer GinkgoRecover()

		r, _, err := testutils.CmdAddWithArgs(args, func() error {
			return cmdAdd(args)
		})
		Expect(err).NotTo(HaveOccurred())

		result, err := current.GetResult(r)
		Expect(err).NotTo(HaveOccurred())

		Expect(len(result.Interfaces)).To(Equal(1))
		Expect(result.Interfaces[0].Name).To(Equal(IFNAME))
		Expect(len(result.IPs)).To(Equal(1))
		Expect(result.IPs[0].Address.String()).To(Equal("10.0.0.2/24"))
		Expect(result.Routes[0].Dst.String()).To(Equal("30.0.0.0/24"))
		Expect(result.Routes[0].GW).To(BeNil())
		Expect(result.Routes[1].Dst.String()).To(Equal("0.0.0.0/0"))
		Expect(result.Routes[1].GW.String()).To(Equal("10.0.0.254"))
		Expect(result.Routes[2].Dst.String()).To(Equal("20.0.0.0/24"))
		Expect(result.Routes[2].GW).To(BeNil())

		Expect(err).NotTo(HaveOccurred())
	})

})
