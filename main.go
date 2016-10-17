package main

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/floatingip"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

type TVM interface {
	Delete() error
}

type TunirVM struct {
	IP string
	Hostname string
	Port string
	KeyFile string
	Client *gophercloud.ServiceClient
	Server *servers.Server
}

func (t TunirVM) Delete() error {
	res := servers.Delete(t.Client, t.Server.ID)
	return res.ExtractErr()
}

func BootInstanceOS() (TVM, error) {
	var tvm TunirVM
	// If no config is found, use the default(s)
	viper.SetDefault("OS_REGION_NAME", "RegionOne")
	viper.SetDefault("OS_FLAVOR", "m1.medium")


	opts := gophercloud.AuthOptions{
		IdentityEndpoint: viper.GetString("OS_AUTH_URL"),
		Username:         viper.GetString("OS_USERNAME"),
		Password:         viper.GetString("OS_PASSWORD"),
		TenantID:         viper.GetString("OS_TENANT_ID"),
	}
	region := viper.GetString("OS_REGION_NAME")
	provider, err := openstack.AuthenticatedClient(opts)
	client, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: region})


	vmflavor := viper.GetString("OS_FLAVOR")
	imagename := viper.GetString("OS_IMAGE")
	network_id := viper.GetString("OS_NETWORK")
	floating_pool := viper.GetString("OS_FLOATING_POOL")
	keypair := viper.GetString("OS_KEYPAIR")
	security_groups := viper.GetStringSlice("OS_SECURITY_GROUPS")

	sOpts := servers.CreateOpts{
		Name:           "gotun",
		FlavorName:     vmflavor,
		ImageName:      imagename,
		SecurityGroups: security_groups,
	}
	sOpts.Networks = []servers.Network{
		{
			UUID: network_id,
		},
	}

	server, err := servers.Create(client, keypairs.CreateOptsExt{
		sOpts,
		keypair,
	}).Extract()
	if err != nil {
		fmt.Println("Unable to create server: %s", err)
		return tvm, err
	}
	tvm.Server = server
	tvm.Client = client
	fmt.Printf("Server ID: %s booted.\n", server.ID)
	//TODO: Wait for status here
	fmt.Println("Let us wait for the server to be in running state.")
	servers.WaitForStatus(client, server.ID, "ACTIVE", 60)
	fmt.Println("Time to assign a floating pointip.")
	fp, err := floatingip.Create(client, floatingip.CreateOpts{Pool: floating_pool}).Extract()
	fmt.Println(fp)
	// Now let us assign
	floatingip.AssociateInstance(client, floatingip.AssociateOpts{
		ServerID: server.ID,
		FloatingIP: fp.IP,
	})
	tvm.IP = fp.IP
	tvm.Port = "22"
	return tvm, nil

}

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()

	if err != nil {
		fmt.Println("No configuration file loaded - using defaults")
	}

	backend := viper.GetString("BACKEND")
	fmt.Println(backend)
	if backend == "openstack" {
		vm, _ := BootInstanceOS()
		fmt.Println(vm)
	}

}
