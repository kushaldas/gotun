package main

import (
	"fmt"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/floatingip"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"github.com/rackspace/gophercloud/openstack/imageservice/v2/images"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
)

type TunirResult struct {
	Output  string
	Status  bool
	Command string
}

type ResultSet struct {
	Results                   []TunirResult
	Status                    bool // Whole status of the job
	TotalTests                int
	TotalNonGatingTests       int
	TotalFailedNonGatingTests int
}

type TVM interface {
	Delete() error
	FromKeyFile() ssh.AuthMethod
	GetDetails() (string, string)
}

type TunirVM struct {
	VMType       string
	IP           string
	Hostname     string
	Port         string
	KeyFile      string
	Client       *gophercloud.ServiceClient
	Server       *servers.Server
	ClientImage  string
	FloatingIPID string
	AWS_INS      ec2.Instance
	AWS_Client   ec2.EC2
}

func (t TunirVM) Delete() error {
	if t.VMType == "openstack" {
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
	} else if t.VMType == "aws" {
		params := &ec2.TerminateInstancesInput{
			InstanceIds: []*string{ // Required
				t.AWS_INS.InstanceId, // Required
			},
		}
		_, err := t.AWS_Client.TerminateInstances(params)

		if err != nil {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
			return err
		}
	}
	return nil
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

//Poll keeps trying to ssh into the vm.
//We need to this to test if the vm is ready for work.
func Poll(timeout int64, vm TVM) bool {
	ip, port := vm.GetDetails()
	sshConfig := &ssh.ClientConfig{
		User: viper.GetString("USER"),
		Auth: []ssh.AuthMethod{
			vm.FromKeyFile(),
		},
	}
	start := time.Now().Unix()
	for {

		fmt.Println("Polling for a successful ssh connection.\n")
		time.Sleep(5 * time.Second)
		currenttime := time.Now().Unix()
		difftime := currenttime - start
		// Check for timeout
		if timeout >= 0 && difftime >= timeout {
			return false
		}

		// Execute the function
		connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", ip, port), sshConfig)
		if err != nil {
			//TODO: enable for debugging
			//fmt.Println(err)
			continue
		}
		session, err := connection.NewSession()
		if err != nil {
			return false
		}
		session.Close()
		return true

	}
	return false
}

//printResultSet prints the whole test run result to a file, and also on STDOUT.
func printResultSet(result ResultSet) {
	file, _ := ioutil.TempFile(os.TempDir(), "tunirresult_")
	fmt.Println("\nResult file at:", file.Name())
	status := result.Status
	results := result.Results
	fmt.Printf("\n\nJob status: %v\n", status)
	for _, value := range results {
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

//ReadCommands returns a slice of strings with all the commands.
func ReadCommands(filename string) []string {
	data, _ := ioutil.ReadFile(filename)
	return strings.Split(string(data), "\n")
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

//ExecuteTests runs the given commands in the VM.
func ExecuteTests(commands []string, vm TVM) ResultSet {
	var actualcommand string
	var willfail, dontcare bool
	var parts []string
	var output []byte
	var session *ssh.Session

	FinalResult := ResultSet{}
	result := make([]TunirResult, 0)

	vmr, _ := regexp.Compile("^vm[0-9] ")
	ip, port := vm.GetDetails()
	sshConfig := &ssh.ClientConfig{
		User: viper.GetString("USER"),
		Auth: []ssh.AuthMethod{
			vm.FromKeyFile(),
		},
	}

	for i := range commands {
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
		if err != nil {
			output = []byte(err.Error())
			goto ERROR1
		}
		session, err = connection.NewSession()
		if err != nil {
			output = []byte(err.Error())
			goto ERROR1
		}
		defer session.Close()
		fmt.Println("Executing: ", actualcommand)
		output, err = session.CombinedOutput(actualcommand)
		FinalResult.TotalTests += 1
		if dontcare {
			FinalResult.TotalNonGatingTests += 1
		}
	ERROR1:
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
