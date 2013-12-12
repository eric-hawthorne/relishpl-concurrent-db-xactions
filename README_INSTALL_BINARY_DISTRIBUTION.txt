RELISH INSTALLATION NOTES - BINARY DISTRIBUTION
===============================================

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

   Windows is not a supported platform for relish yet. Maybe once Windows adopts sensible filesystem path conventions, it will be,
   or maybe after the grumbling developer generalizes all the path-convention dependent code in relish. 
   relish is currently mostly for developing server-side internet applications, and why would you want
   to be doing that on a Windows PC anyway? 


Steps for Linux    (See Apple Mac steps below this subsection)
---------------

2. You must open a terminal window so that you are at a bash shell prompt.
   relish requires you to issue command-line commands to install it and to run relish programs. 

3. Install the sqlite3 database software on your machine if not already installed.
   Unfortunately, this is not currently distributed with the relish binary distribution.
   
   aptitude show sqlite3
   ...and if State: not installed...
   sudo apt-get install sqlite3

   works nicely on Ubuntu,
   or equivalent package installation procedure on your linux distribution.

4. Create your desired relish home directory.
   Reference is made from here on in these instructions to a relish home directory. 
   You must CREATE this directory manually if you do not already have one. 

   It is strongly recommended that you stick to using either a direct subdirectory
   of your home directory i.e. ~/relish or else create /opt/relish as your relish home.

   Some relish tools in future MAY depend on these locations. In any case,
   difference is dangerous (because complexity of a whole system generally increases 
   as an exponential function of the number of differences in different parts)
   so why do you want to be different if you don't have to be?

5. Extract the downloaded binary distribution tarball in the relish directory you
   created, as described in the list of installation commands in the main DOWNLOAD page.

6. As root, run the ./install.sh script in the relish directory. This will place
   links in /usr/local/bin to the relish command and other commands you need to develop
   and run relish programs. /usr/local/bin is already in your PATH, so after this,
   you are able to run the command: relish ... to run a relish program.
   If you prefer, you can append ~/relish/bin (or /opt/relish/bin) to your PATH
   environment variable in your ~/.bashrc or similar file, instead of running ./install.sh

7.    




   We will use /opt/relish as an example relish home directory in further instruction steps

5. Move your relish binary distribution "tarball" file into your relish home and extract 
   the files from it.

   cd /opt/relish
   mv ~/Downloads/relish_0.0.8.darwin_amd64.tar.gz .
   tar -zxvf relish_0.0.8.darwin_amd64.tar.gz

   You should see a pl (programming language) subdirectory in /opt/relish

6. Add /opt/relish/pl/bin to your PATH environment variable

   E.g. edit ~/.bashrc and add the following line:

   export PATH=$PATH:/opt/relish/pl/bin

7. Open a new terminal window to have a shell that recognizes the new PATH.

8. Issue a command to compile and run a relish program. This command will first
   download the relish program (including source code) from the Internet, will
   create standard subdirectories of /opt/relish to put the program and its data into,
   then will load and run the program.

   relish shared.relish.pl2012/hello_application hello

   After downloading and loading the program, this should print the following to standard output:

   Hello World

   You can examine the program by looking at what is in the following relish source code file:
   /opt/relish/shared/relish/artifacts/shared.relish.pl2012/hello_application/v0001/src/hello/main.rel

9. If you want to try your hand at modifying the program (after reading about the relish language),
   then first you should create your own unshared local private copy of the program:

   relish -dev shared.relish.pl2012/hello_application

   This will create your own copy, under /opt/relish/artifacts instead of /opt/relish/shared/relish/artifacts 

   Now you can safely edit the file:

   /opt/relish/artifacts/shared.relish.pl2012/hello_application/v0001/src/hello/main.rel

   and run it with the same command as above, as long as you are not running it from a directory
   which is /opt/relish/shared/relish/artifacts or below, because that would run the shared program and
   you want to run your local private copy that you modified.
   As long as your shell working directory is not somewhere in the shared artifacts tree, then

   relish shared.relish.pl2012/hello_application hello

   will run your local copy of the program.

10. If you want to download and install an upgrade of the relish binary distribution, 
   -Obtain the new tarball file (e.g. relish_0.0.9.darwin_amd64.tar.gz) and place it in /opt/relish
   -Move or remove the old /opt/relish/pl directory   e.g. rm -fr pl
   -NOTE: You should KEEP your artifacts and data and shared and data_for_shared directories.
   -Extract the new tarball.   tar -zxvf relish_0.0.9.darwin_amd64.tar.gz

   HOW DO THE NEW STANDARD LIBRARIES VERSIONS GET UPDATED LOCALLY ON A DIST UPGRADE??? DOES NEXT relish
   COMMAND EXECUTION CHECK VERSION OF SOMETHING AND IF OUT OF DATE LOAD NEW ONES?


Steps for Apple Mac
-------------------

2. You must open a terminal window so that you are at a bash shell prompt.
   relish requires you to issue command-line commands to install it and to run relish programs. 

3. Create your desired relish home directory.
   Reference is made from here on in these instructions to a relish home directory. 
   You must CREATE this directory manually if you do not already have one. 
   Several standard locations are recognized by the relish tools, so that if you
   choose to create a standard-location relish home directory you don't have to set
   an environment variable to tell relish where the relish home is.

   Standard locations (create one of these directories if it does not exist already):

   mkdir ~/relish             - creates relish subdirectory of your user home dir e.g. /Users/eric/relish
   mkdir /Library/relish      - may require your account password - do I need to sudo? - CHECK !!!!!          

   If you choose a different directory, be aware that the directory must be named relish.
   Also, in this case, you must edit your ~/.bash_profile file or similar to add a line that sets an 
   environment variable to tell relish tools where to find your relish home:
   Add a line like this to your profile file:

   export RELISH_HOME=/my/random/location/relish

   We will use ~/relish as an example relish home directory in further instruction steps

4. Move your relish binary distribution "tarball" file into your relish home and extract 
   the files from it.

   cd ~/relish
   mv ~/Downloads/relish_0.0.8.darwin_amd64.tar.gz .
   tar -xvf relish_0.0.8.darwin_amd64.tar.gz

   You should see a pl (programming language) subdirectory in ~/relish

5. Add ~/relish/pl/bin to your PATH environment variable

   FIND OUT THE BEST PROCEDURE ON MAC TO DO THIS - THERE IS SOME DIRECTORY AND YOU
   CREATE A FILE FOR JUST THIS PATH ENTRY !!!!!

6. Open a new terminal window to have a shell that recognizes the new PATH.

7. Issue a command to compile and run a relish program. This command will first
   download the relish program (including source code) from the Internet, will
   create standard subdirectories of ~/relish to put the program and its data into,
   then will load and run the program.

   relish shared.relish.pl2012/hello_application hello

   After downloading and loading the program, this should print the following to standard output:

   Hello World

   You can examine the program by looking at what is in the following relish source code file:
   ~/relish/shared/relish/artifacts/shared.relish.pl2012/hello_application/v0001/src/hello/main.rel

8. If you want to try your hand at modifying the program (after reading about the relish language),
   then first you should create your own unshared local private copy of the program:

   relish -dev shared.relish.pl2012/hello_application

   This will create your own copy, under ~/relish/artifacts instead of ~/relish/shared/relish/artifacts 

   Now you can safely edit the file:

   ~/relish/artifacts/shared.relish.pl2012/hello_application/v0001/src/hello/main.rel

   and run it with the same command as above, as long as you are not running it from a directory
   which is ~/relish/shared/relish/artifacts or below, because that would run the shared program and
   you want to run your local private copy that you modified.
   As long as your shell working directory is not somewhere in the shared artifacts tree, then

   relish shared.relish.pl2012/hello_application hello

   will run your local copy of the program.

9. If you want to download and install an upgrade of the relish binary distribution, 
   -Obtain the new tarball file (e.g. relish_0.0.9.darwin_amd64.tar.gz) and place it in ~/relish
   -Move or remove the old ~/relish/pl directory   e.g. rm -fr pl
   -NOTE: You should KEEP your artifacts and data and shared and data_for_shared directories.   
   -Extract the new tarball.   tar -xvf relish_0.0.9.darwin_amd64.tar.gz

   HOW DO THE NEW STANDARD LIBRARIES VERSIONS GET UPDATED LOCALLY ON A DIST UPGRADE??? DOES NEXT relish
   COMMAND EXECUTION CHECK VERSION OF SOMETHING AND IF OUT OF DATE LOAD NEW ONES?






