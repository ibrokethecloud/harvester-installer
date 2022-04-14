package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/harvester/harvester-installer/pkg/util"
)

func TestToHarvesterConfig(t *testing.T) {
	testCases := []struct {
		input    []byte
		expected *HarvesterConfig
		err      error
	}{
		{
			input: util.LoadFixture(t, "harvester-config.yaml"),
			expected: &HarvesterConfig{
				ServerURL: "https://someserver:6443",
				Token:     "TOKEN_VALUE",
				OS: OS{
					SSHAuthorizedKeys: []string{
						"ssh-rsa AAAAB3NzaC1yc2EAAAADAQAB...",
						"github:username",
					},
					Hostname: "myhost",
					Modules: []string{
						"kvm",
						"nvme",
					},
					Sysctls: map[string]string{
						"kernel.printk":        "4 4 1 7",
						"kernel.kptr_restrict": "1",
					},
					DNSNameservers: []string{
						"8.8.8.8",
						"1.1.1.1",
					},
					NTPServers: []string{
						"0.us.pool.ntp.org",
						"1.us.pool.ntp.org",
					},
					Wifi: []Wifi{
						{
							Name:       "home",
							Passphrase: "mypassword",
						},
						{
							Name:       "nothome",
							Passphrase: "somethingelse",
						},
					},
					Password: "rancher",
					Environment: map[string]string{
						"http_proxy":  "http://myserver",
						"https_proxy": "http://myserver",
					},
				},
				Install: Install{
					Mode: "create",
					Networks: map[string]Network{
						MgmtInterfaceName: {
							Interfaces: []NetworkInterface{{Name: "ens0"}, {Name: "ens3"}},
							Method:     "dhcp",
						},
					},
					ForceEFI: true,
					Device:   "/dev/vda",
					Silent:   true,
					ISOURL:   "http://myserver/test.iso",
					PowerOff: true,
					NoFormat: true,
					Debug:    true,
					TTY:      "ttyS0",
				},
				LogCollector: LogCollector{
					Enabled: true,
					UploadConfig: UploadConfig{
						NFSConfig: &NFS{
							Endpoint: "localhost",
						},
						ObjectStoreConfig: &ObjectStore{
							Endpoint:        "play.min.io",
							AccessKeyID:     "Q3AM3UQ867SPQQA43P2F",
							SecretAccessKey: "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG",
							BucketName:      "harv-test",
						},
					},
				},
			},
			err: nil,
		},
	}

	for _, testCase := range testCases {
		output, err := LoadHarvesterConfig(testCase.input)
		assert.Equal(t, testCase.expected, output)
		assert.Equal(t, testCase.err, err)
	}
}
