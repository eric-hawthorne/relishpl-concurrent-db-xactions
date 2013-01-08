RELISH INSTALLATION AND BUILD INSTRUCTIONS
==========================================

BINARY DISTRIBUTION


1. Make sure it is the correct binary distribution for your platform.
  Choices are: darwin_amd64    - for Apple Macs
               linux_amd64     - for 64-bit linux OS installs on recent PC hardware
               linux_386       - for 32-bit linux OS installs on Atom subnotebooks/netbooks 
                                 or older PC hardware

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
   Several standard locations are recognized by the relish tools, so that if you
   choose to create a standard-location relish home directory you don't have to set
   an environment variable to tell relish where the relish home is.

   Standard locations (create one of these directories if it does not exist already):

   sudo mkdir /opt/relish
   sudo chmod go+w /opt/relish     - make the relish home directory writable by other than root
         
   mkdir ~/relish             - creates relish subdirectory of your user home dir e.g. /home/eric/relish   

   sudo mkdir /usr/local/lib/relish
   sudo chmod go+w /usr/local/lib/relish     - make the relish home directory writable by other than root

   sudo mkdir /usr/local/relish  
   sudo chmod go+w /usr/local/relish     - make the relish home directory writable by other than root

   If you choose a different directory, be aware that the directory must be named relish.
   Also, in this case, you must edit your ~/.profile or ~/.bashrc file or similar to add a line that sets an 
   environment variable to tell relish tools where to find your relish home:
   Add a line like this to your profile file:

   export RELISH_HOME=/my/random/location/relish

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

   relish examples.relish.pl2012/hello_application hello

   After downloading and loading the program, this should print the following to standard output:

   Hello World

   You can examine the program by looking at what is in the following relish source code file:
   /opt/relish/shared/relish/artifacts/examples.relish.pl2012/hello_application/v0001/src/hello/main.rel

9. If you want to try your hand at modifying the program (after reading about the relish language),
   then first you should create your own unshared local private copy of the program:

   relish -dev examples.relish.pl2012/hello_application

   This will create your own copy, under /opt/relish/artifacts instead of /opt/relish/shared/relish/artifacts 

   Now you can safely edit the file:

   /opt/relish/artifacts/examples.relish.pl2012/hello_application/v0001/src/hello/main.rel

   and run it with the same command as above, as long as you are not running it from a directory
   which is /opt/relish/shared/relish/artifacts or below, because that would run the shared program and
   you want to run your local private copy that you modified.
   As long as your shell working directory is not somewhere in the shared artifacts tree, then

   relish examples.relish.pl2012/hello_application hello

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

   relish examples.relish.pl2012/hello_application hello

   After downloading and loading the program, this should print the following to standard output:

   Hello World

   You can examine the program by looking at what is in the following relish source code file:
   ~/relish/shared/relish/artifacts/examples.relish.pl2012/hello_application/v0001/src/hello/main.rel

8. If you want to try your hand at modifying the program (after reading about the relish language),
   then first you should create your own unshared local private copy of the program:

   relish -dev examples.relish.pl2012/hello_application

   This will create your own copy, under ~/relish/artifacts instead of ~/relish/shared/relish/artifacts 

   Now you can safely edit the file:

   ~/relish/artifacts/examples.relish.pl2012/hello_application/v0001/src/hello/main.rel

   and run it with the same command as above, as long as you are not running it from a directory
   which is ~/relish/shared/relish/artifacts or below, because that would run the shared program and
   you want to run your local private copy that you modified.
   As long as your shell working directory is not somewhere in the shared artifacts tree, then

   relish examples.relish.pl2012/hello_application hello

   will run your local copy of the program.

9. If you want to download and install an upgrade of the relish binary distribution, 
   -Obtain the new tarball file (e.g. relish_0.0.9.darwin_amd64.tar.gz) and place it in ~/relish
   -Move or remove the old ~/relish/pl directory   e.g. rm -fr pl
   -NOTE: You should KEEP your artifacts and data and shared and data_for_shared directories.   
   -Extract the new tarball.   tar -xvf relish_0.0.9.darwin_amd64.tar.gz

   HOW DO THE NEW STANDARD LIBRARIES VERSIONS GET UPDATED LOCALLY ON A DIST UPGRADE??? DOES NEXT relish
   COMMAND EXECUTION CHECK VERSION OF SOMETHING AND IF OUT OF DATE LOAD NEW ONES?






