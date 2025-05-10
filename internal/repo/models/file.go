package models

import "time"

type File struct {
	ID             int64
	UploaderUserID int64
	Name           string
	ContentType    string
	Bucket         string
	Key            string
	Size           int64
	UploadedAt     time.Time
}
