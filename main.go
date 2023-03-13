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
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/TwiN/go-color"
	"github.com/go-echarts/go-echarts/v2/components"
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

var r, _ = regexp.Compile("\\\\|\\||-|\"|/|:|\\*|\\?|<|>|\\s+")

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
	bCache := flag.Bool("cache", false, "Make Cache Data")
	bConsoleText := flag.Bool("console", false, "output console text")
	bDebug := flag.Bool("debug", false, "Explain what's happening while program runs")
	bChart := flag.Bool("chart", false, "Explain what's happening while program runs")
	iMaxJobs := flag.Int("max", 40, "Max number of downloads concurrently")
	sStatusReq := flag.String("status", "SIGNED", "Document Status")
	bVerify := flag.Bool("verify", false, "Explain what's happening while program runs")
	sHttpProxy := flag.String("proxyaddr", "", "set proxy server")
	flag.Parse()
	//Set Env
	if len(*sHttpProxy) > 0 {
		fmt.Println("Set Http Proxy to:", *sHttpProxy)
		os.Setenv("HTTP_PROXY", *sHttpProxy)
		os.Setenv("HTTPs_PROXY", *sHttpProxy)
		os.Setenv("http_proxy", *sHttpProxy)
		os.Setenv("https_proxy", *sHttpProxy)
	}
	fmt.Fprintf(os.Stdout, "\nAdobe Sign Document Downloader by Daniel.Rapp@Kontract.se\n")
	fmt.Fprintf(os.Stdout, "\nSource: https://github.com/BilboTheGreedy/AdobeSignDocDownloader\n")
	fmt.Fprintf(os.Stdout, "\nCache mode set? %v\n", *bCache)
	debug = *bDebug
	ConsoleText = *bConsoleText
	StatusRequested = *sStatusReq
	cfg = LoadConfiguration("config.json")
	cfg.QueryEndpoint()

	if *bVerify == true {

		VerifyPaths(StatusRequested)
		return
	}

	if *bChart == true {
		GenerateStatusTable()
		VerifyTable()

		f, err := os.Create("charts.html")
		if err != nil {
			fmt.Println(err)
		}
		page := components.NewPage()
		page.Assets.AddCustomizedCSSAssets("background-color: coral")
		page.Theme = "vintage"
		page.PageTitle = "Adobe Sign Graphs"
		page.SetLayout(components.PageCenterLayout)

		page.AddCharts(
			GetADInfo(),
			GetPieData(),
		)

		page.Render(io.MultiWriter(f))
		return
	}

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

		for _, g := range data.Groups.GroupInfoList {

			//for each member
			for _, m := range g.GroupMembers.UserInfoList {
				var ma = m
				//Maximum of concurrent downloads at one time
				//So Adobe don't lock us out
				maxGoroutines := *iMaxJobs
				sem := make(chan int, maxGoroutines)
				//for each agreement

				for i, a := range m.Agreements.UserAgreementList {
					var ag = a
					if a.Status == StatusRequested {
						if ConsoleText != true {
							pb.Add(1)
						}

						sem <- 1
						go func(i int) {
							ag.GetDocuments(ma.ID)
							<-sem // removes an int from sem, allowing another to proceed
						}(i)

					}

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
			var ga = g
			//for each member
			for _, m := range g.GroupMembers.UserInfoList {
				var ma = m
				//Maximum of concurrent downloads at one time
				//So Adobe don't lock us out
				maxGoroutines := *iMaxJobs
				sem := make(chan int, maxGoroutines)
				//for each agreement

				for i, a := range m.Agreements.UserAgreementList {
					var ag = a
					if a.Status == StatusRequested {
						if ConsoleText != true {
							pb.Add(1)
						}

						sem <- 1
						go func(i int) {

							DownloadWorker(i, cfg, ga, ma, ag, false)
							<-sem // removes an int from sem, allowing another to proceed
						}(i)

					}

				}

			}
		}

	}

	logFile.Close()
}

// Init download of document
func DownloadWorker(id int, cfg Configuration, group *GroupInfoList, User *UserInfoList, agreement *UserAgreementList, debug bool) {

	if ConsoleText == true {
		println(color.Colorize(color.Yellow, "Worker "+strconv.Itoa(id)+": Received job to download document:"+agreement.Name))
	}

	agreement.DownloadUserAgreement(cfg, group.GroupName, User.ID, User.Email)
	agreement.DownloadAgreementDocuments(cfg, group.GroupName, User.ID, User.Email)

}

// Get Agreement Documents for given userID
func (data *UserAgreementList) GetDocuments(UserID string) {
	GetAgreementDocuments(cfg.Session.AccessToken, cfg.Session.baseURI.APIAccessPoint, &data, UserID, data.ID, true)
}

// Counts the number of agreements for given Status
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

// Checks if Agreement is already downloaded
func (data *UserAgreementList) IsAgreementDownloaded(cfg Configuration, GroupName string, UserID string, UserEmail string) bool {
	name := MakeFilenameWindowsFriendly(data.Name) + "-Combined.pdf"
	//fullPath := cfg.DownloadLocation + "\\" + GroupName + "\\" + UserEmail + "\\" + MakeFilenameWindowsFriendly(data.Name) + "(" + data.ID + ")\\"
	p := filepath.Join(cfg.DownloadLocation, GroupName, UserEmail, MakeFilenameWindowsFriendly(data.Name)+" ("+data.ID+")")
	completePath := filepath.Join(p, name)
	if _, err := os.Stat(completePath); err == nil {
		return true

	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist
		return false

	}
	return false
}

// Returns path to agreement
func (data *UserAgreementList) GetAgreementPath(cfg Configuration, GroupName string, UserID string, UserEmail string) string {
	name := MakeFilenameWindowsFriendly(data.Name) + "-Combined.pdf"
	//fullPath := cfg.DownloadLocation + "\\" + GroupName + "\\" + UserEmail + "\\" + MakeFilenameWindowsFriendly(data.Name) + "(" + data.ID + ")\\"
	p := filepath.Join(cfg.DownloadLocation, GroupName, UserEmail, MakeFilenameWindowsFriendly(data.Name)+" ("+data.ID+")")
	completePath := filepath.Join(p, name)
	if _, err := os.Stat(completePath); err == nil {
		return completePath
	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist
		return ""
	}
	return ""
}

// check if agreement is already downloaded or starts it.
func (data *UserAgreementList) DownloadUserAgreement(cfg Configuration, GroupName string, UserID string, UserEmail string) {
	name := MakeFilenameWindowsFriendly(data.Name) + "-Combined.pdf"
	//fullPath := cfg.DownloadLocation + "\\" + GroupName + "\\" + UserEmail + "\\" + MakeFilenameWindowsFriendly(data.Name) + "(" + data.ID + ")\\"
	p := filepath.Join(cfg.DownloadLocation, GroupName, UserEmail, MakeFilenameWindowsFriendly(data.Name)+" ("+data.ID+")")
	completePath := filepath.Join(p, name)
	if _, err := os.Stat(completePath); err == nil {
		if ConsoleText == true {
			println(color.Colorize(color.Cyan, "File already exists: "+completePath))
		}

	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist
		err := DownloadAgreement(cfg.Session.AccessToken, cfg.Session.baseURI.APIAccessPoint, data.ID, UserID, p, name, debug)

		if err != nil {
			if ConsoleText == true {
				println(color.Colorize(color.Red, "Download Failed: "+completePath))
			}
			return
		} else {
			if ConsoleText == true {
				println(color.Colorize(color.Green, "Download Success: "+name))
			}
		}
	} else {
		// file may or may not exist. See err for details.
		fmt.Println(err)
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence

	}

}

// Download user agreement documents. Checks if they are already downloaded.
func (data *UserAgreementList) DownloadAgreementDocuments(cfg Configuration, GroupName string, UserID string, UserEmail string) {

	for _, document := range data.Documents {
		var completePath string
		var name string
		name = MakeFilenameWindowsFriendly(document.Name)
		//fullPath := cfg.DownloadLocation + "\\" + GroupName + "\\" + UserEmail + "\\" + MakeFilenameWindowsFriendly(data.Name) + "(" + data.ID + ")\\"
		p := filepath.Join(cfg.DownloadLocation, GroupName, UserEmail, MakeFilenameWindowsFriendly(data.Name)+" ("+data.ID+")")
		extension := filepath.Ext(document.Name)
		if extension == "" {
			name = name + ".pdf"
		}
		completePath = filepath.Join(p, name)
		if _, err := os.Stat(completePath); err == nil {
			if ConsoleText == true {
				println(color.Colorize(color.Cyan, "File already exists: "+completePath))
			}

		} else if errors.Is(err, os.ErrNotExist) {
			// path/to/whatever does *not* exist
			err := DownloadDocuments(cfg.Session.AccessToken, cfg.Session.baseURI.APIAccessPoint, data.ID, document.ID, UserID, p, name, debug)

			if err != nil {
				if ConsoleText == true {
					println(color.Colorize(color.Red, "Download Failed: "+completePath))
				}
				return
			} else {
				if ConsoleText == true {
					println(color.Colorize(color.Green, "Download Success: "+name))
				}
			}
		} else {
			// file may or may not exist. See err for details.
			fmt.Println(err)
			// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence

		}
	}

}

// HTTP Get download document
func DownloadDocuments(ACCESSTOKEN string, baseUri string, AgreementID string, DocumentID string, UserID string, filePath string, fileName string, debug bool) error {
	path := filePath
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Println(err)
	}

	// Create the file
	out, err := os.Create(filepath.Join(filePath, fileName))
	if err != nil {
		return err
	}
	defer out.Close()

	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS10, MaxVersion: tls.VersionTLS13},
	}

	url := baseUri + "api/rest/v6/agreements/" + AgreementID + "/documents/" + DocumentID
	if debug == true {
		fmt.Println("URL:>", url)
	}

	for i := 0; i < 3; i++ { // try 3 times
		req, err := http.NewRequest("GET", url, nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-user", "userid:"+UserID)
		req.Header.Add("Authorization", "Bearer "+ACCESSTOKEN)
		client := &http.Client{Transport: tr}
		resp, err := client.Do(req)
		if err != nil {
			// error occurred
			println(color.Colorize(color.Yellow, "\nRetry Download Count: "+strconv.Itoa(i)) + " : " + fileName)
			time.Sleep(time.Second * 10) // wait for 10 seconds
			continue                     // retry request
		}
		defer resp.Body.Close()
		// Writer the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return err
		}

		break // no error or other error
	}

	return nil
}

// HTTP Download Agreement Combined Document
func DownloadAgreement(ACCESSTOKEN string, baseUri string, AgreementID string, UserID string, filePath string, fileName string, debug bool) error {
	path := filePath
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Println(err)
	}

	// Create the file
	out, err := os.Create(filepath.Join(filePath, fileName))
	if err != nil {
		return err
	}
	defer out.Close()

	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS10, MaxVersion: tls.VersionTLS13},
	}

	url := baseUri + "api/rest/v6/agreements/" + AgreementID + "/combinedDocument"
	if debug == true {
		fmt.Println("URL:>", url)
	}
	for i := 0; i < 3; i++ { // try 3 times
		req, err := http.NewRequest("GET", url, nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-user", "userid:"+UserID)
		req.Header.Add("Authorization", "Bearer "+ACCESSTOKEN)
		client := &http.Client{Transport: tr}
		resp, err := client.Do(req)
		if err != nil {
			// error occurred
			println(color.Colorize(color.Yellow, "\nRetry Download Count: "+strconv.Itoa(i)) + " : " + fileName)
			time.Sleep(time.Second * 10) // wait for 10 seconds
			continue                     // retry request
		}
		defer resp.Body.Close()
		// Writer the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return err
		}

		break // no error or other error
	}
	return nil
}

// Query given user for Agreements and saves them to data struct
func (data *UserInfoList) QueryUseAgreement(AccessToken string, URI string) {

	GetUserAgreements(AccessToken, URI, &data.Agreements, data.ID, debug)
	for _, a := range data.Agreements.UserAgreementList {
		fmt.Println("Found Agreement", a.ID, "On user:", data.ID)
	}

}

// Worker for getting Agreements
func GetUserAgreementsWorker(id int, AccessToken string, URI string, members <-chan *UserInfoList, done chan<- bool) {
	for member := range members {

		if debug != true {
			fmt.Println("Worker", id, ": Received job", member.Email)
		}

		member.QueryUseAgreement(AccessToken, URI)

		done <- true
	}

}

// Gets User Documents
func GetUserDocuments(id int, AccessToken string, URI string, MemberID string, agreements <-chan *UserAgreementList, done chan<- bool) {
	for agreement := range agreements {

		if debug != true {
			fmt.Println("Worker", id, ": Received job", MemberID)
		}

		agreement.GetDocuments(MemberID)
		done <- true
	}

}

// HTTP GET: Get Agreement Documents
func GetAgreementDocuments(ACCESSTOKEN string, baseUri string, target interface{}, UserID string, agreementID string, debug bool) error {
	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	url := baseUri + "api/rest/v6/agreements/" + agreementID + "/documents?pageSize=10000"
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

// HTTP GET: Get User Agreements
func GetUserAgreements(ACCESSTOKEN string, baseUri string, target interface{}, UserID string, debug bool) error {
	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
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

// HTTP GET: Query Adobe Sign Endpoint
func (cfg *Configuration) QueryEndpoint() {
	//GetEndpoint(cfg.Session.AccessToken, cfg)

	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
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

// HTTP GET: Query groups in Adobe Sign
func (data *Data) QueryGroups(cfg Configuration) {
	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
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

// Get Group members for given Group
func (data *GroupInfoList) QueryGroupMembers(AccessToken string, URI string) {
	GetGroupMembers(AccessToken, URI, data.GroupID, data.GroupName, &data.GroupMembers, debug)

}

// HTTP GET: Get Group member
func GetGroupMembers(AccessToken string, URI string, GroupID string, GroupName string, target interface{}, debug bool) error {

	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
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

// GO Routine Worker func: Get Group Member
func GetGroupMembersWorker(id int, AccessToken string, URI string, groups <-chan *GroupInfoList, done chan<- bool) {
	for group := range groups {
		if debug != true {
			fmt.Println("Worker", id, ": Received job", group.GroupName)
		}

		group.QueryGroupMembers(AccessToken, URI)
		done <- true
	}

}

// Load config.json
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
