Locations (hosts) for downloading relish source code
====================================================

Normally, relish downloads relish source code from a standard default server hostname 
for each code origin.
So for example, source code whose origin is code.coolthings.net2013 would be found at 
the server http://code.coolthings.net

If the code is not found (at port 80 or port 8421) at the standard default server hostname,
then relish will also finally try finding the code at http://shared.relish.pl, in case
the code has been uploaded to that code repository.

It is strongly recommended that you have your relish installation rely on these
standard conventions and places for where to find (and where to publish) relish source code.

--------------
ADVANCED TOPIC
--------------
However, you have the additional option of making your relish installation aware
of other, non-standard locations on the Internet (or a LAN) where relish source code
may be found.

You do that by creating configuration files in this xtras directory, where each file specifies
a list of locations where source code may be found.

----------
NOTE: IT IS STRONGLY RECOMMENDED THAT YOU DO NOT USE THESE CONFIGURATION FILES,
because doing so will render your relish installation fragile and hard to maintain
in that it is dependent on hardcoded non-standard locations for code.
YOU HAVE BEEN WARNED!
----------

Only a maximum of 4 code-locations files can reside in this xtras directory. Those
files must be named with particular filenames, and depending on the filename, there
is a slightly different meaning to how relish interprets the list of locations in the
file.

Here are the source-code locations files that you can create:
-------------------------------------------------------------
relish_code_origins.txt  - a list of code origins, with alternate servers (owned by originator) listed for each.

relish_code_replicas.txt - a list of code origins, with mirror servers (owned by others) listed for each.

relish_code_repositories.txt - repository servers each of which may have replicas of many code artifacts from many origins.

relish_code_staging_servers.txt - a list of code origins, with staging servers (owned by originator) listed for each.


Details of each source-code locations file
------------------------------------------
(Example file contents shown below each filename, between ----- lines.)

relish_code_origins.txt
-----------------------------
code.coolthings.net2012 142.37.19.236 coolcode.com s2.coolcode.com s3.coolcode.com:8088
cs.ubc.ca1992 andromeda.cs.ubc.ca whirlpool.cs.ubc.ca:8081 sombrero.cs.ubc.ca
-----------------------------

Each line of the text file begins with a relish code origin identifier and is followed
by a space-separated list of server hostnames or IP addresses.
By example the first line of this file is used to list servers:
1. which are owned or controlled by the code originator "code.coolthings.net2012"
2: which host relish code artifacts from that origin but 
3. whose host name is not "code.coolthings.net"


relish_code_replicas.txt
-----------------------------
code.coolthings.net2012 everybitcounts.net cs.waterloo.ca
cs.ubc.ca1992 arnold.uwash.edu
alphaworks.ibm.com1999 asterix.cs.cornell.edu
-----------------------------

Each line of the text file begins with a relish code origin identifier and is followed
by a space-separated list of server hostnames or IP addresses.
By example the first line of this file is used to list servers:
1. which are NOT owned or controlled by the code originator "code.coolthings.net2012"
2: which host relish code artifacts from that origin 


relish_code_repositories.txt
-----------------------------
codewarehaus.net
relishchunks.org
-----------------------------

Each line of the text file contains the host name or IP address of a
relish source-code repository which serves code artifacts from multiple origins.


relish_code_staging_servers.txt
-----------------------------
code.coolthings.net2012 staging.coolcode.com:8080
cs.ubc.ca1989 dev1.cs.ubc.ca
-----------------------------

Each line of the text file begins with a relish code origin identifier and is followed
by a space-separated list of server hostnames or IP addresses.
By example the first line of this file is used to list servers:
1. which are owned or controlled by the code originator "code.coolthings.net2012"
2: which host versions of relish code artifacts from that origin which are being
   tested prior to general public release
3. whose host name is not "code.coolthings.net"


Order of network search for relish code
---------------------------------------
When a relish installation is trying to find on the network a particular version or the current version of a 
software artifact, it searches servers in a particular order, and stops searching after it has found the
artifact it is looking for.

The order searched is as follows:
1. servers in relish_code_staging_servers.txt. If the staging servers file contains the origin of the desired
software artifact, then the list of staging servers for that origin is tried in left to right order.
If a server entry in the file does not have a :port suffix, then port 80 then port 8421 are tried for each
server.
2. The standard server for the origin, e.g. http://code.coolthings.net followed by http://code.coolthings.net:8421
3. servers in relish_code_origins.txt. If the origins servers file contains the origin of the desired
software artifact, then the list of servers for that origin is tried in left to right order.
If a server entry in the file does not have a :port suffix, then port 80 then port 8421 are tried for each
server.
4. servers in relish_code_replicas.txt. If the replicas servers file contains the origin of the desired
software artifact, then the list of servers for that origin is tried in left to right order.
If a server entry in the file does not have a :port suffix, then port 80 then port 8421 are tried for each
server.
5. servers in relish_code_repositories.txt. Repository servers are tried in file line order top to bottom.
If a server entry in the file does not have a :port suffix, then port 80 then port 8421 are tried for each
server.
6. (Future) Other servers returned by a google search for the artifact's standard artifact name info in the 
artifact metadata.txt file, in random order.
7. http://shared.relish.pl





