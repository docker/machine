module github.com/docker/machine

go 1.13

require (
	github.com/Azure/azure-sdk-for-go v5.0.0-beta+incompatible
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Azure/go-autorest v7.2.1+incompatible
	github.com/Sirupsen/logrus v0.0.0 // indirect
	github.com/aws/aws-sdk-go v1.4.10
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/bugsnag/bugsnag-go v1.0.6-0.20151120182711-02e952891c52
	github.com/bugsnag/osext v0.0.0-20130617224835-0dd3f918b21b // indirect
	github.com/bugsnag/panicwrap v0.0.0-20160118154447-aceac81c6e2f // indirect
	github.com/cenkalti/backoff v0.0.0-20141124221459-9831e1e25c87 // indirect
	github.com/codegangsta/cli v1.11.1-0.20151120215642-0302d3914d2a
	github.com/dgrijalva/jwt-go v3.0.1-0.20160831183534-24c63f56522a+incompatible // indirect
	github.com/digitalocean/godo v1.0.1-0.20170317202744-d59ed2fe842b
	github.com/docker/docker v1.13.1
	github.com/docker/go-units v0.2.1-0.20151230175859-0bbddae09c5a // indirect
	github.com/exoscale/egoscale v0.9.23
	github.com/go-ini/ini v0.0.0-20151124192405-03e0e7d51a13 // indirect
	github.com/google/go-querystring v0.0.0-20140804062624-30f7a39f4a21 // indirect
	github.com/gophercloud/gophercloud v0.4.0
	github.com/gophercloud/utils v0.0.0-20190829151529-94e6842399e5
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/intel-go/cpuid v0.0.0-20181003105527-1a4a6f06a1c6
	github.com/jinzhu/copier v0.0.0-20180308034124-7e38e58719c3 // indirect
	github.com/jmespath/go-jmespath v0.0.0-20151117175822-3433f3ea46d9 // indirect
	github.com/juju/loggo v0.0.0-20190526231331-6e530bcce5d8 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mitchellh/mapstructure v0.0.0-20140721150620-740c764bc614 // indirect
	github.com/onsi/ginkgo v1.10.1 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/rackspace/gophercloud v1.0.1-0.20150408191457-ce0f487f6747
	github.com/samalba/dockerclient v0.0.0-20151231000007-f661dd4754aa
	github.com/sirupsen/logrus v0.0.0 // indirect
	github.com/skarademir/naturalsort v0.0.0-20150715044055-69a5d87bef62
	github.com/smartystreets/goconvey v0.0.0-20190731233626-505e41936337 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.3.0
	github.com/tent/http-link-go v0.0.0-20130702225549-ac974c61c2f9 // indirect
	github.com/vmware/govcloudair v0.0.2
	github.com/vmware/govmomi v0.6.2
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/net v0.0.0-20190311183353-d8887717615a
	golang.org/x/oauth2 v0.0.0-20151117210313-442624c9ec92
	golang.org/x/sys v0.0.0-20190422165155-953cdadca894
	google.golang.org/api v0.0.0-20180213000552-87a2f5c77b36
	google.golang.org/appengine v0.0.0-20160205025855-6a436539be38 // indirect
	google.golang.org/cloud v0.0.0-20151119220103-975617b05ea8 // indirect
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
)

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.0.4
	github.com/sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
)
