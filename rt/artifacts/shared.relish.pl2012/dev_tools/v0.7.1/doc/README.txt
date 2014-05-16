
////////////////////
dev_tools README.txt
////////////////////

How to use the playground (simple program trial environment) and explorer (data explorer) tools.


Note: All terminal command-line commands mentioned below need to be run while the working directory
is in a relish directory tree in the filesystem.


~~~~~~~~~~
PLAYGROUND
~~~~~~~~~~

The playground is an environment for trying out simple programs to learn relish, or it
can also be used to add a simple debugging program onto a bigger application to help
debug the bigger application.


Use of playground on independent simple programs (learning relish)
==================================================================

relish -web 8080 shared.relish.pl2012/dev_tools playground

In a web browser, browse to localhost:8080


Try out one of the existing example programs:
---------------------------------------------

Choose a program such as "example2_loops" 
from the drop-down menu [ -- Choose existing trial program --  v ]

Press the [ Run ] button. The program output should appear in the Output Console.

You can modify the existing example programs to try out variations.
If you do, and you want the original programs back, you need to stop your
playground process, then

rmdb shared.relish.pl2012/dev_tools     

then restart the playground. CAUTION that doing that this use of rmdb will also 
remove your own simple programs you've created as shown below.


Write and test your own simple program:
---------------------------------------

Enter a program name such as   test_multi_assignment   
You can also enter a description of what your program does.

Enter relish statements into the "Statements of Main Method" text area.
You can also enter:
-method declarations into the "Methods" area,
-your own data type declarations into the "Data Types and Relations" area,
-constant declarations into the "Constants" area, and 
-imports of standard library packages or packages from other artifacts in the
"Imports" area.

Press the [ Run ] button. If you included any print statements, the output of your program
should appear in the Output Console.

If relish detects a compilation error, the output will describe the location of the error
e.g. Data Types and Relations:line#:character-column# followed by a description of the nature
of the error. Note that in current relish, compilation error reports are sometimes not clear at 
identifying the reason for the error. If this is the case, just check at the location in the code, 
and refer to relish reference manuals and examples to diagnose the difference from legal relish 
code.

In the case of an error detected at runtime, the output MAY describe the location of the error,
but in any case will describe the nature of the error.



~~~~~~~~
EXPLORER
~~~~~~~~

The data explorer allows you to visualize the data objects in a relish program.
The explorer focusses on one data object, which it displays, including its
type and primitive attributes, at the top left of the screen. The explorer
also shows the network of objects associated with the focus object. 
That is, the objects which are object-valued attributes of the focus object,
or are in a relation with the focus object. The explorer shows up to 2 levels
of associated objects along with the focus objects.

You can click on an associated object in the explorer to refocus the explorer on that
object (that is, to traverse the associated-object network.)

Invoking the explorer
=====================

First, start your target program, which contains the data, but use the -explore
option to make sure the program serves the data reflection web service, which
is used by the data explorer tool.

relish -explore 8089 -web 80 myorigin.org2014/mywebapp

Now start the data explorer tool:

relish -web 8080 shared.relish.pl2012/dev_tools explorer

In a web browser, browse to localhost:8080

Enter the url of the reflection web service of your target program:

Target Program URL: [ localhost:8089 ] [connect to program]

Finding objects to explore
==========================

Upon connecting to the program, You will see listed the data types of any data objects
that have been created or loaded in the program.

Find Persistent Objects by Data Type
------------------------------------
If you select a datatype and press [find] you will see on the right side of the screen
a green coloured list of the persistent object instances which are of that type or a subtype.

Clicking on one of these objects in the right-side list will make it the focus object of the 
data explorer, so the explorer will display the object and its
associated object network.

Find Persistent or Non-Persistent Objects by Name
-------------------------------------------------

Some persistent data objects in the target program were persisted by being dubbed
with a name. dub myObject "Official Name"

Pressing the [find named objects] button in the data explorer will display the list
of named objects in the green list on the right.

Even if a data object in the target program is not persistent, it can be found by name
in the data explorer, so long as you labelled the object.

To label a non persistent object for finding in the data explorer, do the following
in your target program: 

import
   reflect

label myNonPersistentObject "Some Label"

Setting up breakpoints (pause points) in your target program
============================================================

After labelling some objects of debugging interest, you may want to ensure
that your target program pauses at a particular point in its execution of instructions,
so that you can view the data state as of that point.

import
   reflect

label myNonPersistentObject "Some Label"
...
doThingsTo mynonPersistentObject
...
reflect.pause


Combining "label" and "pause"
-----------------------------

Instead of the above code, you can just put in your target program:

import
   reflect

...
doThingsTo mynonPersistentObject
...
explore myNonPersistentObject


Which will label the object "**Object in Focus**" then pause the target program.
The data explorer will in this case immediately display the **Object in Focus**
when the data explorer next connects to that target program.

Note that you can place multiple 

   label obj name
   unlabel name
   reflect.pause
   explore obj

statements throughout a target program you are debugging, 
to explore at multiple breakpoints. The [resume] button on the data explorer screen
will resume from the current pause/explore breakpoint and run the target program
until either it ends or encounters the next pause/explore statement.



~~~~~~~~~~~~~~~~~~~~~~~~
EXPLORER FROM PLAYGROUND
~~~~~~~~~~~~~~~~~~~~~~~~

When running the playground (simple program trial environment), you can press the 
[explore program data] button to display in the explorer's object tree view those 
data objects you have created in your playground program.

Here's what this does:
1. Adds an import of the reflect package (from relish standard library) to your program.

2. Adds an "explore nil" statement to the end of your main program statements. This will
cause your program to pause there when that statement is reached, rather than terminating.

3. Runs your program with relish -explore 8089, so that your program serves the data
reflection web service api.

4. Switches from the playground to the explorer (Data Explorer) screen, with the explorer
reading the data from your playground program.

Ways of finding and displaying your playground program data objects in the explorer:
-----------------------------------------------------------------------------------
1. If your playground program has defined any of its own datatypes, and used them in the
main program, you will see these types listed in the explorer. Only if you have persisted
some instances of these datatypes in your program (e.g. dub myObject "Some name") 
will you be able to [find] them in the explorer by searching by datatype though.
The DataType:[  ] Where:[  ] [Find] function in the explorer works on persisted data objects only.

2. If you want to explore the value of a variable in your playground main program,
you can:
a) press the [ < return to program trial environment] button in the explorer.
b) replace the explore nil statement with explore <variable>
e.g.
myList = ["Albert" "Belulah" "Charlie" "Dora"]
explore myList
c) press "explore program data" again.

3. You also have the option to label your data for exploration in the explorer.
a) In your playground main program, you could say:
myPeople = ["Albert" "Belulah" "Charlie" "Dora"]
myCities = ["Vancouver" "Seattle" "San Francisco"]
label myPeople "People"
label myCities "Cities"
explore nil
b) Then press "explore program data" again.
c) Now use the [find named objects] button in the explorer and click on "People" or "Cities"
to see these data objects in the explorer.



~~~~~~~~~~~~~~~~~~
ADVANCED DEBUGGING
~~~~~~~~~~~~~~~~~~

Use of playground to debug a full-scale relish application
==========================================================

Imagine you are developing a full scale, multi-package relish artifact, for example a
web application with persistent data. And you want to debug it.
Currently, there is no full interactive debuggger tool for relish, but you
can approximate part of that. You can at least write a simple test program in the
playground that can import the packages (and a copy of the persistent data)
from your application, and use this to explore the state of your program's data model.

You can use the following procedure: 

relish -web 8080 shared.relish.pl2012/dev_tools playground

In a web browser, browse to localhost:8080

Enter a program name, such as "test_mywebapp_data"

At the bottom left of the trial environment screen, enter the full name
of the artifact you wish to debug.
e.g. myorigin.org2014/mywebapp
Press the [Copy DB] button. 

This copies the local database of your full-scale
web application to be the database of the current simple program in the
playground. Your simple program can summon or query for those data objects,
as long as you make sure to add the packages that define their data types
to the Imports area of your simple program.

You can then use the techniques described in the EXPLORER FROM PLAYGROUND section above
to explore the object-network of the web application's data model, or at least
to explore the snapshot-copy of the web application's data model.

The database is copied into the playground rather than referred to for safety,
to prevent unwanted experimental data modifications you make in your playground test
program from corrupting the state of the real database of the web application. 





