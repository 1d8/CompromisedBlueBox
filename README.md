# What Is Compromised Blue Box

Compromised Blue Box is a remote access trojan designed to work without the need to create the usual server that is required with traditional RATs. Instead it functions using the Dropbox API to issue commands and receive data. 

## Dropbox Setup

In order for Compromised Blue Box to work, you need an access token which can be acquired by creating an app with Dropbox [here](https://www.dropbox.com/developers/apps/create?_tk=pilot_lp&_ad=ctabtn1&_camp=create)

* Choose scoped access for the app 
* Give the app *Full Dropbox* access
* Give it a silly name, like Compromised Blue Box

Now we must give our new app the appropriate permissions. Go [here](https://www.dropbox.com/developers/apps/) and select the app you'd like to work with and select the *Permissions* tab. Listed below are the permissions that are required for proper functionality:

* account_info.write
* account_info.read
* files.metadata.write
* files.metadata.read
* files.content.write
* files.content.read
* files_requests.write
* files_requests.read

After adding these permissions, there should be a submit button, click this. Then switch back to the *Settings* tab and scroll down to the *Access Token* section and set the expiration to *No expiration* and then click generate:

![](/imgs/img3.png)

## Infection Flow

When a machine runs Compromised Blue Box, it will generate hash (this will be the machine's unique identifier) based off the results of the following executed commands:

* `hostname`
* `whoami`
* `getmac`

Then the machine will take this identifier and check to see if a folder exists on Dropbox with the same name:
    If it does:
        * It will simply do nothing
    else:
        * It creates the folder with its unique identifier. Then it performs a "check-in"

After this, it will go inside the newly created folder on Dropbox and then check to see if a log file exists:
    If it does:
        * This log file will be downloaded and a new event will be written to it that simply says the victim was performing a "check-in". Then it will reupload the log file to its unique Dropbox folder, overwriting the old one.
    else:
        * It will create the log file, write a new event to it to indicate that the machine was simply "checking-in" and upload this log file to the machine's unique Dropbox folder.

## Issuing Commands and Receiving Data

To issue commands, a text file with the desired command and who the command is meant for must be present at the base directory of Dropbox.

Commands follow the following structure:
	
* `everyone or a UUID+-the desired command+-any optional arguments`
* EX: ![](/imgs/img1.png) <- this will create a profile (Windows version, architecture, etc) of every machine that's infected 

Notice that each component of the command must be separated by "+-". 

The result of the command is then uploaded to the victim's unique folder with the name results-currentTime&DateStamp.txt where the second part is a timestamp. Any errors encountered when executing the command is written to the log file.

## Available Commands

The list of available commands is very short as the purpose of this write-up is mainly to show how to use Dropbox (or any other file-sharing services that have public APIs) as a C&C server and provide sample code for doing so.

The short list of commands you can issue are:

1. Profile
    * Grabs the Windows version & architecture.
    * Syntax when issuing: `UUID or everyone+-profile`
2. Location
    * Grabs information from victim machine such as: public ip address, ISP, whether or not ip address used for web request is a proxy, city, region, country, zip code, and a rough estimate of the coordinates. Uses https://ip-api.com to accomplish this.  
    * Syntax when issuing: `UUID or everyone+-location`
3. downloadNexecute
    * Will download a binary from a user-specified URL to a file and execute that binary. The user must specify what to name the file when it's downloaded and the environment variable of where to download the file to (EX: APPDATA, USERPROFILE, TEMP).
    * Syntax when issuing: `UUID or everyone+-downloadNexecute+-URL+-filename+-environment variable`

The results of an executed command are written to a file that is prepended with the word *results* and has a timestamp with the date and the time it was uploaded. This file is located in a user's unique folder:

![](/imgs/img13.png)

## Tailoring The Code

There are only a few things that should be changed to make the code truly yours. These are all listed in this section

1. The name of the command file. By default the code searches for a file called **tasks.txt** to find the commands. This can be changed by changing the argument passed to *downloadDBFile* in the **Main** function (see img).
    * ![](/imgs/img2.png) 
2. The log file name. By default the code will name the log file **log.txt**. This is how it will be named in each machine's unique Dropbox folder. Simply change the variable **logFileName**  in the Var() section at the beginning of the code if you'd like.
    * ![](/imgs/img4.png) 
4. The access token. Change this to yours. I explain how to acquire it in the *Dropbox Setup*. And make sure you have the appropriate permissions granted to the access token. You simply need to change the variable **token** which is located in the Var() section.

## Cross-Compilation
Before compiling, note that you need to install the **grequests** library as this is used for the downloadNexecute command. To do so, run the following command:

* `go get -u github.com/levigross/grequests`

To compile this code to a 64-bit PE binary:

* GOOS=windows GOARCH=amd64 go build `filename`

To compile this code to a 32-bit PE binary:

* GOOS=windows GOARCH=386 go build `filename`
