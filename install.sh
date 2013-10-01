#!/bin/bash

# install.sh
#
# Completes the installation of relish, after the tarball has been extracted.
#
# must be run as root, so
#
# For installation of a relish binary distribution:
#
# cd /opt/relish
# sudo ./install.sh
#
# Or for installation of a relish source distribution:
#
# cd /opt/relish
# sudo ./install.sh
#
# Make relish command available in default PATH 
#
if [ -d /opt/devel/relish ] # It is a source distribution  
   ln -s /opt/devel/relish/bin/relish /usr/local/bin/relish
   ln -s /opt/devel/relish/bin/rmdb /usr/local/bin/rmdb   
   ln -s /opt/devel/relish/bin/clean /usr/local/bin/clean    
   ln -s /opt/devel/relish/bin/makedist /usr/local/bin/makedist        
else
   ln -s /opt/relish/pl/bin/relish /usr/local/bin/relish
   ln -s /opt/relish/pl/bin/rmdb /usr/local/bin/rmdb   
   ln -s /opt/relish/pl/bin/clean /usr/local/bin/clean      
fi 

