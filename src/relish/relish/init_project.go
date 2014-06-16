// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

/*
Creates new relish project directory tree templates.
There is a template for an application project, 

Usage: cd ~/relish
       relish -init willywonka.com2014/chocolate_factory_controller

and a template for a web-application project.

Usage: cd ~/relish
       relish -init willywonka.com2014/online_chocolate_shop webapp


*/
package main

import (
        "fmt"
        "util/gos"
        "strings"
        "time"
        "encoding/base64"
)


const APP_METADATA_FILE = `relish artifact metadata: %s
origin: %s
artifact: %s
current version: 0.1.0
release date: 2099/09/19
tags: application,ADD YOUR TAGS HERE COMMA SEPARATED

Please put a concise description of your software artifact (application or library)
here. A paragraph would do nicely. What is it for? In what context is it to be used?
`


const WEB_APP_METADATA_FILE = `relish artifact metadata: %s
origin: %s
artifact: %s
current version: 0.1.0
release date: 2099/09/19
tags: web application,ADD YOUR TAGS HERE COMMA SEPARATED

Please put a concise description of your web application software artifact 
here. A paragraph would do nicely. What is it for? In what context is it to be used?
`


/*
relish artifact metadata: 2013/09/25
origin: relish.pl2012
artifact: relish_website
current version: 0.1.0
release date: 2012/09/29
tags: web application,relish

The relish.pl website.
*/




const APP_MAIN_PROGRAM_FILE = `origin   %s
artifact %s
package  main

"""
 main.rel
 
 The main program (program entry point) for the %s
 software artifact.
"""


main
"""
 The main method.
 Initializes this and that and the other then prints a start confirmation message on the
 standard output stream. 
"""
   print "\nHello from the %s/%s program.\n"
`


const WEB_APP_MAIN_PROGRAM_FILE = `origin   %s
artifact %s
package  main

"""
 main.rel
 
 The main program (program entry point) for the %s
 web application.
"""


main
"""
 The main method.
 Initializes this and that and the other then prints a start confirmation message on the
 standard output stream. 
"""
   print "\n%s/%s web application started.\n"
`


const INDEX_HTML_FILE = `<html>
  <head>
  <title>Current Local Time</title>

  <link rel="stylesheet" type="text/css" id="stylesheet"
        href="/styles/default.css" />  
  
  </head>
  <body>
      <center> 
        <h1>Current Local Time</h1>
        <br/>
        <br/>
        <p>
        It is always handy to know that
        the current time is <span class="time">{{.time}}</span> &nbsp; &nbsp; <img src="img/relish_logo_small.png"/> 
        </p>
        <br/>
        <p>
        <a href="example.html">An example static web page</a>
        
        
        </p>
        <br/>
        <p>
        The fruit on special is {{.fruit}}s.
        </p>                
        <br/>
        <p>
        <a href="visitCount">Number of times this web site has been visited</a>
        </p>      
        <br/>
        <p>
        <form action="guessingGameAction" method="POST">
           Guess a number between 1 and 10 : &nbsp;
           <input type="text" name="guess" size="4" /> &nbsp;
           <input type="submit" value="submit your guess" />           
        </form>
        </p>          
     </center>
  </body>
</html>
`

const DIALOG_FILE = `origin   %s
artifact %s
package  web

"""
 dialog.rel
 
 Web application dialog handling methods.
 Typically divided into:
 -action methods which accept and process html form input data or 
  AJAX requests, and
 -dynamic-page methods, which lookup persistent data from the database
 and insert the data into a dynamic web page template to create and serve a web page.
"""


import
   datetime


FRUITS = ["apple" "pear" "banana" "orange"]


VisitLog
"""
 An object that remembers the number of visits to the main page of the web app.
 Also, on a completely different note, stores a secret number that the web app user
 must guess, and the number of guesses in total so far.
 This whole thing is just to show the concept of storing data from the web app in
 relish local persistent storage.
"""
   numVisits Int
   secretNumber Int
   numGuesses Int



index > String Map
"""
 Handles a request for the root (i.e. empty path) url on the server and port, or a request for /index.html
"""
   t = now "Local"
   ts = format t "3:04PM"
   
   if exists "Visit Log"
      visitLog = summon "Visit Log"
      visitLog.numVisits = plus visitLog.numVisits 1
   else
      visitLog = VisitLog
      visitLog.numVisits = 1
      visitLog.secretNumber = 5
      visitLog.numGuesses = 0
      dub visitLog "Visit Log"
   
   => "index.html"
      {
         "time"  => ts
         "fruit" => FRUITS[mod visitLog.numVisits (len FRUITS)]
      }String > String


visitCount > String Any
"""
 Prints the number of visits to the main page of the web app.
"""
   if not exists "Visit Log"
      => "REDIRECT"
         "/"
         
   visitLog = summon "Visit Log"
   => "visit_count.html"
      visitLog.numVisits


guessingGameAction guess Int > String String
"""
 Accepts a guess for a number between 1 and 10.
 Reports whether it is the right number or not.
 "right" is defined as one more than the last entered number, except wrapping around at 10.
"""
   if not exists "Visit Log"
      => "REDIRECT"
         "/"
         
   visitLog = summon "Visit Log"
   visitLog.numGuesses = plus visitLog.numGuesses 1   

   currentSecret = visitLog.secretNumber
   offset = 1
   if eq guess visitLog.secretNumber
      => "REDIRECT"
         "gotIt"
             
   => "REDIRECT"
      "tryAgain"
           
      
tryAgain > String Int
"""
 Displays that the user guessed wrong. Show the number anyway. Invite to try again.
""" 
   visitLog = summon "Visit Log"  
   oldNumber = visitLog.secretNumber
   visitLog.secretNumber = plus 
                              mod 
                                 plus visitLog.numVisits visitLog.numGuesses
                                 10
                              1
   => "try_again.html"
      oldNumber


gotIt > String Int
"""
 Displays that the user guessed correctly. Show the number. Invite to play again.
""" 
   visitLog = summon "Visit Log"  
   oldNumber = visitLog.secretNumber
   visitLog.secretNumber = plus 
                              mod 
                                 plus visitLog.numVisits visitLog.numGuesses
                                 10
                              1
   => "got_it.html"
      oldNumber

   
default > String String
"""
 Handles all url paths on this server and port which are not otherwise handled.
"""
   => "HTML"
      "<html><body><center><br/><br/><h3>Oops. Nothing here.</h3></center></body></html>"
`

const CSS_FILE = `
body
{
   font-family: sans-serif;
}

span.time
{
   font-size: 30px;
   font-weight: bold;
   color: green;
}
`


const STATIC_HTML_FILE = `<html>
  <head>
  <title>Example Static Html File</title>
  </head>
  <body>
     <h1>Example Static Html File</h1>
     <p>
     No dynamic template content here.
     </p>
     <p>
     Just a plain old html file. In relish, it goes in the web/static directory,
     or in a subdirectory of web/static. But the url to reach the html file that is
     called ~/relish/artifacts/you.com2014/v1.0.0/src/web/static/products/example1.html would be 
     http://you.com/products/example1.html
     </p>
  </body>
</html>
`


const VISIT_COUNT_HTML_FILE = `<html>
  <head>
  <title>Visit Count</title>
  </head>
  <body>
     <center>
     <h1>Visit Count</h1>
     <p>
     Wow. This web site has been visited {{.}} times.
     </p>
     <p>
         <input type="button" value="&lt; back" onclick="history.back()"/>
     </p>
     </center>
  </body>
</html>
`


const GOT_IT_HTML_FILE = `<html>
  <head>
  <title>You Got It!</title>
  </head>
  <body>
     <center>
     <h1>You Got It!</h1>
     <p>
     The number is {{.}}
     </p>
     <p>
         I know, this is the worst game ever. It's just a dummy web app demo.
      </p>
         <p>
             Press the <input type="button" value="&lt; back" onclick="history.back()"/> button to play again.
         </p>         
     </p>
     </center>
  </body>
</html>
`


const TRY_AGAIN_HTML_FILE = `<html>
  <head>
  <title>Try Again</title>
  </head>
  <body>
     <center>
     <h1>Sorry. That's not the number.</h1>
     <p>
     The number was {{.}}
     </p>   
     <p>
         I know, this is the worst game ever. It's just a dummy web app demo.
     </p>
     <p>
         Press the <input type="button" value="&lt; back" onclick="history.back()"/> button to try again.
     </p>
     </center>     
  </body>
</html>
`



const RELISH_ICON_PNG = `iVBORw0KGgoAAAANSUhEUgAAAAsAAAAQCAYAAADAvYV-AAAA_UlEQVR42mNgwAL8CutaN9zK_chAFGhg-L_-P8N_ohTCFG_4x4lF
Qw2DJsMNIASZNguhGLvpfyCCIg0yYIUwiF3xfyTFSJD9D7ozQCb9QZgg8F30P8Mihv-KHzXBtuidcP8E0rj5p-x_BuUmk38o7qtk
-I1iWD2yk6BW2j7zgihOZFgwbdrFlNnLN-TBnYhsWMCBrPfoEoTDGdkUEJzJ4Igir5Di9JypngViynk0k_9DwhyEs9pnHGeQr7cE
c3Z99IdIQIFJcxhcIVsND8KDIAZTLZoboSbnHykEy--9O_EUAy4PRe-r3gOSS7zrA5RjJJyoKlctso-_HglXCACGvcrTaHgvvAAA
AABJRU5ErkJggg==`

const RELISH_LOGO_PNG = `iVBORw0KGgoAAAANSUhEUgAAAF0AAACICAYAAACFgudfAAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAADdwAAA3cBr5JgKAAAABl0
RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAACAASURBVHic7Z13fFRF9__fc3c3vWx6IEDoifRAaCG00ARsIAgiggiR
EqqAgEi3gPqoCKIiKqIiVaPgAygiiiAoIE1CCQECIfRASLIpuzu_PxZCyu5mdxMS-P6eD68L7L1Tzv3cuWdmzpw5V0gp-b-A58Yv
GxIcoh0_f3KfxhUtS0lQKlqAssCIMWtb_unz6adOwVcbrVi3_aOKlqckPPCkx8XFB-33-_qPE-wWnsEJBHV5e_iHHx5uXtFyWcMD
TXpcXLzTcd9NiX-L79UAGvfr6Lx_pHrMf7cvWnTgvr23-1YwW3De--_jv6qXedz5LZQ8AHLrzHULDDuzrcIEKwEPLOn9Xpq_bbPz
u9WNGPPPSUyDAqlk4d1uXvuly_cMqCj5rOGBJH3gxEULt3i80zEXXeEL4u4DyHbZT2jnb1YsXnxQW87ilYgHjvTnx3829FfvRWNv
yivFrkkMhX5nV1mkCml2YF95yWYrHijSR4xZ23qXzyefXJAnLaSQRX4Z0UTOqLl8zbZX7710tuOBIT0uLr7Sfr-VO46zW1hKIwuo
lzvQa85RqfOHLy9ZcqjuPRXQDjwQpN8eGp74W8SrrKWTmJ9d63zXidB2W_-6J8I5gAeC9HPee04WHBpagqB4S78DQ73Z3t_8-MO3
ZSqYg7jvSe875bXftji_V81ohdA7MArLdiSjcgufjvN7fbzs725lKZ8juK9Jf3rSO0u2uL_TLpdsG3NYfzDZbn9SrfO6HxYvPuBS
eukcx31L-rMTPhy-zev9kbfkddszmelIiyIn9B2nSg2P7i6FaKXGfUn6kLHL2-7ULv3wsjxrZ04bzNRCj0vrWY0_--bXcQ4JVwa4
70iPi4sP2ef75a-nOWBxaGgJlkYvRZHnlEiVzsvf-eCDgyF2C1gGuK9Ij4uLdzngt-rEYbHN6tDQEuxZkMkKWKFUbb2zQmar9xXp
idptibtUq90czW_vIphoPDPoy_hNSx2tz1HcN6Q_MWXmzm1OH5XqdbdVvdyBQXWNwE7vxH788T8tSlOvvbgvSO836a2lW93ej9KT
V6pyHFnt1XlupWrHjb-W56JHhZM-cMKSUdu8FsVmcrP0hTm4xq6v87pbUHhiuS16VCjpz49d3n6nduniq_JcmZRndNCzwSiy8Wg7
r_2yFTsHlokgJaDCSI-Li6-6z_erX85w0O6hoSWUxpkk1-UQVTt9s7w8Fj0qhPS4uHiXg35rjx0Svzg0NLSE0rrw6EKWqKo033fP
Z6sVQnqSdvupnaqVDg8NLaH0jlMS9warapSJMFZQLqQLIVyFECqAXlNm7drqtKTyvajH3iGjWQh5zzlRl3WBIlxUphP9qEMd_KmG
F9XZSSWy0Wt-dDYaIw3BxmwDXMF05ADeZVP3g-IhWCaki8aiCt2JpT5RrKY5jc3TqCe3-MnzwG7g0u0jHdNDcKC9OTp6KW-UinTR
WNSgN5P4iF60ppJDhVQB-hT4nQbsAFKBvbev_x-DQ6SLtqIuPXiR5fQmmwAeKkOJfIDamFq_nUYBpey15T2B3VKKF0Qs7zALf0L4
FPDgEmc5wmVOcoZErpKOBhVqVDijojYhVKIR_tQnkqp4llDBN8ABwM6xTRPlYfxuPY3acBi9qrhPjK0QWFnzKyPYTLoQwpUZLGYa
z_ItGk6RzVmm8l8WSykNNuTXNH7i4VkHG26eTmWgOtABuLNwdh1YAuiwi3CBoK3TQNKeP8Ucrym8uusNQqJewqCyY8WpACSyzCZr
lmBTdyUeEc34gm3M4Xn80DAOcOdd-aNcaAvhAFLKvGa-jx9GhanD3AVMA17DRPabgAFwsl14NU509x7N6dF_cjhoF9dcL_JK5Cuk
_PkaitHL9oLKGSWSLqJECHF8zSBaIQAvYD9X2czCUtWsArSAHtPQ0dW-7O5oeazaBHaN_IZznon556-5XmRGs7lc3P06wuheKhHv
FaySLoRQ6MuHdCes0IULnJGH5aV7KpkVBIhQukfEsnHQQm44Xy12_aprKtMjXiM38e0cRdr5NMsB1lt6H55kJI8WOy8oU5uJPahB
BK26dGf9Y2-Tq7LsmnHVNZW4qvOcSX4XIW3XWeXRkVonPYa2mPMQ0ZTVHNI-NFa6Uu3pmmxo_ZFNU_4rrhcYFjgX5dzbIG0dM9x7
y4j1Grwxb_ypQxURKmreC4HM4fYIRcrYLH6ru96uvFdcLzDUfwHqlPm2vaA2PxzHYZ30HLLMnq-FE-3pei8EKgo1Grq4D-Ns3F_i
UPAfDpVxxS2F533fRZUyD7A-IlTKQXNaJ93SGpoCtCT6HshTCO540z14BH-NjifZ60SpyrrilsJQnyWoL5RAvNSUqh5bYJ30ZM5Y
vFaXaCFEmdvE78CfqnStN4Sfhi3jhovjM8yCuOx2nmHeS9GkzMUS8UKqK7gjPcCRott68hFDKCMYW_YiQQ0a0y66F9_3WUSOypIA
juGSezKx3p_idGEO5ogX5WC_sU76z2xiO-b3mqiAaHoLIcq0u29i7EL9xyL5ttP7GIVNk127kepxhhe8luN8YVbxixXdkUop8zjD
bxYTPEEk_RlUFoIIBB2MA9EO8mJjxKdlUaRVpHgkMdzza1wuzC5ypaJbOsCPrOaqudUHwB1BJ54trRAanHlUjOfWyPNsr2XfkLA0
OOd5khEe3-CaOjP_nFLhOh2QG-VWfuB3iwk600Z0EA7vxfcUfvRzfZmEMdvYF7zd0WIcxlmv44xwW3uX-IpWL_nYxjcWRuxQHWdi
eNqRyoNFTZ72f5lf4j7jpM9BR4owQQesAiaBhXfSKs54JzDK9VtcU18BKnDI-NZbe2rl__ia5fwXy27FYbS0t-K6tKBvnfGsHvo6
qe72Ov_fxiFgITAd-Bdoil2m4YJI0h5hotNPOKcMSnGsBNth9l2Ki4vXNmpU4991G_ceTj2t6SilzBBTxbf0oZnZUqoSJoRwl1Jm
2lJpC_kYLVvF8EmXqWSrLL1CVrAF2ImphXtgWsjOAp6yv6g7aHmpixxxcep3urSIfo6XYhvMkp7isX9VoGuSc6UecyIfOjX3-tKl
6gFs5XP2MZlmFHc7a4wf0bQGtpZUYTfDC4T2qMqSlhMxODIk1APbMK0u3dnkmAX0s3Q31lEloxbDz05P6pnbKyYiQuvgK2cfiqmX
uLh4dYLbz1303kcxKulk1BmvqTFg4trVs3_9RhwWu8yW4gw4U_I-T8GBx4KfTdN53UJldFB3rrhd3x1ITEt_Te0rxlXvzrDEl3Vf
HP_5qVfqD6lVXoSDGdIvuSZ8cJI9il65-9pnemzBpXuv9k96dTJv5BKAseROedmyEcdHjYr2rbG65fx58WvyIi63s0_ao0ACFLJJ
6YAhthchEMSkPCkX7fnu82WPvN4_plmNtfYJUXoUIyrJfccQicSoFNa1RtUNAv23mn-BE8nmOHttrXTOnN7Tkn6-qO29dNK6sbve
ltrsgJIzSWAlFHqfcoB2YNbmbwa1bjTk7SOrD58bfnzAsIVdI-jMF6KRaG-r3NYwY8ZKp5ET1zwy5PWlH7333r6h1tIWInHghMXj
j2i3aQCMihmbR46FUo5zVqbIZHuE_Oij2Cyg74QJX1Udt--D73b3jG_2U41VZoMqAGCuPboCPUuuyyvHl-fOTkqPSer69BMLInvQ
mE_xue1z0JD3RA3RXZ6WF22VfezYbyOuVEnulx6S0uFWwOXaN4MuaK_GnVVd9j-LR6YP_XI6Df_xeN13L-zuOjJ2cOuvi-YvRPpp
j11z824zK80thVnqqM7b3sqL4t13B54DIqdNM0aHNW259oeuHwWf8U4onOgmJm-vguvMWcBw62UrUkXP5GeNQ9PGvfnE8xFb3w97
eQHtaVbIzlWXJuSwXtQW_WVi4d0JY8ZsqKTzvPHEtepJj6QHpza8GXQh8HpwsvPFBUnoXG-ZrTPDPQ2jvERe2ErPyjU_-2rDkZHv
p_7dcdALQ1r-eCdNPo1Dxy3vcchvs-edVTBhxvJ51dwEOQ9Gujzz-OLFh2qNHt3olHUaLOONN_r9AVRymuU7LKnH7kWbm65wydLc
vrEvKEx4HtAQsLJ21eBqK-JSZuwMT4t8uOOLQaPpQDzeFjr7hkThwU-imXhJ7pMbeiwffyC5xT8NLs8-o7rmex6jDTux70AUeKJ5
mmRoMM23St06G78_MCL1wp6OfUcOj9iZT_pZj78-yZBpZjPfwXkzlYTuhJiBX3t43RInP189dMGQfh2m2SyhGbw1Z-AyGLhs3HsN
Fid02zHyl8zVijHZQCHPMAV4xnx-f10lhp2ZduWp7EHdIyK8TRO6TNaQzgsWSQeoQTh-rBXdxPqGsTFh_9b73aElJGHGVynX6SQ0
mVipenjDP35KGH1CARg9-vvwI86bC_mMiyJ9bK4OjhdobZqL0Ow3eC4I1E6Q5feVCOg9aOr3e5afXbToYJAjAhfEwvEjRru_HOjn
syjoRiHCdcCjFDOFawzO9E8ao__y2LZJbzw0JjCfcEAel0nsYwLXuGa1Ui-ciWLAsR07nfkAkz-OvZACIcyvmue4HCY7bEJdNUCy
x-7Vl-TpQgmKtvRtf8G1tuCRAM1ToWtlCC_S7-s156Dl89XqVh96YdX3zzzV__EOpTIZxqe99yKNikzGtEAR81rLi10YcXHaj801
HXvXjzBvfZH_yO9FS-FEJB8QiNXhUp5vjuAqMB_THODR2__aAIFApShW1gPVqOPi4n0TgrY2LPpsCq5NGA2wcw88lgs9G0CgVS9d
iS5omVK1nf94cJx0UVOE0oXYQi9cJsYY936ZxzL-8rzgcZqQjJqMPPtKcjfdo10iI_1LXESVe-Ra0VCk8BDvUY_mJaxRm7zZrgOL
gGCgLdCqBLlLcIUUUoP6qvuJrxLl3mIpPRXf_P_n6mB6HLjY4aUmzHUK9qAGr1CZ4ELnTrH3lzdXtXxp9pcjsztdmtrbve_MDk1D
v7CnWHlY7hJCtKE9swljKJUpWRV6YRotbQA2AfWxbOcpkXQ16iTXPzoUveCOFo-zfXDJuES2x2-4lDjBL1uIMBFCOx4vduEGvwK8
OfvZD4EPHS1fSpkHTBeh4lPqMJMwHiMQnxIzajDZfmpbTlJyW1Oj5MjMYsbQ-nTgg8ffYHlyE3L2vYsmL7REecoUIYyiShG9m40k
hS1lWY08K5PkVvkcG2nLbj7mApb9M3WYVMxMrNp5bFIvRYcBajRUaVKDv3zjOeH7DxvyPOmROIABri4o1T5BKg6YYu2FnxlHpguc
47iV9dpSQJ6V_wIjRA0xm-qMcwnzmJIdkGFyDJCYFkb6AJEll2Vq6ZaJF1KNUtRhspmxB3tifsj_naW5xbqHPmZA1U9YefoRlBvd
bboRR1W6qCaaE0qTYhcyOSaltH2W4gDkaXlR_iqnNejbPpv6mCZh_sAsbCLcVIgAq15iatQFyXHFk-qRYezx_L5Y0mxVFmtqreFb
o4Zh53rySOBJ8pytDRjUjhEUSjc8zRgcss3Oze4JFJVa0t-xvEIKhLRM-p2Wnn-ilaE3OztYD12oV_L4qOqPjMjN49qZ51CM5ntZ
RWocI92XcLPncxyaqjgExei4P6Np76810jV3R8FaEUi1NtU4X2BXgzWc9zxNbOgK1qU8rve41ad4AunsGOkuFiwqOaQ7VJ4DUOkd
d8MoqaWD-i7p0Xn92Ra1yq4KpDDyZdWv1QOVv7h8YRbOuXdnTQInx0h3wvwW9mwyHCrPAagMjrthKCW29NukB4maBLcP5JynpWjN
1nHJPZnhlefwWlpVnK5MQTF6oEj7SRdC-ONJYKGTWeSyj41k8otDwjmA0pAujCqwqtM1pg6rbV4_fmpp18TOLHYG_URvww7e_XcJ
od51TuFnZwE-VMLt9pYvHXqO8AunWCwT5MZSC2cHFL3GYfWiSIXbsScsQI06xBhuDOikJdnBVl4UeUouIrvSNWdNvTi7M3tTGSMG
9rKdRD6UCXL9mDEbw0dOX_lqpSC_pg-FBtzZpyhu_317FHDb-q_SqYRap6DKUlDrVKh0CqpsRe2sw9VDL529MjQaJ2V7G7-ZVpc_
VKXpSI0KKJbzq_VBqBvI6MxVzWY46KJTHM8kTjC0VLdoFh7uY7d6qebd5JZHUsDqmn2bVb9ZOWVZ_YR2a1PnJYob2otM-CMOTfRi
h-XS3T78bo4qkdHSdKSKVJCmSMv5w0LnnEa4pHXPUKW12aBcaztBrWgMn7c70nvsliZfqi-7lW4o3PhKG4bfmNg3orlPie4Mc-as
dUrNMjx-qdbxQWlVkiOvVTkTkPXLNdU5v4Mc5edSyVFaKKXqSFUgnYRbVnucb3TNUF1v-6O42XBymzbac3fMd-rXZ_efOHFi7uSh
W9-ad679oRd_iVzp4oibmzY7gJfP_-fTNs2DvzN3feTItS0yq18cnFo3oXNa5fNVrwxNcr0QfBK92jbnQxc3JRub1_1LB99rlfKa
Hu6GEaNESCmEEFIlJUIihZSAVKRiVEmVUUiVUZGKQZGKQUhFr80K0DnlvPqLSnF6tU0b7SlzYzE1wH_-M8gITIf-06fOqj_ufPsj
87a3WOOZ4pFkk5Bqo4bpJ98__lREy2GW0mS1OrX7y8HTHDb3erm5l1v47ccrta20ouFUB1xRbUOx92j-nGcWAgtfeqX-U-fbHfng
rxY_-p_SHrFayNhjb2Q9oulW3F5SAHma0t1DmYSKshFPPtnmnhEOVrz_3nx14BpgzZQpjdrfiDz_-ZHWv9XYE7KlmP9h31OjZP_M
55uHN_exGlneoj_L_4coscdYsGDAb0DNl14KadCm9pPLT7Td3WxbrXXcckqjderDTLg2J7Z5C5-jJZVjVP5H-h3Y3E2_-eYzR4DI
yZOdQpoEdf3sUvTJTr3EU2tbt_S3aYOQvEebth5E2D02euutgSlAtzfe-MGt67RaNq9o_K-l34XD2xGnTXvMriUkee-DSzwwKLcI
pP_rSO-i3Eg3KnbodB1w456JUuG4v1p6BrAcmIBjLm02oeL1XLkFMrTakaYB64ETmLxz3YE6jtdl0MNXv8MexVSUvwJ1cqBrC_Cr
uKBM-Sg30s229GuYnP2TuEs2YN3T0DrOnYelJ-FIB_Lf40Rgt4T_7oOo5G3-7SYLb3lKlsGnCRxDOaoXM2_1ekzhAou66wUXT6rY
EIfl590wOx2OdKT4nQm4Fgkbeh_R8jG7xWBhZmG3fFCO6sVMR2puM0MW0Kj4aWu78XSZsPQ3-D0KjLbE-e9MOOF8Ll4UdeU78nUb
cpQpKrYjNectbgAiip9WWYk4t-0n8HSDtv9A7V9AfcEGgargwSzmiMlivg2pyxQV15Gex-Su5lwkYQBmvdLUxqIJ76Jnr8K__z0E
Px-CPWGQbS3uvzdqZjFJzBBaXiXO1miqpUXFdaT7wGywYzP6HEBlsH2zb_1GpuPYcVi-BY53xbJ7oQcqZjEcLSGijnhGnpTpHTvO
8Xdy9tgqGlyqXc2vyq6mPtHTRoxoWmafXCtH0ot0pOZUQAaY8WIEQGWlpVtCeBjMqw7vxcOunlgO1qAGJvII_mxs1Wn0BnWA1-sH
q69Qp7knQS5d3JIDunz4UlRuE882f7cMaj9z1AstSvVNJFH6j3fYhphtg42_xqy4296mU5yEbOAN8_k__fsLfJsPdrj-jzbDTx1A
lrD-5PqeL9nX0rAUa8cl14eaeS0NjT2iD0f6tpv34ui2dn9C2eaWHhPzag1FUY-qVq1qzc6d6z87YEATOw1eBdTLGUwbgYuSbs0N
Pk-jw-4wyHcx4mFw_xE2RKhlXmW9xWVD3fjrpgdvYe0o2ymNo06bVUfl5ibrL3ms_3hyK2Mz947HIn06_ic9rfXy2bNL_i6zRdJj
YuZ5oSZW5542UOdx46GssCvO6a7n2O2SSvL-_t3T0wdHjhjRtMTFizuQBTvSPyiuz28BfS3n12cppY5U_GxPqLJ2wNn366_woh6-
FhMOBt6nxEecq87ghMdW5QRb6313RftpjayoZUkzOyVG-rdZ3KiRbkmHDh305vLlq5eYmHlKrnPm0-m-58ake6U0uumZ7HrLNQWD
Yv6Rh9_sLmNrTB3y4ph2NrmGtdvxtHFH21WmFjaX4p_fygNetZz_sy3x-HR7wpaqrCLs7Llf671cdaEyXv2FsbnecszgzwG7Nt7f
hVtOANUzo2Qrz65nGvo3_SwszO-t7t3r5G_yV7fsPWbtdd9TnW5EnNFe90gURsW2LyUe894k5iefWX7lzTkd33ip73MlpVfnORsA
Necx2VoKTmKMQD3r-RVD6fyhFKMH7lmd9GT572psHDTz6Kr13kZPPRYcs03bGOdjfoRVArKcr3DU-Xtx6dbxGh3DZs0LD29xAliT
L0twsF-fU1W2-Fz1SrCZ8Du44pHAwpsjBw-Z9f6_K1cesMrK0NSRfjPXrHk1fFHLG8XC9GQBT1qvS22wb_QiUOGqa4XXpRdv-SZ-
9bPP3oOd5vZt1KfnwkdfOhT-ZdM8r0zT9ndLGtifUtmAvDOr80b0DLy6jkIpsq9UrXk4m9q_dOKkr2NeVTqna3xpnFzv4j_jr-p0
xsZDhzY9bS7dM8-0TIeWM-aueSqGhkQVvKatHEi65hpGLM9NVMaSW7pTbl2cb7XPVaU3PS6ut_okwLXJx_Xqkdux_xy1Trt3y_GG
38fccC8gnhr4GiwGOfTF_NC2BHhkB_Na87lUenQ8elVasevq81WOEdTDnwu_BZLpfNn-GgCDkssW97c9s_7JPqlbbOgzenTzeHPp
RGPxKA8XCbKWQmrlLTXbTAr87KmE4N3PJwYdqHms2l_qm153ZVEZ1ZDjg0BBFmiaKoM_LhkdjZr0FqmqGy3i1Rn132jd2i8FP7gT
hDwmZl6na1VPbkiovt5VXzSyhxpTwDUj5g0iDiykuOb6MqvBa4T2esniV2jURsXAn53XErV3MDv0n-HoByalMLIjYJEq65Du25z_
6N-cOLH11GKJqjIUtyIG7ZN8_e-uP08DC6DnAoAlSw40OVVp74wTQftjTlTeq00JOgFGDS7ZkQiDJ5r0NjdVNyN2qq-3mR8VFbAD
b6Bq4aqEmKO0725cn9hwyxMpvnssC64Cfga6FTj3O_ALpj2jdnQlTnoPXq79Gg_1m0We2nL4GNHsYHe5r9EmQlLDcHoziNNay3Ev
bYOgcfIgeoc-93udOtqOTz_dxAggaonm9GQ7fgU-qHOGM2ylmTwnLX4fZ-3ao07nMq4OraGt0jYw0G-bj4_3inr1rEdfbN9-blRO
0I3Nh2t_5ZnlbMMSVBVgKLAf047oDOz2mlQbXJhUZT7RQxaSpymsYSPOyv6hoay-m1YxDSVTKh0n6uHGXPotAJsENQNPXQgP5XZi
RN-HCWt8pF2l7NGtMQXzg9qMLkS4BE6ywhrhAH371svFtDvaph3SvXu_M90YnDfv7_BFwqiYHSYXRxLwNnAV09jcTsIVo4ZRgbNo
-9wH5GrMdmmFoC5o597T-Vva7HmeHcZlNq_eu-RpqZnZliFRj9A8-hJZvuvIcV7B9bxqVE4ZLQBEA9GcGArbAo9xlJ2WJv2Ow7u-
uldewwzRdfN09vqu4apnQsmZVEAmDs13hVR4TjuFh1_4khwn2zZWFCLdoNKTNOx3wl_vSULABiuZnAlJb0H_Ro_QrZOR3MDv0LkM
p1A_rRTotEJ5Gb8CI94M8khgnpTSqv-jI9B53lRvfOo9AmNCidrUl2u_d-JAwBrrgwQHl00Fgv7u4-k7agPZTjZPzguTDnC-8jHa
9Yjk4vbapLnf3d4opEJgegPaB3bJfv7Jyi4iZANZblO5aaHjNQoTnyJC9KRTkfBm-1krD0r7tvLZCL0mVwG47H-WH559m8pd6xDz
_WDO_nOWhIAN5JVREHyB4EnncQwa8zvZziXFCTaIgk82X6cXxK6ua4j5cxzbDIvwyAmmvr4d_TpUo3bUYXJ9F7vohKWwdXchRTaZ
mRkKNZiGZ4GowUkkcpLJNt-dndA75RYa_F0IOsmFF96iWko9unw3goQT-znju6NUzk8CweNirGHYuD0qnUvJMeNMwfcLkq4qTrpe
ncvJYT8xa_M8mj5yBEPwBgyq6xYjBJqDFLn0mRc1jAa0yT-ZQQ6HWCCTpQNTDtug1-SZtSAmhxwlefRRap6JoOP64RxO3c4VW_R9
EQgEPfLGMmLaPyqd65825jIUkkkxmiEd4JrfeRp0_5PckC8c-uJh4hFIDEh4Iv8BG4BdLJX_yGV2F2YH7qgXS0iq_g_bJn6Iz3Av
2rgMRptV3eayhVTophvLmKlH0LnbPrQWGAuTbk69wO1NqA7CaIDlX4PeX3-389zNJnYxweFCbYShBNLv4ESdPeyasoIqQ0JlV784
PLKtfzhYSIUumeOYMP0EOg_79hEX3UhhVqeDaWueSi0MBgf69hUfwr8FLfVHOMxhBpfHwq9ek2ezh4NEciT8N_Fv2O9EHO2K9089
2H31O3ROhd9sIRU6p49j4qxTZHluckCqwurFrE4H0yZUtcpJZS9Lm9bBjyncnWCc5BgnGCBTZblEsTCoLa8KWYIUkv31t6B6SE2z
ww-j2uzNvozvyVVnIKRCzI3xTJ6dTKb3DyUXVgSKwQd9TmFnE7UlnW4KLGAf9u-ElXtBf3uS4XXO7Xz631m95HF5zO7CHIbja74G
Rc9fjTeiru9E631PkLXJgOfVUKbMvUimdp0DJQp8zr57Jqy2z-aCZxXtzWCzUipWw2cUR1ICfPwDZNwmvHomvNztqffKl_CygV6d
y46Wazg65Sf6D61Epu9Kh8rxujI2T1x4vGHR80rrxJ5m567CaHtLP3YA3voCLrsBRogwwoxJENOp6T0bGpYHdK63SHeyHrTUElyz
onBOerFbdLS2WMgUJfJ85-E-N4r33Laql4O74Z1VcNEVNNnwWCC8MhN8AwGpVLgveEVAZfDH7dRbdGgaNgAABAxJREFUS9q0rPar
uetK3NBW21sffbTY07Clpf-xBRbGwxVXqJQJo6JhSNzdIBBK-blK3kcQeJ9ZeKptwyiLUUAUgFZnuy9S6wtb663pdEMefLYIFm0D
NyP09IPXJ0GHYgHk__9r6drLk_KU1B5m_I7vQg0QnBH6SpPjnabsrb8pv2ma4pYUz5ByCpYsAI9QeKE1tO9hiir9P4BbVjvUSeNi
oqO1Vh2x1ACxsRHGQxu679hbf1N-nGhzOn33Rkg4CGNfgaBqJQshK6Clh-TUOuGS414r29mmTy6VGVT6INwSF7wX3SqkxM8D5zPb
6GL04Cqpd51AzJHe6hEYMt02wm-XUe6kfzNgTo9xG5d8Xi2lfrnVKVDQnnn_RHSjVjaZOfKZjY2NOBt1_NHU_IJKYXspUHyF6PT5
Tw56fsLPH3To9PfA3JJi35YFvC9NyxEXuza2NX2h5tw8ueskN50pTFapAkKi4H6zt16v8_vX4UJKifHPtf-t34EXPWP_-9Yhr1v-
96wet4yuqJJGd4yO1tq8ClaI9EmDOq9scbRHLjhmBhAouN94Uq9NWPtBF-_1mhZVH7PfYF2GiI2NyP2458TGE__78dT6idEOvXXS
yihOnVcNt5ML3mzbOthWwzpgZnzS6nSPb4QUds1ITS37SYPPsXUfddGu07R_qPdoe4S415jZr_eCcb--X6fX7-NuavR2uudZ8CwT
0gntqY-PREc0mWKvPMWYrZnWYFR4UmtpS0u_3bIN2qPfLuvivU7dLrzXSHsFKC_Exkac-rbde9rRGxZtDLpibSNSYQijeW9yn5Q3
s7nayqGP3xZjNjY2IqvVqR7_WpscmVp2H4NPQvznVS8uc2pf7_FYRyqvCLzTK_bRiVs-6hN1sJdNVmtzLd3zxrNSdXZwtD16vCDM
PsbmyV0Hn_Y6ug_vwmpQoOB2s7fR5cLgrwJE9JDwh7QPZGiLyQO7rtd-EuBf92LTf9a2e7t6pqvlzdOiyKZh55xGuJycN6dNG63j
G7-klGaPGZ-tv_Dfi0_L7yQyXirypxt9DL_9u_HrhIQ0taU8D-IxY_Wqd2ufaWZEmv_z7l8fyO8k8juJ_MHgLXcd3rO3tHVaVNwh
uaFvIBXcb_Y1-iZsXF019RPndvV6PhMerrXRV-3BwNyn-k148ecPI3vsHK5TmdPf-X7xAr_kjzINN-pGFU9kJ6w9kZ93HH07ISHN
qaJbY3kcS5fuV4-NX7LL_3qVQi194c4V8juJ3H5lsvGPP9KalEVdFX6z99sx_8tNIyIPdzfkk_77Krkls4Pc-eeVqWVVR4Xf5P14
LF26v9LALTMvu-g85PvbfpB_Hjj2Z1mWX26bdx9ETF-78uuaLrXbhPnUrR0dXXZ92f8DYLtGj2iBhgsAAAAASUVORK5CYII=`



/*
Creates a directory tree for a new relish artifact.
Creates template versions of some of the files you will need.
projectType can currently be "webapp" or ""
*/
func initProject(relishRoot, projectPath, projectType string) (err error) {
   artifactDir := relishRoot + "/artifacts/" + projectPath
   slashPos := strings.Index(projectPath,"/")
   origin := projectPath[0:slashPos]
   artifact := projectPath[slashPos+1:]
   metadataFilePath := artifactDir + "/metadata.txt"
   date := time.Now().Format("2006/01/02")  
   
   err = gos.MkdirAll(artifactDir,0777)       
   if err != nil {
      return 
   } 
   mainPackageDir := artifactDir + "/v0.1.0/src/main"
   err = gos.MkdirAll(mainPackageDir,0777)       
   if err != nil {
      return 
   } 
   mainFilePath := mainPackageDir + "/main.rel"    
     
   if projectType == "webapp" {
      
      metadata := fmt.Sprintf(WEB_APP_METADATA_FILE,date,origin,artifact)     
      
      err = gos.WriteFile(metadataFilePath, ([]byte)(metadata), 0777)  
      if err != nil {
         return 
      }      
      
      // TODO Create web package directory with template index.html and dialog.rel and static/ with styles and an image and a static html file.
      webPackageDir := artifactDir + "/v0.1.0/src/web"
      err = gos.MkdirAll(webPackageDir,0777)       
      if err != nil {
         return 
      } 
      
      indexPath := webPackageDir + "/index.html"
      
      err = gos.WriteFile(indexPath, ([]byte)(INDEX_HTML_FILE), 0777)  
      if err != nil {
         return 
      }    
      
      visitCountPath := webPackageDir + "/visit_count.html"
      
      err = gos.WriteFile(visitCountPath, ([]byte)(VISIT_COUNT_HTML_FILE), 0777)  
      if err != nil {
         return 
      }      
      
      
      gotItPath := webPackageDir + "/got_it.html"
      
      err = gos.WriteFile(gotItPath, ([]byte)(GOT_IT_HTML_FILE), 0777)  
      if err != nil {
         return 
      }     
      
      tryAgainPath := webPackageDir + "/try_again.html"
      
      err = gos.WriteFile(tryAgainPath, ([]byte)(TRY_AGAIN_HTML_FILE), 0777)  
      if err != nil {
         return 
      }         
      
      dialogPath := webPackageDir + "/dialog.rel"
      
      dialogContent := fmt.Sprintf(DIALOG_FILE,origin,artifact) 
            
      err = gos.WriteFile(dialogPath, ([]byte)(dialogContent), 0777)  
      if err != nil {
         return 
      }      
      
      stylesDir := webPackageDir + "/static/styles"
      err = gos.MkdirAll(stylesDir,0777)       
      if err != nil {
         return 
      }    
      
      cssPath := stylesDir + "/default.css"
      
      err = gos.WriteFile(cssPath, ([]byte)(CSS_FILE), 0777)  
      if err != nil {
         return 
      }      
      
      staticPath := webPackageDir + "/static/example.html"
      
      err = gos.WriteFile(staticPath, ([]byte)(STATIC_HTML_FILE), 0777)  
      if err != nil {
         return 
      }      
      
      
      imgDir := webPackageDir + "/static/img"
      err = gos.MkdirAll(imgDir,0777)       
      if err != nil {
         return 
      }            
      
      imagePath := imgDir + "/relish_logo_small.png"    
   
      var imageBytes []byte
	   imageBytes,err = base64.URLEncoding.DecodeString(RELISH_LOGO_PNG) 
      if err != nil {
         return 
      }	       
      err = gos.WriteFile(imagePath, imageBytes, 0777)  
      if err != nil {
         return 
      }
      
      mainContent := fmt.Sprintf(WEB_APP_MAIN_PROGRAM_FILE,origin,artifact,artifact,origin,artifact) 
 
      err = gos.WriteFile(mainFilePath, ([]byte)(mainContent), 0777)  
      if err != nil {
         return 
      }      
      
      // TODO create index.html, dialog.rel,default.css,an image??,a static html file.
   } else if projectType == "" {
      
      metadata := fmt.Sprintf(APP_METADATA_FILE,date,origin,artifact)        
      
      err = gos.WriteFile(metadataFilePath, ([]byte)(metadata), 0777)  
      if err != nil {
         return 
      }            
      
      mainContent := fmt.Sprintf(APP_MAIN_PROGRAM_FILE,origin,artifact,artifact,origin,artifact)
      
      err = gos.WriteFile(mainFilePath, ([]byte)(mainContent), 0777)  
      if err != nil {
         return 
      }      
   } else {
      err = fmt.Errorf("Unrecognized relish project type '%s'.",projectType)
   }
   return
}