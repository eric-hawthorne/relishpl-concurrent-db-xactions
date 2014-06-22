@ECHO OFF 

REM
REM Creates a new release of the relish binary distribution, by copying appropriate files
REM out of your local copy of the source distribution.
REM
REM First, when you are on a given platform, e.g. a Mac, 
REM compile the relish compiler-interpreter program by using
REM go install relish/relish    (assumes your GOPATH env var is set to your relish source distribution directory e.g. to ~/devel/relish)
REM Now you have a relish executable for Mac as ~/devel/relish/bin/relish
REM Now 
REM makedist.bat 1.0.2 to produce the binary distribution
REM directory tree %USERPROFILE%\Desktop\devel\relish_dist\relish_1.0.2.windows_amd64
REM
REM which produces
REM  %USERPROFILE%\Desktop\devel\relish_dist\relish_1.0.2.windows_amd64.tar.gz
REM
REM but also leaves around the un-tarred distribution directory.
REM

SET SRCDIR=%USERPROFILE%\Desktop\devel\relish
ECHO "SRCDIR is"
ECHO %SRCDIR%   
SET DISTDIR=%SRCDIR%_dist
ECHO "DISTDIR is"
ECHO %DISTDIR%

REM Create release directory if does not exist.
REM
MKDIR %DISTDIR%

REM Read release version and architecture from the command line. e.g. 1.0.2
SET RELEASE_VERSION=%1
REM Default the target architecture because we are making distribution on Windows
SET RELEASE_ARCHITECTURE=windows_amd64

SET RELEASEDIR=%DISTDIR%\relish_%RELEASE_VERSION%.%RELEASE_ARCHITECTURE%
ECHO "RELEASEDIR is"
ECHO %RELEASEDIR%

MKDIR %RELEASEDIR%\pl\bin

REM preserve executable mode, timestamps
XCOPY %SRCDIR%\bin\relish.exe %RELEASEDIR%\pl\bin
REM XCOPY %SRCDIR%\bin\cleanr.bat %RELEASEDIR%\pl\bin
REM XCOPY %SRCDIR%\bin/rmdb.bat %RELEASEDIR%\pl\bin

XCOPY %SRCDIR%\LICENSE* %RELEASEDIR%\pl
XCOPY %SRCDIR%\COPYING* %RELEASEDIR%\pl
XCOPY %SRCDIR%\README_WINDOWS.rtf %RELEASEDIR%\pl

MKDIR %RELEASEDIR%\tools
XCOPY %SRCDIR%\tools\README_WINDOWS.txt %RELEASEDIR%\tools
XCOPY %SRCDIR%\tools\Relish.tmLanguage %RELEASEDIR%\tools

REM XCOPY %SRCDIR%\install.sh %RELEASEDIR%
REM chmod +x $RELEASEDIR/install.sh

MKDIR %RELEASEDIR%\keys\private
MKDIR %RELEASEDIR%\keys\public
XCOPY %SRCDIR%\rt\keys\private\README.txt %RELEASEDIR%\keys\private
XCOPY %SRCDIR%\rt\keys\public\origin__shared.relish.pl2012__public_key.pem %RELEASEDIR%\keys\public

REM cd %RELEASEDIR%
REM SET TARBALL=relish_%RELEASE_VERSION%.%RELEASE_ARCHITECTURE%.tar.gz
REM tar -czvf ../$TARBALL pl install.sh keys
REM echo created %DISTDIR%\%TARBALL%
