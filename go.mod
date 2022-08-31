module github.com/redhat-nfvpe/cni-route-override

go 1.13

require (
	github.com/containernetworking/cni v1.1.2
	github.com/containernetworking/plugins v0.8.3
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.17.0
	github.com/vishvananda/netlink v1.0.0
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
