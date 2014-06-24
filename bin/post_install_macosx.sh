#!/bin/bash

# post_install_macosx.sh
#
# Completes the installation of relish binary distribution on macosx.
#
# Must be run as root
#
# Make relish command available in default PATH 
# and make sure relish directory tree is writeable and commands are executable.
#
mkdir -p /usr/local/bin   # harmless if directory already exists
chmod +x /opt/relish/pl/bin/*   
ln -sf /opt/relish/pl/bin/relish /usr/local/bin/relish
ln -sf /opt/relish/pl/bin/rmdb /usr/local/bin/rmdb   
ln -sf /opt/relish/pl/bin/cleanr /usr/local/bin/cleanr   
chmod go+w /opt/relish      
chmod -R go+w /opt/relish/keys     

osascript <<END
tell app "Terminal" to do script "cd /opt/relish; relish -web 8080 shared.relish.pl2012/dev_tools playground" 
END

