Configuration of Jobs
======================

gotun expects the job configuration in a yaml file. The following are two
different examples of the job. Each job has two files, one is the yaml file
which contains the configuration (say AWS or Openstack), and also the jobname.txt
file which contains the commands to execute.


OpenStack based job
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


Multiple VM(s) on OpenStack
----------------------------

In case you want to spin up more than one vm on OpenStack, then add a *NUMBER* value to the yml file.
::

    NUMBER: 3


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
    PORT: 22
    USER: "username"
    VMS:
        vm1: IP1
        vm2: IP2

The keys of *VMS* are the vm numbers, you will have to mark at least *vm1* and the corresponding
IP address.


.. note:: The default username is *fedora*, and default port is *22*.




