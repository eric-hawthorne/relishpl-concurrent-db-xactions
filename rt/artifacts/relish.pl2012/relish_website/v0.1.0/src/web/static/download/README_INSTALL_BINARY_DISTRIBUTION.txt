RELISH INSTALLATION NOTES - BINARY DISTRIBUTION
===============================================


Steps for Linux and Apple Mac   
-----------------------------

The following notes are in moderately technical linux/unix administration language.
They are included to help you cure any issues you encounter in trying to follow
the list of installation command-line commands given in the main DOWNLOAD page.
That list of commands should be sufficient, but you can check here if you need 
motivating discussion or further explanation of each installation step.


1. Make sure you download the correct binary distribution for your computing platform.
   Choices are: darwin_amd64    - for Apple Macs
                linux_amd64     - for 64-bit linux OS installs on recent PC hardware
                linux_386       - for 32-bit linux OS installs on Atom subnotebooks/netbooks 
                                  or older PC hardware or 32-bit linux virtual machines (client slices) on a virtualized server.

   Determining whether you have 64-bit linux or 32-bit:
   You'd think this would be straightforward. But here are some suggested ways, thanks to linuxquestions.org:
   a) getconf LONG_BIT will show you the number of bits in a LONG, which typically matches the architecture type of the OS (i.e. 32 bits for a 32 bit OS and 64 bits for a 64 bit OS).
   b) uname -a
   Output may look somewhat like: Linux <hostname> 3.2.0-35-generic #55-Ubuntu SMP Wed Dec 5 17:42:16 UTC 2012 x86_64 x86_64 x86_64 GNU/Linux
   for 64-bit Linux, 
   versus something like:
   Linux <hostname> 2.6.24-16-server #1 SMP Thu Apr 10 13:58:00 UTC 2008 i686 GNU/Linux
   or
   Linux <hostname> 2.6.9-22.ELsmp #1 SMP Mon Sep 19 18:32:14 EDT 2005 i686 athlon i386 GNU/Linux
   for 32-bit Linux
   c) vi /boot/config-$(uname -r)
   If 64-bit Linux, this kernel configuration file should contain:
   CONFIG_X86_64=y
   CONFIG_64BIT=y
   CONFIG_X86=y

2. You must open a terminal window so that you are at a bash shell prompt.
   relish requires you to issue command-line commands to install it and to run relish programs.
   (On Mac, you can type 'terminal' into the Spotlight search to launch a terminal window.) 

3. ( SKIP STEP 3. on Apple Mac. sqlite3 comes pre-installed with MACOSX 10.5+ ) 
   
   Install the sqlite3 database software on your machine if not already installed.
   Unfortunately, this is not currently distributed with the relish binary distribution.
   
   aptitude show sqlite3
   ...and if State: not installed...
   sudo apt-get install sqlite3

   works nicely on Ubuntu,
   or equivalent package installation procedure on your linux distribution.

4. Create your relish home directory.
   You must CREATE this directory manually if you do not already have one. 

   It is strongly recommended that you stick to using either a direct subdirectory
   of your home directory i.e. ~/relish or else create /opt/relish as your relish home.

   Some relish tools in future MAY depend on these exact locations. In any case,
   difference is dangerous (because complexity of a whole system generally increases 
   as an exponential function of the number of differences in different parts)
   so why do you want to be different if you don't have to be?

   ~/relish will be used as an example in the discussion below.
   
5. Extract the downloaded binary distribution tarball in the relish directory you
   created, as described in the list of installation commands in the main DOWNLOAD page.

6. As root, run the ./install.sh script in the relish directory. This will place
   links in /usr/local/bin to the relish command and other commands you need to develop
   and run relish programs. /usr/local/bin is already in your PATH, so after this,
   you are able to run the command: relish ... to run a relish program.
   If you prefer, you can append ~/relish/bin to your PATH
   environment variable in your ~/.bashrc or similar file, instead of running ./install.sh

7. Try running your first relish program.
   The "relish" command that you just finished installing is used to run programs,
   as well as for a number of different development utility purposes.
   
   relish shared.relish.pl2012/hello_application  
   
   Note that this will first download the shared.relish.pl/hello_application relish
   software artifact from http://shared.relish.pl, will verify the authenticity and
   uncorruptedness of the artifact, then will install it locally, compile its
   relish source code, load .rlc intermediate-code into the runtime environment,
   then execute the main function in the main package of the software artifact. 
   
   Executing this program should print the version of the relish interpreter,
   then print an international version of "Hello World".

   You can find the auto-installed code of the downloaded artifact under the directory
   ~/relish/shared/relish/replicas/shared.relish.pl2012/hello_application/


8. If you want to download and install an upgrade of the relish binary distribution, 
   -Obtain the new tarball file (e.g. relish_0.1.1.darwin_amd64.tar.gz) and place it in ~/relish
   -cd ~/relish
   -Move or remove the old ~/relish/pl directory   e.g. mv pl pl0.0.9
   -NOTE: You should KEEP the other directories under ~/relish, such as the 
   keys, artifacts, data, shared, data_for_shared directories.
   
   -Extract the new tarball file:   tar -zxvf relish_0.1.1.darwin_amd64.tar.gz







