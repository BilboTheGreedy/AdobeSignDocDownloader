package main

import "time"

const (
	Signed              string = "SIGNED"
	Cancelled           string = "CANCELLED"
	OutForSignature     string = "OUT_FOR_SIGNATURE"
	Expired             string = "EXPIRED"
	WaitingForAuthoring string = "WAITING_FOR_AUTHORING"
)

type Session struct {
	AccessToken string
	baseURI     baseURI
}

type baseURI struct {
	APIAccessPoint string `json:"apiAccessPoint"`
	WebAccessPoint string `json:"webAccessPoint"`
}

type Configuration struct {
	DownloadLocation string
	Session          Session
}

type Data struct {
	Groups Groups
}

type Groups struct {
	GroupInfoList []*GroupInfoList `json:"groupInfoList"`
}

type GroupInfoList struct {
	GroupID        string    `json:"groupId"`
	CreatedDate    time.Time `json:"createdDate"`
	GroupName      string    `json:"groupName"`
	IsDefaultGroup bool      `json:"isDefaultGroup"`
	GroupMembers   GroupMembers
}

type GroupMembers struct {
	UserInfoList []*UserInfoList `json:"userInfoList"`
}

type UserInfoList struct {
	Email          string `json:"email"`
	ID             string `json:"id"`
	PrimaryGroupID string
	IsGroupAdmin   bool   `json:"isGroupAdmin"`
	Company        string `json:"company"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	Agreements     Agreements
}

type Agreements struct {
	UserAgreementList []*UserAgreementList `json:"userAgreementList"`
}

type UserAgreementList struct {
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
	Documents       []struct {
		CreatedDate string `json:"createdDate"`
		ID          string `json:"id"`
		Label       string `json:"label"`
		NumPages    int    `json:"numPages"`
		MimeType    string `json:"mimeType"`
		Name        string `json:"name"`
	} `json:"documents"`
	SupportingDocuments []struct {
		DisplayLabel  string `json:"displayLabel"`
		FieldName     string `json:"fieldName"`
		ID            string `json:"id"`
		MimeType      string `json:"mimeType"`
		NumPages      int    `json:"numPages"`
		ParticipantID string `json:"participantId"`
	} `json:"supportingDocuments"`
}

type GeoData struct {
	Email           string `json:"Email"`
	Groups          string `json:"Groups"`
	AgreementsCount string `json:"AgreementsCount"`
	Country         string `json:"Country"`
	City            string `json:"City"`
	Coordinates     struct {
		Value []float64 `json:"value"`
		Count int       `json:"Count"`
	} `json:"Coordinates"`
}

type GeoCount struct {
	Name            string
	TotalAgreements int
}

type GeoMap struct {
	Country []struct {
		Name            string
		Coordinates     []float64
		CountAgreements int
		City            []struct {
			Name            string
			Coordinates     []float64
			CountAgreements int
		}
	}
}

type Entry struct {
	AgreementID     string
	Name            string
	AgreementStatus string
	AgreementOwner  string
	Group           string
	IsDownloaded    string
	FilePath        string
}

type LineEntry [][]string

type Result struct {
	GroupInfoList []struct {
		CreatedDate    string `json:"createdDate"`
		ID             string `json:"id"`
		IsGroupAdmin   bool   `json:"isGroupAdmin"`
		IsPrimaryGroup bool   `json:"isPrimaryGroup"`
		Name           string `json:"name"`
		Settings       struct {
			LibaryDocumentCreationVisible struct {
				Inherited bool `json:"inherited"`
				Value     bool `json:"value"`
			} `json:"libaryDocumentCreationVisible"`
			SendRestrictedToWorkflows struct {
				Inherited bool `json:"inherited"`
				Value     bool `json:"value"`
			} `json:"sendRestrictedToWorkflows"`
			UserCanSend struct {
				Inherited bool `json:"inherited"`
				Value     bool `json:"value"`
			} `json:"userCanSend"`
			WidgetCreationVisible struct {
				Inherited bool `json:"inherited"`
				Value     bool `json:"value"`
			} `json:"widgetCreationVisible"`
		} `json:"settings"`
		Status string `json:"status"`
	} `json:"groupInfoList"`
}
