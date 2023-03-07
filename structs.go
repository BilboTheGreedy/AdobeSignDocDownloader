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
	Email        string `json:"email"`
	ID           string `json:"id"`
	IsGroupAdmin bool   `json:"isGroupAdmin"`
	Company      string `json:"company"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Agreements   Agreements
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
}
