package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"os"

	"github.com/TwiN/go-color"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	scribble "github.com/nanobox-io/golang-scribble"
)

type DocumentStatus struct {
	//Name   string `json:"name"`
	User   string `json:"User"`
	Status string `json:"status"`
	Count  int
}

type ADInfo struct {
	Userprincipalname string `json:"userprincipalname"`
	Country           string `json:"co"`
	City              string `json:"city"`
}

func GenerateStatusTable() {
	println(color.Colorize(color.Cyan, "Generate Status Table"))
	db, _ := scribble.New("./Data", nil)
	// Read from the cache database
	groups, _ := db.ReadAll("Groups")
	// Map of maps to store counts for each user and status
	counts := make(map[string]map[string]int)
	// iterate
	c_data := Data{}
	for _, group := range groups {
		f := GroupInfoList{}
		json.Unmarshal([]byte(group), &f)
		c_data.Groups.GroupInfoList = append(c_data.Groups.GroupInfoList, &f)
	}

	// Loop over slice of structs
	for _, group := range c_data.Groups.GroupInfoList {
		for _, member := range group.GroupMembers.UserInfoList {
			for _, agreement := range member.Agreements.UserAgreementList {

				// Get the user and status from each struct
				user := member.Email
				status := agreement.Status

				// Check if user exists in counts map
				if _, ok := counts[user]; !ok {
					// If not, create a new map for that user
					counts[user] = make(map[string]int)
				}

				// Increment the count for that user and status combination
				counts[user][status]++
			}
		}

	}
	v_data := []DocumentStatus{}
	// Print the counts for each user and status
	for user, statuses := range counts {
		for status, count := range statuses {
			v_data = append(v_data, DocumentStatus{user, status, count})
		}
	}
	t := template.Must(template.ParseFiles("templates/status.html", "templates/Links.html", "templates/sources.html"))

	f, err := os.Create("status.html")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	t.Execute(f, v_data)
	return
}

func VerifyPaths(StatusRequested string) {
	println(color.Colorize(color.Cyan, "Verify Paths"))
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
	println(color.Colorize(color.Cyan, "Generate Verify Table"))
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

	t := template.Must(template.ParseFiles("templates/table.html", "templates/Links.html", "templates/sources.html"))

	f, err := os.Create("table.html")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	t.Execute(f, data)
	return
}

func GetADInfo() *charts.Map {
	data := make([]ADInfo, 0)
	// Read JSON data from file
	jsonData, err := ioutil.ReadFile("./Data/AdInfo.json")
	if err != nil {
		panic(err)
	}
	// Unmarshal JSON data into slice of Data structs
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		panic(err)
	}

	println(color.Colorize(color.Cyan, "Generate Status by Country"))
	db, _ := scribble.New("./Data", nil)
	// Read from the cache database
	groups, _ := db.ReadAll("Groups")
	// Map of maps to store counts for each user and status
	counts := make(map[string]int)
	// iterate
	c_data := Data{}
	for _, group := range groups {
		f := GroupInfoList{}
		json.Unmarshal([]byte(group), &f)
		c_data.Groups.GroupInfoList = append(c_data.Groups.GroupInfoList, &f)
	}

	for _, item := range data {
		var country string
		var upn string
		country = item.Country
		upn = item.Userprincipalname
		if _, ok := counts[item.Country]; !ok {
			// If not, create a new map for that country
			counts[country] = 0
		}
		for _, group := range c_data.Groups.GroupInfoList {
			for _, member := range group.GroupMembers.UserInfoList {
				for i := 0; i < len(member.Agreements.UserAgreementList); i++ {
					if member.Email == upn {
						if _, ok := counts[item.Country]; ok {
							// If not, create a new map for that country
							counts[country]++
						}
					}

				}
			}
		}
	}
	mapData := []opts.MapData{}
	for country, count := range counts {
		mapData = append(mapData, opts.MapData{Name: country, Value: count})
	}

	mc := charts.NewMap()

	mc.RegisterMapType("world")
	//mc.Assets.JSAssets.Add("maps/sweden.js")
	mc.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Epiroc - Agreement owner by Country"}),
		charts.WithVisualMapOpts(opts.VisualMap{
			Show:       true,
			Calculable: true,
			Type:       "piecewise",
			InRange:    &opts.VisualMapInRange{Color: []string{"#50a3ba", "#eac736", "#d94e5d"}},
			Range:      []float32{20000},
			Text:       []string{"Many", "Few"},
			Max:        20000,
			Min:        0,
		}),
	)
	mc.SetGlobalOptions(charts.WithTooltipOpts(opts.Tooltip{Show: true}),
		charts.WithParallelComponentOpts(opts.ParallelComponent{Top: "20%", Bottom: "20%"}))
	mc.AddSeries("map", mapData)
	return mc
}

func GetPieData() *charts.Pie {
	println(color.Colorize(color.Cyan, "Generate Status Pie"))
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

	statusCount := make(map[string]int)
	for _, group := range c_data.Groups.GroupInfoList {
		for _, userInfo := range group.GroupMembers.UserInfoList {
			for _, agreement := range userInfo.Agreements.UserAgreementList {
				statusCount[agreement.Status]++
			}
		}
	}
	var pdata []opts.PieData
	for status, count := range statusCount {
		pdata = append(pdata, opts.PieData{Name: status, Value: count})
	}
	pie := charts.NewPie()

	pie.SetGlobalOptions(
		charts.WithLegendOpts(opts.Legend{
			Show: true,
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true}),
		charts.WithParallelComponentOpts(opts.ParallelComponent{Top: "100%", Bottom: "100%"}),
	)

	pie.AddSeries("Agreement Status", pdata).
		SetSeriesOptions(charts.WithLabelOpts(
			opts.Label{
				Show:      true,
				Formatter: "{b}: {c}",
			}),
			charts.WithPieChartOpts(opts.PieChart{
				Radius:   []string{"30%", "75%"},
				RoseType: "radius",
				Center:   []string{"50%", "60%"},
			}),
		)
	return pie
}
