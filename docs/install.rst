Install Instructions
======================

*gotun* is written in golang. To install the tool, you will need golang on your system.


Install Go & setup the environment
-----------------------------------

You can install from your distro's package, or you can install the upstream package. After
that we will create a workspace *~/gocode/*. Add the following in your *~/.bashrc* file,
and then source it.

::

    export PATH=~/gocode/bin:$PATH
    export GOPATH=~/gocode/


Install gotun
---------------

::

    $ go get github.com/kushaldas/gotun

After this you should have the gotun binary in the *~/gocode/bin/* directory.