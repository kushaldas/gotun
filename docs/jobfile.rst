Detailed description of the tests
==================================

The *jobname.txt* text file contains the bash commands to run in the system, one command per line. In case you are
rebooting the system, you may want to use **SLEEP NUMBER_OF_SECONDS** directive there.

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


Multiple VM based tests
-------------------------

In case of tests containing multiple VM(s), one mark the tests with vm numbers. This way, we decide which test will
run on which vm. The numbers start from *vm1* to *vm9*.

Example::

    vm1 wget https://kushaldas.in
    vm2 sudo mkdir /root/hello_dir
    vm1 sudo dnf install pss -y
    vm1 which pss

If no vm number is marked at the begining of any line, gotun assumes that the test is supposed to run on *vm1*.


Rebuild of VM(s) on OpenStack
------------------------------

.. note:: This feature is only available for OpenStack based jobs. For other kind of tests, this will do nothing.

*REBUILD_SERVERS* directive will rebuild all of the avialable VM(s) on OpenStack. They will try to POLL the VM(s) after
rebuilding them. This step is sequential for now. In future, we will be doing this in parallal.
::

    echo "hello asd" > ./hello.txt
    vm1 sudo cat /etc/machine-id
    mkdir {push,pull}
    ls -l ./
    pwd
    REBUILD_SERVERS
    sudo cat /etc/machine-id
    ls -l ./
    pwd

The following is the output from the above mentioned test.
::

    $ gotun --job fedora
    Starts a new Tunir Job.

    Server ID: e0d7b55a-f066-4ff8-923c-582f3c9be29b
    Let us wait for the server to be in running state.
    Time to assign a floating pointip.
    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Server ID: a0b810e6-0d7f-4c9e-bc4d-1e62b082673d
    Let us wait for the server to be in running state.
    Time to assign a floating pointip.
    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Executing:  echo "hello asd" > ./hello.txt
    Executing:  vm1 sudo cat /etc/machine-id
    Executing:  mkdir {push,pull}
    Executing:  ls -l ./
    Executing:  pwd
    Going to rebuild: 209.132.184.241
    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Going to rebuild: 209.132.184.242
    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Polling for a successful ssh connection.

    Executing:  sudo cat /etc/machine-id
    Executing:  ls -l ./
    Executing:  pwd

    Result file at: /tmp/tunirresult_180507156


    Job status: true


    command: echo "hello asd" > ./hello.txt
    status:true



    command: sudo cat /etc/machine-id
    status:true

    e0d7b55af0664ff8923c582f3c9be29b


    command: mkdir {push,pull}
    status:true



    command: ls -l ./
    status:true

    total 4
    -rw-rw-r--. 1 fedora fedora 10 Jan 25 13:58 hello.txt
    drwxrwxr-x. 2 fedora fedora  6 Jan 25 13:58 pull
    drwxrwxr-x. 2 fedora fedora  6 Jan 25 13:58 push


    command: pwd
    status:true

    /var/home/fedora


    command: sudo cat /etc/machine-id
    status:true

    e0d7b55af0664ff8923c582f3c9be29b


    command: ls -l ./
    status:true

    total 0


    command: pwd
    status:true

    /var/home/fedora


    Total Number of Tests:8
    Total NonGating Tests:0
    Total Failed Non Gating Tests:0

    Success.

Creating inventory file for Ansible based tests
------------------------------------------------

`Ansible <https://www.ansible.com/>`_ is a powerful choice with many different usecases. One such usecase is about testing.
Sometimes we just setup the whole test environment using Ansible, and some other times the whole testsuite is written
on top of ansible. To enable using of predefined Ansible playbooks, gotun provides a file *current_run_info.json* for each
run of job. This file contains a dictionary of vm numbers, and corresponding IP address, and also the *keyfile* value with
the path of the private keyfile. This can be used with a simple Python or shell script to create the actual inventory file.
For example, the following script *createinventory.py* will create a file called *inventory* in the current directory, and it assumes that there
will be 2 VM(s) are avaiable (means it is running on OpenStack).

::

    #!/usr/bin/env python3
    import json

    data = None
    with open("current_run_info.json") as fobj:
        data = json.loads(fobj.read())

    user = data['user']
    host1 = data['vm1']
    host2 = data['vm2']
    key = data['keyfile']

    result = """{0} ansible_ssh_host={1} ansible_ssh_user={2} ansible_ssh_private_key_file={3}
    {4} ansible_ssh_host={5} ansible_ssh_user={6} ansible_ssh_private_key_file={7}""".format(host1,host1,user,key,host2,host2,user,key)
    with open("inventory", "w") as fobj:
        fobj.write(result)

As you can see, we are reading the *current_run_info.json* file first, and then creating a file called *inventory*. We can
then execute this script by using the *HOSTCOMMAND* directive in the test.
::

    HOSTCOMMAND: ./createinventory.py


Running Ansible on the HOST as part of a test
----------------------------------------------

The next step is to run Ansible playbook on the host system as a test. This can be done with a *HOSTTEST* directive. The
following example test file will first create the inventory file using a *HOSTCOMMAND* directive, and then execute the an
ansible playbook.
::

    HOSTCOMMAND: ./onevm.py
    HOSTTEST: ansible-playbook -b -i inventory atomic-host-tests/tests/improved-sanity-test/main.yml
