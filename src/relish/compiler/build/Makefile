# Copyright 2009 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE_GO file.

include $(GOROOT)/src/Make.inc

TARG=relish/compiler/build
GOFILES=\
	build.go\
	dir.go\
	path.go\
	syslist.go\

CLEANFILES+=syslist.go pkgtest/_obj cmdtest/_obj cgotest/_obj

include $(GOROOT)/src/Make.pkg

syslist.go: $(GOROOT)/src/Make.inc Makefile
	echo '// Generated automatically by make.' >$@
	echo 'package build' >>$@
	echo 'const goosList = "$(GOOS_LIST)"' >>$@
	echo 'const goarchList = "$(GOARCH_LIST)"' >>$@
