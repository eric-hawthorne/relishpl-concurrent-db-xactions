#!/bin/bash

# install.sh
#
# Completes the installation of relish, after the tarball has been extracted.
#
# must be run as root, so
#
# For installation of a relish binary distribution:
#
# cd /opt/relish or cd ~/relish
# sudo ./install.sh
#
# Or for installation of a relish source distribution:
#
# cd /opt/devel/relish
# sudo ./install.sh
#
# Make relish command available in default PATH 
# and make sure relish directory tree is writeable and commands are executable.
#
mkdir -p /usr/local/bin   # harmless if directory already exists
if [ -d /opt/devel/relish ] # It is a source distribution  
then   
   chmod +x /opt/devel/relish/bin/*
   ln -s /opt/devel/relish/bin/relish /usr/local/bin/relish
   ln -s /opt/devel/relish/bin/rmdb /usr/local/bin/rmdb   
   ln -s /opt/devel/relish/bin/clean /usr/local/bin/clean    
   ln -s /opt/devel/relish/bin/makedist /usr/local/bin/makedist     
   chmod -R go+w /opt/devel/relish      
elif [ -d ~/relish ] # It is a home directory binary distribution
then
   chmod +x ~/relish/pl/bin/*   
   ln -s ~/relish/pl/bin/relish /usr/local/bin/relish
   ln -s ~/relish/pl/bin/rmdb /usr/local/bin/rmdb   
   ln -s ~/relish/pl/bin/clean /usr/local/bin/clean   
   chmod go+w ~/relish      
   chmod -R go+w ~/relish/keys   
else
   chmod +x /opt/relish/pl/bin/*   
   ln -s /opt/relish/pl/bin/relish /usr/local/bin/relish
   ln -s /opt/relish/pl/bin/rmdb /usr/local/bin/rmdb   
   ln -s /opt/relish/pl/bin/clean /usr/local/bin/clean   
   chmod go+w /opt/relish      
   chmod -R go+w /opt/relish/keys      
fi 

