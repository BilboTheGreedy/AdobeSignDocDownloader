package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"sync"
	"time"

	"github.com/manifoldco/promptui"
)

type Agreements struct {
	Page struct {
		NextCursor string `json:"nextCursor"`
	} `json:"page"`
	UserAgreementList []struct {
		DisplayDate                string `json:"displayDate"`
		DisplayParticipantSetInfos []struct {
			DisplayUserSetMemberInfos []struct {
				Email    string `json:"email"`
				Company  string `json:"company"`
				FullName string `json:"fullName"`
			} `json:"displayUserSetMemberInfos"`
			DisplayUserSetName string `json:"displayUserSetName"`
		} `json:"displayParticipantSetInfos"`
		Esign           bool   `json:"esign"`
		GroupID         string `json:"groupId"`
		Hidden          bool   `json:"hidden"`
		LatestVersionID string `json:"latestVersionId"`
		Name            string `json:"name"`
		ID              string `json:"id"`
		ParentID        string `json:"parentId"`
		Status          string `json:"status"`
		Type            string `json:"type"`
	} `json:"userAgreementList"`
}
type groupInfoList struct {
	GroupInfoList []struct {
		GroupID        string    `json:"groupId"`
		CreatedDate    time.Time `json:"createdDate"`
		GroupName      string    `json:"groupName"`
		IsDefaultGroup bool      `json:"isDefaultGroup"`
	} `json:"groupInfoList"`
}

type Response struct {
	GroupInfoList []struct {
		GroupID        string    `json:"groupId"`
		CreatedDate    time.Time `json:"createdDate"`
		GroupName      string    `json:"groupName"`
		IsDefaultGroup bool      `json:"isDefaultGroup"`
	} `json:"groupInfoList"`
	Page struct {
		NextCursor string `json:"nextCursor"`
	} `json:"page"`
}

type GroupMembers struct {
	Page struct {
		NextCursor string `json:"nextCursor"`
	} `json:"page"`
	UserInfoList []struct {
		Email        string `json:"email"`
		ID           string `json:"id"`
		IsGroupAdmin bool   `json:"isGroupAdmin"`
		Company      string `json:"company"`
		FirstName    string `json:"firstName"`
		LastName     string `json:"lastName"`
	} `json:"userInfoList"`
}
type baseURI struct {
	APIAccessPoint string `json:"apiAccessPoint"`
	WebAccessPoint string `json:"webAccessPoint"`
}

func GetURI(ACCESSTOKEN string, target interface{}, debug bool) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	url := "https://api.na1.adobesign.com/api/rest/v6/baseUris"
	if debug == true {
		fmt.Println("URL:>", url)
	}

	req, err := http.NewRequest("GET", url, nil)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+ACCESSTOKEN)

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(&target)

}

func GetGroups(ACCESSTOKEN string, baseUri string, target interface{}, debug bool) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	url := baseUri + "api/rest/v6/groups?pageSize=10000"
	if debug == true {
		fmt.Println("URL:>", url)
	}

	req, err := http.NewRequest("GET", url, nil)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+ACCESSTOKEN)

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(&target)

}

func GetGroupsMembers(ACCESSTOKEN string, baseUri string, target interface{}, GroupID string, debug bool) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	url := baseUri + "api/rest/v6/groups/" + GroupID + "/users?pageSize=10000"
	if debug == true {
		fmt.Println("URL:>", url)
	}

	req, err := http.NewRequest("GET", url, nil)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+ACCESSTOKEN)

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(&target)

}

func GetUserAgreements(ACCESSTOKEN string, baseUri string, target interface{}, UserID string, debug bool) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	url := baseUri + "api/rest/v6/agreements?pageSize=10000"
	if debug == true {
		fmt.Println("URL:>", url)
	}

	req, err := http.NewRequest("GET", url, nil)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-user", "userid:"+UserID)
	req.Header.Add("Authorization", "Bearer "+ACCESSTOKEN)

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(&target)

}

func DownloadAgreement(ACCESSTOKEN string, baseUri string, AgreementID string, UserID string, filePath string, fileName string, wg *sync.WaitGroup, debug bool) error {
	defer wg.Done()
	path := filePath
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Println(err)
	}

	// Create the file
	out, err := os.Create(filePath + fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	url := baseUri + "api/rest/v6/agreements/" + AgreementID + "/combinedDocument"
	if debug == true {
		fmt.Println("URL:>", url)
	}

	req, err := http.NewRequest("GET", url, nil)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-user", "userid:"+UserID)
	req.Header.Add("Authorization", "Bearer "+ACCESSTOKEN)

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

var r, _ = regexp.Compile("\\\\|/|:|\\*|\\?|<|>")

// MakeFilenameWindowsFriendly removes characters not permitted in file/directory names on Windows
func MakeFilenameWindowsFriendly(name string) string {
	return r.ReplaceAllString(name, "")
}

func Run(ACCESSTOKEN string, out string, groupName string, groupID string, selectAUser bool, pVerbose bool) {

	start := time.Now()

	var count = 0
	var wg sync.WaitGroup
	baseuri := baseURI{}
	var gn []string
	var UserID string
	var UserEmail string
	var agreementsCount int
	groupMembers := GroupMembers{}
	agreements := Agreements{}
	file_path := out

	GetURI(ACCESSTOKEN, &baseuri, pVerbose)

	//GetGroups(ACCESSTOKEN, baseuri.APIAccessPoint, &groups, pVerbose)

	fmt.Println("GroupName:", groupName+"\tGroupID:", groupID)
	GetGroupsMembers(ACCESSTOKEN, baseuri.APIAccessPoint, &groupMembers, groupID, pVerbose)
	for _, v := range groupMembers.UserInfoList {
		gn = append(gn, v.Email)
	}
	prompt := promptui.Select{
		Label: "Select User",
		Items: gn,
		Size:  20,
	}

	_, selectedUser, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	fmt.Printf("You choose %q\n", selectedUser)
	for _, v := range groupMembers.UserInfoList {
		if v.Email == selectedUser {
			UserID = v.ID
			UserEmail = v.Email
		}
	}
	fmt.Println("Main : Submitting job to workers")

	//Selected User
	if selectAUser {
		fmt.Println("Email:", UserEmail+"\tID:", UserID)
		GetUserAgreements(ACCESSTOKEN, baseuri.APIAccessPoint, &agreements, UserID, pVerbose)

		for _, v := range agreements.UserAgreementList {
			if v.Status == "SIGNED" {
				agreementsCount++
			}

		}
		wg.Add(agreementsCount)
		count += agreementsCount

		for _, agreemment := range agreements.UserAgreementList {
			if agreemment.Status == "SIGNED" {
				fmt.Println("\t- Downloading document:", agreemment.Name)
				fullPath := file_path + "\\" + groupName + "\\" + UserEmail + "\\" + MakeFilenameWindowsFriendly(agreemment.Name) + "(" + agreemment.ID + ")\\"
				go DownloadAgreement(ACCESSTOKEN, baseuri.APIAccessPoint, agreemment.ID, UserID, fullPath, MakeFilenameWindowsFriendly(agreemment.Name)+".pdf", &wg, pVerbose)
			}

		}
		fmt.Println("Count:", count)
	} else {

		//All Users
		for _, user := range groupMembers.UserInfoList {
			fmt.Println("Email:", user.Email+"\tID:", user.ID)
			GetUserAgreements(ACCESSTOKEN, baseuri.APIAccessPoint, &agreements, user.ID, pVerbose)
			agreementsCount := len(agreements.UserAgreementList)
			wg.Add(agreementsCount)
			count += agreementsCount

			for _, agreemment := range agreements.UserAgreementList {
				fmt.Println("\t- Downloading document:", agreemment.Name)
				fullPath := file_path + "\\" + groupName + "\\" + user.Email + "\\" + MakeFilenameWindowsFriendly(agreemment.Name) + "(" + agreemment.ID + ")\\"
				go DownloadAgreement(ACCESSTOKEN, baseuri.APIAccessPoint, agreemment.ID, user.ID, fullPath, MakeFilenameWindowsFriendly(agreemment.Name)+".pdf", &wg, pVerbose)
			}
			fmt.Println("Count:", count)
		}

	}

	fmt.Println("Main: Waiting for goroutines to finish before next group...")
	wg.Wait()

	fmt.Println("Done!")
	// Code to measure
	duration := time.Since(start)

	// Formatted string, such as "2h3m0.5s" or "4.503μs"
	fmt.Println(duration)

	// Nanoseconds as int64
	fmt.Println(duration.Nanoseconds())
}
func GetStructFieldName(Struct interface{}, StructField ...interface{}) (fields map[int]string) {
	fields = make(map[int]string)

	for r := range StructField {
		s := reflect.ValueOf(Struct).Elem()
		f := reflect.ValueOf(StructField[r]).Elem()

		for i := 0; i < s.NumField(); i++ {
			valueField := s.Field(i)
			if valueField.Addr().Interface() == f.Addr().Interface() {
				fields[i] = s.Type().Field(i).Name
			}
		}
	}
	return fields
}

func main() {
	accessTokenPtr := flag.String("accesstoken", "none", "API Access Token")
	outPathPtr := flag.String("out", "C:\\documents", "Download Location")
	pVerbose := flag.Bool("v", false, "Explain what's happening while program runs")

	flag.Parse()
	fmt.Println("Access Token:", *accessTokenPtr)
	fmt.Println("out path:", *outPathPtr)
	fmt.Fprintf(os.Stdout, "Verbose mode set? %v\n", *pVerbose)

	var ACCESSTOKEN = *accessTokenPtr
	baseuri := baseURI{}
	groups := Response{}
	var selectUser bool
	var gn []string
	var groupid string
	var groupname string
	GetURI(ACCESSTOKEN, &baseuri, *pVerbose)

	GetGroups(ACCESSTOKEN, baseuri.APIAccessPoint, &groups, *pVerbose)

	for _, v := range groups.GroupInfoList {
		gn = append(gn, v.GroupName)
	}

	prompt := promptui.Select{
		Label: "Select Group",
		Items: gn,
		Size:  20,
	}
	promptB := promptui.Prompt{
		Label:     "Select Specifc User?",
		IsConfirm: true,
	}

	selectAUser, err := promptB.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	fmt.Printf("You choose %q\n", selectAUser)
	if selectAUser == "y" {
		selectUser = true
	} else {
		selectUser = false
	}

	_, selectedGroup, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	for _, v := range groups.GroupInfoList {
		if v.GroupName == selectedGroup {
			groupid = v.GroupID
			groupname = v.GroupName
		}
	}
	fmt.Printf("You choose %q\n", selectedGroup)
	Run(ACCESSTOKEN, *outPathPtr, groupname, groupid, selectUser, *pVerbose)
}
