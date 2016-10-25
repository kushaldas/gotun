package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/viper"

	"fmt"
	"time"
)

func BootInstanceAWS() (TunirVM, error) {
	var tvm TunirVM
	tvm.VMType = "aws"
	region := viper.GetString("AWS_REGION")
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(viper.GetString("AWS_KEY"), viper.GetString("AWS_SECRET"), "")}))
	// Specify the details of the instance that you want to create.
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		// An Amazon Linux AMI ID for t2.micro instances in the us-west-2 region
		ImageId:          aws.String("ami-055b1265"),
		InstanceType:     aws.String("t2.medium"),
		KeyName:          aws.String("kushal-tunir"),
		SubnetId:         aws.String("subnet-2d0c9448"),
		SecurityGroupIds: aws.StringSlice([]string{"sg-cf95d7aa"}),
		MinCount:         aws.Int64(1),
		MaxCount:         aws.Int64(1),
	})

	if err != nil {
		fmt.Println("Could not create instance", err)
		return tvm, err
	}
	tvm.AWS_Client = *svc
	ins := *runResult.Instances[0]
	// Now we will wait for 60 seconds before refreshing the information.
	fmt.Println("Waiting for 60 seconds before checking the instance.")
	time.Sleep(time.Duration(60) * time.Second)
	// Now please get the data from the server
	params := &ec2.DescribeInstancesInput{

		InstanceIds: []*string{ins.InstanceId},


	}
	resp, err := svc.DescribeInstances(params)
	if err != nil {
		fmt.Println("Error in describing the instance.")
		return tvm, err
	}
	ins = *resp.Reservations[0].Instances[0]
	fmt.Println("The instance ID:", *ins.InstanceId)
	tvm.IP = *ins.PublicIpAddress
	tvm.Port = viper.GetString("PORT")
	tvm.KeyFile = viper.GetString("Key")
	tvm.AWS_INS = ins
	return tvm, nil
}
