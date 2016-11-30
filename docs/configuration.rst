Configuration of Jobs
======================

gotun expects the job configuration in a yaml file. The following are two
different examples of the job. Each job has two files, one is the yaml file
which contains the configuration (say AWS or Openstack), and also the jobname.txt
file which contains the commands to execute.


Openstack based job
-------------------

::

    ---
    BACKEND: "openstack"

    OS_AUTH_URL: "URL"
    OS_TENANT_ID: "Your tenant id"
    OS_USERNAME: "USERNAME"
    OS_PASSWORD: "PASSWORD"
    OS_REGION_NAME: "RegionOne"
    OS_IMAGE: "Fedora-Atomic-24-20161031.0.x86_64.qcow2"
    OS_FLAVOR: "m1.medium"
    OS_SECURITY_GROUPS:
        - "group1"
        - "default"
    OS_NETWORK: "NETWORK_POOL_ID"
    OS_FLOATING_POOL: "POOL_NAME"
    OS_KEYPAIR: "KEYPAIR NAME"
    key: "Full path to the private key (.pem file)"

In the above example *gotun* expects the Image is already available in the
cloud. If you want to upload a new image for the test, and then delete it after
the test, then provide a full path to the image .qcow2 file in *OS_IMAGE*.
::

    OS_IMAGE: "/home/kdas/Fedora-Atomic-24-20161031.0.x86_64.qcow2"


AWS based job
--------------

::

    ---
    BACKEND: "aws"

    AWS_AMI: "ami-df3367bf"
    AWS_INSTANCE: "t2.medium"
    AWS_KEYNAME: "The name of the key"
    AWS_SUBNET:  "subnet-ID"
    AWS_SECURITYGROUPIDS:
        - "sg-groupid"
    AWS_REGION: "us-west-1"
    AWS_KEY: "YOURKEY"
    AWS_SECRET: "SECRET KEY PART"
    key: "PATH to the .pem file"

Update the configuration based on your need. You can see that you will need to
find subnet-id, security group ids for each region to work with.

For remote systems
-------------------

::

    ---
    BACKEND: "bare"
    key: "Path to the .pem file"
    IP: "IP of the remote system"
    PORT: 22
    USER: "username"


.. note:: The default username is *fedora*, and default port is *22*.



jobname.txt
------------

This text file contains the bash commands to run in the system, one command per line. In case you are
rebooting the system, you may want to use **SLEEP NUMBER_OF_SECONDS** command there.

If a command starts with @@ sign, it means the command is supposed to fail. Generally, we check the return codes
of the commands to find if it failed, or not. For Docker container-based systems, we track the stderr output.

We can also have non-gating tests, means these tests can pass or fail, but the whole job status will depend
on other gating tests. Any command in jobname.txt starting with ## sign will mark the test as non-gating.

Example::

    ## curl -O https://kushal.fedorapeople.org/tunirtests.tar.gz
    ls /
    ## foobar
    ## ls /root
    ##  sudo ls /root
    date
    @@ sudo reboot
    SLEEP 40
    ls /etc

