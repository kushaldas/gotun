package main

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/floatingip"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"strings"
	"time"
	"strconv"
	"regexp"
	"os"
)

type TunirResult struct {
	Output string
	Status bool
	Command string
}

type ResultSet struct {
	Results []TunirResult
	Status bool // Whole status of the job
	TotalTests int
	TotalNonGatingTests int
	TotalFailedNonGatingTests int
}

type TVM interface {
	Delete() error
	FromKeyFile() ssh.AuthMethod
	GetDetails() (string, string)
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

func printResultSet(result ResultSet) {
	file, _ := ioutil.TempFile(os.TempDir(), "tunirresult_")
	fmt.Println("\nResult file at:", file.Name())
	status := result.Status
	results := result.Results
	fmt.Printf("\n\nJob status: %v\n", status)
	for _,value := range results {
		output := fmt.Sprintf("\n\ncommand: %s\nstatus:%v\n\n%s", value.Command, value.Status, value.Output)
		fmt.Printf(output)
		file.WriteString(output)
	}

	fmt.Printf("\n\nTotal Number of Tests:%d\nTotal NonGating Tests:%d\nTotal Failed Non Gating Tests:%d\n",
		result.TotalTests, result.TotalNonGatingTests, result.TotalFailedNonGatingTests)

	if status {
		fmt.Println("\nSuccess.")
	} else {
		fmt.Println("\nFailed.")
	}
}

func ReadCommands(filename string) []string {
	data, _ := ioutil.ReadFile(filename)
	return strings.Split(string(data), "\n")
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func ExecuteTests(commands []string, vm TVM) ResultSet {
	var actualcommand string
	var willfail, dontcare bool
	var parts []string

	FinalResult := ResultSet{}
	result := make([]TunirResult,0)

	vmr, _ := regexp.Compile("^vm[0-9] ")
	ip, port := vm.GetDetails()
	sshConfig := &ssh.ClientConfig{
		User: "fedora",
		Auth: []ssh.AuthMethod{
			vm.FromKeyFile(),
		},
	}
	for i := range(commands) {
		willfail = false
		dontcare = false
		command := commands[i]
		if command != "" {
			if strings.HasPrefix(command, "SLEEP") {
				d := strings.Split(command, " ")[1]
				fmt.Println("Sleeping for ", d)
				t, _ := strconv.ParseInt(d, 10, 64)
				time.Sleep(time.Duration(t) * time.Second)
				continue
			}
			if vmr.MatchString(command) {
				parts = strings.Split(command, " ")
				actualcommand = strings.Join(parts[1:], " ")

			} else {
				actualcommand = command
			}
			parts = strings.Split(actualcommand, " ")
			if parts[0] == "@@" {
				willfail = true
				actualcommand = strings.Join(parts[1:], " ")

			} else if parts[0] == "##" {
				dontcare = true
				actualcommand = strings.Join(parts[1:], " ")

			}

		} else {
			// stupid code, revert the if else
			continue
		}

		connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", ip, port), sshConfig)
		check(err)
		session, err := connection.NewSession()
		check(err)
		defer session.Close()
		fmt.Println("Executing: ", actualcommand)
		output, err := session.CombinedOutput(actualcommand)
		FinalResult.TotalTests += 1
		if dontcare {
			FinalResult.TotalNonGatingTests += 1
		}
		rf := TunirResult{Output: string(output), Command: actualcommand}
		if err != nil {

			rf.Status = false
			if willfail || dontcare {
				result = append(result, rf)
				if dontcare {
					FinalResult.TotalFailedNonGatingTests += 1
				}
				continue
			} else {
				result = append(result, rf)
				FinalResult.TotalTests += 1
				FinalResult.Status = false
				FinalResult.Results = result
				return FinalResult
			}
		} else {
			rf.Status = true
		}
		result = append(result, rf)


	}
	FinalResult.Status = true
	FinalResult.Results = result
	return FinalResult


}

func main() {
	var vm TVM
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()

	if err != nil {
		fmt.Println("No configuration file loaded - using defaults")
	}

	viper.SetDefault("PORT", "22")
	backend := viper.GetString("BACKEND")
	fmt.Println("Starts a new Tunir Job.\n")

	if backend == "openstack" {
		vm, _ = BootInstanceOS()
		fmt.Println(vm)
	} else if backend == "bare" {
		vm = TunirVM{IP: viper.GetString("IP"), KeyFile: viper.GetString("key"),
		Port: viper.GetString("PORT")}
	}
	commands := ReadCommands("./commands.txt")
	result := ExecuteTests(commands, vm)
	printResultSet(result)
	if !result.Status {
		os.Exit(200)
	}
	os.Exit(0)
}
