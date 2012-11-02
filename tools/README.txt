Relish.tmLanguage (and its companion file, Relish.tmLanguage.ForBundleEditor) is a 
syntax highlighter configuration file for relish programming language source code files (.rel files)

This configuration file works in Sublime Text 2 and TextMate (a text editor for Macs) to enable
relish source code syntax colouring in these editors.

------------------------------
To install in Sublime Text 2, place the file as follows

Linux: ~/.config/sublime-text-2/Packages/Relish/Relish.tmLanguage
OS X: ~/Library/Application Support/Sublime Text 2/Packages/Relish/Relish.tmLanguage
Windows: %APPDATA%\Sublime Text 2/Packages/Relish/Relish.tmLanguage

and restart Sublime Text 2

Then, in Sublime Text 2
1. Open a .rel file
2. Go to the View menu
      Syntax 
        Open all with current extension as...

        Relish

3. Shield your eyes from the dazzling colours.

----------------------------
To install in TextMate, you should be able to install it like this:
Option 1: 
---------
a. In TextMate, open the file Relish.tmLanguage.ForBundleEditor
b. Select all of the contents of the file and press command-C or Edit/Copy
c. Select Bundles -> Bundle Editor -> Edit Languages  from TextMate menu 
        
d. from the + dropdown at bottom left of screen, choose "New Language"
Name the language "Relish"

e. In the text area on the right which contains example javascript for a language definition,
select all of the existing javascript code, and press command-V or Edit/Paste to replace the
example javascript with the contents of Relish.tmLanguage.ForBundleEditor

f. Exist the bundle editor
g. Quit TextMate and restart it. 
h. Open a .rel file and you should see syntax highlighting (colouring).

If for some reason that method doesn't work, you can try the following.

Option 2:
---------
a. Select Bundles -> Bundle Editor -> Edit Languages  from TextMate menu 
b. from the + dropdown at bottom left of screen, choose "New Language"
Name the language "Relish"
c. Exit the Bundle Editor window. This should have saved a tmBundle file to the location
/Users/you/Library/Application support/TextMate Bundles/Your Name's Bundle

d.Open a Finder window and navigate to the bundle file. 
e. Select the bundle file in Finder.
f. User Finder's "gear" dropdown menu (top centre of window) and select
"Show Package Contents"
This should open a new Finder window displaying an info.plist file and a Relish.tmLanguage file
g. Drag the Relish.tmLanguage file to the TextMate icon in the dock to open the file in TextMate.
You will see the xml form of language grammar specification.
h. Replace the example xml in the file with the full contents of the Relish.tmLanguage file
 that resides in the tools directory with this README.txt file. 
i. Save the Relish.tmLanguage file (the one from the bundle contents) in TextMate.
j. Quit TextMate and restart it. 
k. Open a .rel file and you should see syntax highlighting (colouring). 
