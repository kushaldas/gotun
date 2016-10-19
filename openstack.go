package main

import (
	"fmt"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/floatingip"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"github.com/rackspace/gophercloud/openstack/imageservice/v2/images"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type TunirVM struct {
	IP           string
	Hostname     string
	Port         string
	KeyFile      string
	Client       *gophercloud.ServiceClient
	Server       *servers.Server
	ClientImage  string
	FloatingIPID string
}

func (t TunirVM) Delete() error {
	res := servers.Delete(t.Client, t.Server.ID)
	if t.ClientImage != "" {
		// Delete the image we uploaded
		images.Delete(t.Client, t.ClientImage)
	}
	if t.FloatingIPID != "" {
		// Delete the Floating IP
		floatingip.Delete(t.Client, t.FloatingIPID)
	}
	return res.ExtractErr()
}

func (t TunirVM) FromKeyFile() ssh.AuthMethod {
	file := t.KeyFile
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func (t TunirVM) GetDetails() (string, string) {
	return t.IP, t.Port
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
	// Now let us find if we have to upload an image
	if strings.HasPrefix(imagename, "/") {
		// We have a qcow2 file, let us upload.
		fmt.Printf("Uploading %s to the server.\n", imagename)
		basename := filepath.Base(imagename)
		imageName := fmt.Sprintf("tunir-%s", basename)
		containerFormat := "qcow2"
		prop := make(map[string]string)
		prop["architecture"] = "x86_64"
		c, _ := openstack.NewImageServiceV2(provider, gophercloud.EndpointOpts{
			Region: region,
		})
		createResult := images.Create(c, images.CreateOpts{Name: imageName,
			Properties:      prop,
			ContainerFormat: "bare",
			DiskFormat:      containerFormat})
		image, err := createResult.Extract()
		check(err)
		image, err = images.Get(c, image.ID).Extract()
		f1, err := os.Open(imagename)
		defer f1.Close()

		putImageResult := images.Upload(c, image.ID, f1)

		if putImageResult.Err != nil {
			// Just stop the job here and get out.
			fmt.Println(err)
			images.Delete(client, image.ID)
			os.Exit(300)
		}

		// Everything okay.
		tvm.ClientImage = image.ID
		imagename = imageName

	}
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
		os.Exit(301)
	}
	tvm.Server = server
	tvm.Client = client
	fmt.Printf("Server ID: %s\n", server.ID)
	//TODO: Wait for status here
	fmt.Println("Let us wait for the server to be in running state.")
	servers.WaitForStatus(client, server.ID, "ACTIVE", 60)
	fmt.Println("Time to assign a floating pointip.")
	fp, err := floatingip.Create(client, floatingip.CreateOpts{Pool: floating_pool}).Extract()

	if err != nil {
		// There is an error in getting floating IP
		fmt.Println(err)
		tvm.Delete()
		os.Exit(302)
	}
	// Now let us assign a floating IP
	floatingip.AssociateInstance(client, floatingip.AssociateOpts{
		ServerID:   server.ID,
		FloatingIP: fp.IP,
	})
	tvm.FloatingIPID = fp.ID
	tvm.IP = fp.IP
	tvm.KeyFile = viper.GetString("key")
	tvm.Port = "22"
	return tvm, nil

}
