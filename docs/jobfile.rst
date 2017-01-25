Detailed description of the tests
==================================

The *jobname.txt* text file contains the bash commands to run in the system, one command per line. In case you are
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
