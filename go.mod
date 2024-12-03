module github.com/redhat-nfvpe/cni-route-override

go 1.22

require (
	github.com/containernetworking/cni v1.1.2
	github.com/containernetworking/plugins v0.8.3
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.17.0
	github.com/vishvananda/netlink v1.0.0
)

require (
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/vishvananda/netns v0.0.0-20180720170159-13995c7128cc // indirect
	golang.org/x/net v0.31.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace (
	golang.org/x/net => golang.org/x/net v0.7.0
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
)
