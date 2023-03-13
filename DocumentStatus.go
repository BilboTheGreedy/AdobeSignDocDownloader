package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"

	"github.com/TwiN/go-color"
	scribble "github.com/nanobox-io/golang-scribble"
)

type DocumentStatus struct {
	//Name   string `json:"name"`
	User   string `json:"User"`
	Status string `json:"status"`
	Count  int
}

func NewTable() {
	cfg.QueryEndpoint()
	// Read JSON file
	dataFile, err := os.Open("./Data/DocumentStatus.json")
	if err != nil {
		log.Fatal(err)
	}
	defer dataFile.Close()

	byteValue, err := ioutil.ReadAll(dataFile)
	if err != nil {
		log.Fatal(err)
	}

	var data []DocumentStatus
	json.Unmarshal(byteValue, &data)
	// Map of maps to store counts for each user and status
	counts := make(map[string]map[string]int)

	// Loop over slice of structs
	for _, d := range data {
		// Get the user and status from each struct
		user := d.User
		status := d.Status

		// Check if user exists in counts map
		if _, ok := counts[user]; !ok {
			// If not, create a new map for that user
			counts[user] = make(map[string]int)
		}

		// Increment the count for that user and status combination
		counts[user][status]++
	}
	v_data := []DocumentStatus{}
	// Print the counts for each user and status
	for user, statuses := range counts {
		fmt.Println("User:", user)
		for status, count := range statuses {
			v_data = append(v_data, DocumentStatus{user, status, count})
			fmt.Println("Status:", status, "Count:", count)
		}
	}

	t := template.Must(template.ParseFiles("templates/status.html", "templates/Links.html"))

	f, err := os.Create("status.html")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	t.Execute(f, v_data)
	return
}

func VerifyPaths(StatusRequested string) {
	db, _ := scribble.New("./Data", nil)
	// Read from the cache database
	groups, _ := db.ReadAll("Groups")

	// iterate
	c_data := Data{}
	for _, group := range groups {
		f := GroupInfoList{}
		json.Unmarshal([]byte(group), &f)
		c_data.Groups.GroupInfoList = append(c_data.Groups.GroupInfoList, &f)
	}
	for _, group := range c_data.Groups.GroupInfoList {
		for _, member := range group.GroupMembers.UserInfoList {
			for _, agreement := range member.Agreements.UserAgreementList {
				if agreement.Status == StatusRequested {
					if agreement.IsAgreementDownloaded(cfg, group.GroupName, member.ID, member.Email) == true {
						println(color.Colorize(color.Green, "Yes - "+agreement.Name))
						//data = append(data, Entry{agreement.ID, agreement.Name, agreement.Status, member.Email, group.GroupName, "true", agreement.GetAgreementPath(cfg, group.GroupName, member.ID, member.Email)})
					} else {
						println(color.Colorize(color.Red, "No - "+agreement.Name))
						//data = append(data, Entry{agreement.ID, agreement.Name, agreement.Status, member.Email, group.GroupName, "false", agreement.GetAgreementPath(cfg, group.GroupName, member.ID, member.Email)})
					}
				}

			}
		}
	}
}

func VerifyTable() {

	data := []Entry{}
	db, _ := scribble.New("./Data", nil)
	// Read from the cache database
	groups, _ := db.ReadAll("Groups")

	// iterate
	c_data := Data{}
	for _, group := range groups {
		f := GroupInfoList{}
		json.Unmarshal([]byte(group), &f)
		c_data.Groups.GroupInfoList = append(c_data.Groups.GroupInfoList, &f)
	}
	for _, group := range c_data.Groups.GroupInfoList {
		for _, member := range group.GroupMembers.UserInfoList {
			for _, agreement := range member.Agreements.UserAgreementList {

				if agreement.IsAgreementDownloaded(cfg, group.GroupName, member.ID, member.Email) == true {
					//println(color.Colorize(color.Green, "Yes - "+agreement.Name))
					data = append(data, Entry{agreement.ID, agreement.Name, agreement.Status, member.Email, group.GroupName, "true", agreement.GetAgreementPath(cfg, group.GroupName, member.ID, member.Email)})
				} else {
					//println(color.Colorize(color.Red, "No - "+agreement.Name))
					data = append(data, Entry{agreement.ID, agreement.Name, agreement.Status, member.Email, group.GroupName, "false", agreement.GetAgreementPath(cfg, group.GroupName, member.ID, member.Email)})
				}

			}
		}
	}

	t := template.Must(template.ParseFiles("templates/table.html", "templates/Links.html"))

	f, err := os.Create("table.html")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	t.Execute(f, data)
	return
}
