package collections

import (
	"encoding/json"
	"io/ioutil"
	"github.com/satori/go.uuid"
	"fmt"
	"os"
	"github.com/ONSdigital/go-ns/log"
)

var collectionDirs = []string{"inprogress", "complete", "reviewed"}

type Collection struct {
	ApprovalStatus        string   `json:"approvalStatus"`
	PublishComplete       bool     `json:"publishComplete"`
	IsEncrypted           bool     `json:"isEncrypted"`
	CollectionOwner       string   `json:"collectionOwner"`
	TimeSeriesImportFiles []string `json:"timeseriesImportFiles"`
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	Type                  string   `json:"type"`
}

type CollectionError struct {
	message     string
	originalErr error
}

func New(name string) Collection {
	return Collection{
		ApprovalStatus:        "NOT_STARTED",
		CollectionOwner:       "PUBLISHING_SUPPORT",
		IsEncrypted:           false,
		PublishComplete:       false,
		Type:                  "manual",
		ID:                    fmt.Sprintf("%s-%s", name, uuid.NewV4().String()),
		Name:                  name,
		TimeSeriesImportFiles: []string{},
	}
}

func (e CollectionError) Error() string {
	return e.message + " " + e.originalErr.Error()
}

func (c Collection) Create(path string) error {
	b, err := json.Marshal(c)
	if err != nil {
		return CollectionError{message: "failed to marshall collection json", originalErr: err}
	}

	collectionPath := fmt.Sprintf("%s/%s", path, c.Name)
	log.Info("creating root collection directory", log.Data{
		"path": collectionPath,
	})
	if err := os.Mkdir(collectionPath, 0700); err != nil {
		return CollectionError{"failed to created collection root dir", err}
	}

	for _, dir := range collectionDirs {
		path := fmt.Sprintf("%s/%s", collectionPath, dir)
		log.Info("creating collection sub directory", log.Data{
			"path": collectionPath,
			"dir":  dir,
		})
		if err := os.Mkdir(path, 0700); err != nil {
			return CollectionError{"failed to created collection sub dir", err}
		}
	}

	log.Info("creating collection json file", log.Data{
		"path":     collectionPath,
		"filename": collectionPath + ".json",
	})
	if err := ioutil.WriteFile(collectionPath+".json", b, 0700); err != nil {
		return CollectionError{"failed to write collection json to file", err}
	}

	return nil
}
