package m365

import "time"

// Site represents a SharePoint site.
type Site struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	WebURL      string `json:"webUrl"`
}

// Drive represents a document library or OneDrive.
type Drive struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	DriveType string `json:"driveType"` // personal, business, documentLibrary
	WebURL    string `json:"webUrl"`
}

// DriveItem represents a file or folder.
type DriveItem struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	WebURL       string    `json:"webUrl"`
	CreatedBy    Identity  `json:"createdBy"`
	LastModified time.Time `json:"lastModifiedDateTime"`
	File         *File     `json:"file"`   // Nil if folder
	Folder       *Folder   `json:"folder"` // Nil if file
}

type File struct {
	MimeType string `json:"mimeType"`
}

type Folder struct {
	ChildCount int `json:"childCount"`
}

type Identity struct {
	User struct {
		DisplayName string `json:"displayName"`
		Email       string `json:"email"`
	} `json:"user"`
}
