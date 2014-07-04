native_methods/extensions/protocols/README.txt

This directory should be the parent directory of Go packages which implement specific data-communication protocols.
For each protocol, there should be two Go packages: e.g. 1) modbus which contains the actual Go implementation of
the protocol, and 2) modbus_methods, which contains the rt.CreateMethod calls to map methods into relish methods, and the 
actual wrapper methods (in Go) which change the parameter and return value data types of the protocol functions to
be []RObject.