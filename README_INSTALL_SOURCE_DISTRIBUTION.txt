RELISH INSTALLATION AND BUILD INSTRUCTIONS
==========================================

SOURCE DISTRIBUTION hg clone https://code.google.com/p/relish


Steps for Linux    (See Apple Mac steps below this subsection)
---------------

1. You must open a terminal window so that you are at a bash shell prompt.
   relish requires you to issue command-line commands to install it and to run relish programs. 

2. Install the sqlite3 database software and its header files on your machine if not already installed.

   apt-cache show sqlite3
   ...and if State: not installed...
   sudo apt-get install sqlite3
   apt-cache show libsqlite3-dev
   ...and if State: not installed...
   sudo apt-get install libsqlite3-dev

   works nicely on Ubuntu,
   or equivalent package installation procedure on your linux distribution.

3. Install mercurial version control system on your machine if not already installed.

   sudo aptitude install mercurial

   or equivalent package installation procedure on your linux distribution.

4. Install Go 1.0 or higher (from golang.org)

5. Install the gosqlite package into the Go environment.

   go get code.google.com/p/gosqlite/sqlite

6. Choose where you will create your mercurial repository

   You will probably want to have your relish root i.e. relish home directory be one of the standard
   locations. Several standard locations are recognized by the relish tools, so that if you
   choose to create a standard-location relish home directory you don't have to set
   an environment variable to tell relish where the relish home is.
   So cd to one of the following locations: 
   
   cd                 - change to your home directory e.g. /home/eric 

   or

   sudo chmod go+w /opt
   cd /opt

   or 

   sudo chmod go+w /opt
   cd /opt
   mkdir devel
   cd devel   

   or 

   If you choose a different parent directory, you must edit your ~/.profile or ~/.bashrc file 
   or similar to add a line that sets an environment variable to tell relish tools where to find 
   your relish home. Add a line like this to your profile file:

   export RELISH_HOME=/my/random/location/relish

   cd /my/random/location


7. hg clone https://relishpl@code.google.com/p/relish/ 

   Note that the repo directory that you are developing in and running relish tools from must be called relish.
   If you want to have multiple clones, you can clone them into e.g. relish_mybranch then
   use a symbolic link to decide which one is active. e.g. cd /opt ; ln -s relish_mybranch relish
   

   You might want to unpermit write to the /opt dir for security now if you used that.

   sudo chmod go-w /opt


   We will use /opt/relish as an example relish home directory in further instruction steps

8. Add /opt/relish/bin to your PATH environment variable

   E.g. edit ~/.bashrc and add the following line:

   export PATH=$PATH:/opt/relish/bin

9. Tell Go where to be looking for Go source code and where to be installing packages and executables:
   E.g. edit ~/.bashrc and add the following line:

   export GOPATH=/opt/relish

10. Open a new terminal window to have a shell that recognizes the new PATH.

11. build the relish compiler-interpreter 

    go install relish/relish

    and make it an executable file:

    chmod go+x /opt/relish/bin/relish

12. Issue a command to compile and run a relish program. This command will first
   download the relish program (including source code) from the Internet, will
   create standard subdirectories of /opt/relish to put the program and its data into,
   then will load and run the program.

   relish examples.relish.pl2012/hello_application hello

   After downloading and loading the program, this should print the following to standard output:

   Hello World

   You can examine the program by looking at what is in the following relish source code file:
   /opt/relish/shared/relish/artifacts/examples.relish.pl2012/hello_application/v0001/src/hello/main.rel

13. If you want to try your hand at modifying the program (after reading about the relish language),
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




Steps for Apple Mac
-------------------

1. You must open a terminal window so that you are at a bash shell prompt.
   relish requires you to issue command-line commands to install it and to run relish programs. 

2. The sqlite3 database software should already be installed in MACOSX Leopard or higher.

   BUT DO WE HAVE TO GET THE DEV LIBRARY WITH HEADER FILES?????????? IF SO WHAT IS COMMAND???
   sudo aptitude install libsqlite3-dev   ??????????????

   Someone says install Apple XCode developer tools and you will have the sqlite3-dev library
   with sqlite3 header files that are needed by gosqlite.

3. Install mercurial version control system on your machine if not already installed.
   
   There is a binary distribution for Mac OS X. Google for that and install it, if the command
   which hg   
   does not already tell you about an hg executable.

4. Install Go 1.0 or higher (from golang.org)

5. Install the gosqlite package into the Go environment.

   go get code.google.com/p/gosqlite/sqlite

6. Choose where you will create your mercurial repository

   You will probably want to have your relish root i.e. relish home directory be one of the standard
   locations. Several standard locations are recognized by the relish tools, so that if you
   choose to create a standard-location relish home directory you don't have to set
   an environment variable to tell relish where the relish home is.
   So cd to one of the following locations: 
   
   cd                 - change to your home directory e.g. /Users/eric 

   or

   sudo chmod go+w /Library
   cd /Library 

   or 

   If you choose a different parent directory, you must edit your ~/.bash_profile file 
   or similar to add a line that sets an environment variable to tell relish tools where to find 
   your relish home. Add a line like this to your profile file:

   export RELISH_HOME=/my/random/location/relish

   cd /my/random/location


7. hg clone https://relishpl@code.google.com/p/relish/ 

   Note that the repo directory that you are developing in and running relish tools from must be called relish.
   If you want to have multiple clones, you can clone them into e.g. relish_mybranch then
   use a symbolic link to decide which one is active. e.g. cd or cd /Library; ln -s relish_mybranch relish
   
   You might want to unpermit write to the Library dir for security now if you used that.

   sudo chmod go-w /Library

   We will use ~/relish as an example relish home directory in further instruction steps

8. Add ~/relish/bin to your PATH environment variable

   FIND OUT THE BEST PROCEDURE ON MAC TO DO THIS - THERE IS SOME DIRECTORY AND YOU
   CREATE A FILE FOR JUST THIS PATH ENTRY !!!!!


9. Tell Go where to be looking for Go source code and where to be installing packages and executables:
   E.g. edit ~/.bash-profile and add the following line:

   export GOPATH=~/relish

10. Open a new terminal window to have a shell that recognizes the new PATH.

11. build the relish compiler-interpreter 

    go install relish/relish

    and make it an executable file:

    chmod go+x ~/relish/bin/relish

11. Issue a command to compile and run a relish program. This command will first
   download the relish program (including source code) from the Internet, will
   create standard subdirectories of ~/relish to put the program and its data into,
   then will load and run the program.

   relish examples.relish.pl2012/hello_application hello

   After downloading and loading the program, this should print the following to standard output:

   Hello World

   You can examine the program by looking at what is in the following relish source code file:
   ~/relish/rt/shared/relish/artifacts/examples.relish.pl2012/hello_application/v0001/src/hello/main.rel

12. If you want to try your hand at modifying the program (after reading about the relish language),
   then first you should create your own unshared local private copy of the program:

   relish -dev examples.relish.pl2012/hello_application

   This will create your own copy, under ~/relish/rt/artifacts instead of /opt/relish/rt/shared/relish/artifacts 

   Now you can safely edit the file:

   ~/relish/rt/artifacts/examples.relish.pl2012/hello_application/v0001/src/hello/main.rel

   and run it with the same command as above, as long as you are not running it from a directory
   which is /opt/relish/shared/relish/artifacts or below, because that would run the shared program and
   you want to run your local private copy that you modified.
   As long as your shell working directory is not somewhere in the shared artifacts tree, then

   relish examples.relish.pl2012/hello_application hello

   will run your local copy of the program.

