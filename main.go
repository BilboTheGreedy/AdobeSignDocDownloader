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
	"sync"
	"time"
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

func main() {
	start := time.Now()
	accessTokenPtr := flag.String("accesstoken", "none", "API Access Token")
	outPathPtr := flag.String("out", "C:\\documents", "Download Location")
	pVerbose := flag.Bool("v", false, "Explain what's happening while program runs")
	flag.Parse()
	fmt.Println("Access Token:", *accessTokenPtr)
	fmt.Println("out path:", *outPathPtr)
	fmt.Fprintf(os.Stdout, "Verbose mode set? %v\n", *pVerbose)
	var wg sync.WaitGroup
	var ACCESSTOKEN = *accessTokenPtr
	baseuri := baseURI{}
	groups := Response{}
	groupMembers := GroupMembers{}
	agreements := Agreements{}
	file_path := *outPathPtr

	GetURI(ACCESSTOKEN, &baseuri, *pVerbose)

	GetGroups(ACCESSTOKEN, baseuri.APIAccessPoint, &groups, *pVerbose)
	for _, group := range groups.GroupInfoList {
		fmt.Println("GroupName:", group.GroupName+"\tGroupID:", group.GroupID)
		GetGroupsMembers(ACCESSTOKEN, baseuri.APIAccessPoint, &groupMembers, group.GroupID, *pVerbose)
		fmt.Println("Main : Submitting job to workers")
		for _, user := range groupMembers.UserInfoList {
			fmt.Println("Email:", user.Email+"\tID:", user.ID)
			GetUserAgreements(ACCESSTOKEN, baseuri.APIAccessPoint, &agreements, user.ID, *pVerbose)
			agreementsCount := len(agreements.UserAgreementList)
			wg.Add(agreementsCount)
			for _, agreemment := range agreements.UserAgreementList {
				fmt.Println("Downloading document:", agreemment.Name+" ID:", agreemment.ID)
				fullPath := file_path + "\\" + group.GroupName + "\\" + user.Email + "\\" + agreemment.Name + "(" + agreemment.ID + ")\\"
				go DownloadAgreement(ACCESSTOKEN, baseuri.APIAccessPoint, agreemment.ID, user.ID, fullPath, agreemment.Name+".pdf", &wg, *pVerbose)
			}
		}
	}
	fmt.Println("Waiting for goroutines to finish...")
	wg.Wait()
	fmt.Println("Done!")
	// Code to measure
	duration := time.Since(start)

	// Formatted string, such as "2h3m0.5s" or "4.503μs"
	fmt.Println(duration)

	// Nanoseconds as int64
	fmt.Println(duration.Nanoseconds())
}
