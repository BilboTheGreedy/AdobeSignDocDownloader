package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"

	"github.com/TwiN/go-color"
	scribble "github.com/nanobox-io/golang-scribble"
	"github.com/schollz/progressbar/v3"
)

var (
	done            chan struct{}
	ch              = make(chan int)
	wg              sync.WaitGroup
	debug           bool
	cfg             Configuration
	pb              progressbar.ProgressBar
	ConsoleText     bool
	StatusRequested string
)

var r, _ = regexp.Compile("\\\\|\\||-|\"|/|:|\\*|\\?|<|>")

// MakeFilenameWindowsFriendly removes characters not permitted in file/directory names on Windows
func MakeFilenameWindowsFriendly(name string) string {
	return r.ReplaceAllString(name, " ")
}

func main() {
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	er := io.MultiWriter(os.Stderr, logFile)
	log.SetOutput(er)
	bCache := flag.Bool("v", false, "Cache Data")
	bConsoleText := flag.Bool("console", false, "output console text")
	bDebug := flag.Bool("debug", false, "Explain what's happening while program runs")
	iMaxJobs := flag.Int("max", 40, "Max number of downloads concurrently")
	sStatusReq := flag.String("status", "SIGNED", "Document Status")
	flag.Parse()
	fmt.Fprintf(os.Stdout, "Cache mode set? %v\n", *bCache)
	debug = *bDebug
	ConsoleText = *bConsoleText
	StatusRequested = *sStatusReq
	cfg = LoadConfiguration("config.json")
	cfg.QueryEndpoint()
	if *bCache == true {
		// create a new scribble database, providing a destination for the database to live
		db, _ := scribble.New("./Data", nil)

		//Query endpoint for API URI
		cfg.QueryEndpoint()

		//Struct to store data in
		data := Data{}

		//Query All groups
		data.QueryGroups(cfg)

		//Set up for goroutines: GROUPS/Members
		groupCount := len(data.Groups.GroupInfoList)
		fmt.Println("Main :", groupCount, "groups to collect data from in config")
		groups := make(chan *GroupInfoList, groupCount)
		done := make(chan bool, groupCount)
		fmt.Println("Main : Submitting job to workers")

		for i, _ := range data.Groups.GroupInfoList {
			go GetGroupMembersWorker(i, cfg.Session.AccessToken, cfg.Session.baseURI.APIAccessPoint, groups, done)
		}

		for _, group := range data.Groups.GroupInfoList {
			groups <- group

		}
		close(groups)
		for i := 0; i < groupCount; i++ {
			<-done
		}
		//Goroutine done for: GROUP/Members

		//Goroutine start for: USER/Agreements
		for _, group := range data.Groups.GroupInfoList {
			fmt.Println(group)
			if len(group.GroupMembers.UserInfoList) != 0 {
				memberCount := len(group.GroupMembers.UserInfoList)
				members := make(chan *UserInfoList, memberCount)
				done := make(chan bool, groupCount)
				fmt.Println("Main :", memberCount, "user to collect data from in config")
				for i, _ := range group.GroupMembers.UserInfoList {
					//member.QueryUseAgreement(cfg.Session.AccessToken, cfg.Session.baseURI.APIAccessPoint)
					go GetUserAgreementsWorker(i, cfg.Session.AccessToken, cfg.Session.baseURI.APIAccessPoint, members, done)

				}
				for _, member := range group.GroupMembers.UserInfoList {
					members <- member

				}
				close(members)
				for i := 0; i < memberCount; i++ {
					<-done
				}
			}

		}
		//Save Cache to fs
		for _, item := range data.Groups.GroupInfoList {
			db.Write("Groups", item.GroupName, item)
		}

		wg.Wait()

	} else { //Read from cache
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
		data_agreementCount := c_data.CountAgreements(StatusRequested)
		fmt.Println("Number of agreements to process:", data_agreementCount)
		//bar := progressbar.Default(int64(data_agreementCount))
		if ConsoleText != true {
			pb = *progressbar.Default(int64(data_agreementCount))
			pb.ChangeMax(data_agreementCount)
		}
		//for each group
		for _, g := range c_data.Groups.GroupInfoList {
			//for each member
			for _, m := range g.GroupMembers.UserInfoList {

				//Maximum of concurrent downloads at one time
				//So Adobe don't lock us out
				maxGoroutines := *iMaxJobs
				sem := make(chan int, maxGoroutines)
				//for each agreement

				for i, a := range m.Agreements.UserAgreementList {
					if a.Status == StatusRequested {
						if ConsoleText != true {
							pb.Add(1)
						}

						sem <- 1
						go func() {

							DownloadWorker(i, cfg, g, m, a, false)
							<-sem // removes an int from sem, allowing another to proceed
						}()

					}

				}

			}
		}

	}

	logFile.Close()
}

func DownloadWorker(id int, cfg Configuration, group *GroupInfoList, User *UserInfoList, agreement *UserAgreementList, debug bool) {

	if ConsoleText == true {
		println(color.Colorize(color.Yellow, "Worker "+strconv.Itoa(id)+": Received job to download document:"+agreement.Name))
	}

	agreement.DownloadUserAgreement(cfg, group.GroupName, User.ID, User.Email)

}

func (data *Data) CountAgreements(StatusRequested string) int {
	var count int
	for _, g := range data.Groups.GroupInfoList {
		for _, m := range g.GroupMembers.UserInfoList {
			for _, a := range m.Agreements.UserAgreementList {
				if a.Status == StatusRequested {
					count++
				}

			}

		}

	}
	return count
}
func (data *UserAgreementList) DownloadUserAgreement(cfg Configuration, GroupName string, UserID string, UserEmail string) {

	fullPath := cfg.DownloadLocation + "\\" + GroupName + "\\" + UserEmail + "\\" + MakeFilenameWindowsFriendly(data.Name) + "(" + data.ID + ")\\"

	if _, err := os.Stat(fullPath + MakeFilenameWindowsFriendly(data.Name) + ".pdf"); err == nil {
		if ConsoleText == true {
			println(color.Colorize(color.Cyan, "File already exists: "+fullPath+data.Name+".pdf"))
		}

	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist
		err := DownloadAgreement(cfg.Session.AccessToken, cfg.Session.baseURI.APIAccessPoint, data.ID, UserID, fullPath, MakeFilenameWindowsFriendly(data.Name)+".pdf", debug)

		if err != nil {
			if ConsoleText == true {
				println(color.Colorize(color.Red, "Download Failed: "+fullPath+data.Name))
			}
			return
		} else {
			if ConsoleText == true {
				println(color.Colorize(color.Green, "Download Success: "+data.Name))
			}
		}
	} else {
		// file may or may not exist. See err for details.
		fmt.Println(err)
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence

	}

}

func DownloadAgreement(ACCESSTOKEN string, baseUri string, AgreementID string, UserID string, filePath string, fileName string, debug bool) error {
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

func (data *UserInfoList) QueryUseAgreement(AccessToken string, URI string) {

	GetUserAgreements(AccessToken, URI, &data.Agreements, data.ID, debug)
	for _, a := range data.Agreements.UserAgreementList {
		fmt.Println("Found Agreement", a.ID, "On user:", data.ID)
	}

}

func GetUserAgreementsWorker(id int, AccessToken string, URI string, members <-chan *UserInfoList, done chan<- bool) {
	for member := range members {

		if debug != true {
			fmt.Println("Worker", id, ": Received job", member.Email)
		}

		member.QueryUseAgreement(AccessToken, URI)
		done <- true
	}

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

func (cfg *Configuration) QueryEndpoint() {
	//GetEndpoint(cfg.Session.AccessToken, cfg)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	url := "https://api.na1.adobesign.com/api/rest/v6/baseUris"
	if debug == true {
		fmt.Println("URL:>", url)
	}

	req, err := http.NewRequest("GET", url, nil)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+cfg.Session.AccessToken)

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&cfg.Session.baseURI)

}

func (data *Data) QueryGroups(cfg Configuration) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	url := cfg.Session.baseURI.APIAccessPoint + "api/rest/v6/groups?pageSize=10000"
	if debug == true {
		fmt.Println("URL:>", url)
	}

	req, err := http.NewRequest("GET", url, nil)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+cfg.Session.AccessToken)

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&data.Groups)

}

func (data *GroupInfoList) QueryGroupMembers(AccessToken string, URI string) {
	GetGroupMembers(AccessToken, URI, data.GroupID, data.GroupName, &data.GroupMembers, debug)

}

func GetGroupMembers(AccessToken string, URI string, GroupID string, GroupName string, target interface{}, debug bool) error {
	fmt.Println("Group Members for group", GroupName)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	url := URI + "api/rest/v6/groups/" + GroupID + "/users?pageSize=10000"
	if debug == true {
		fmt.Println("URL:>", url)
	}

	req, err := http.NewRequest("GET", url, nil)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+AccessToken)

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(&target)
}

func GetGroupMembersWorker(id int, AccessToken string, URI string, groups <-chan *GroupInfoList, done chan<- bool) {
	for group := range groups {
		if debug != true {
			fmt.Println("Worker", id, ": Received job", group.GroupName)
		}

		group.QueryGroupMembers(AccessToken, URI)
		done <- true
	}

}

func LoadConfiguration(file string) Configuration {
	var config Configuration
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}
