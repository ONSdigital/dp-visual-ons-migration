package zebedee

import (
	"encoding/json"
	"io/ioutil"
	"github.com/satori/go.uuid"
	"fmt"
	"os"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/dp-visual-ons-migration/mapping"
	"regexp"
	"strings"
)

var (
	collectionDirs   = []string{"inprogress", "complete", "reviewed"}
	CollectionsRoot  = ""
	validFilePattern = "[^a-zA-Z0-9]+"
)

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

func CreateCollection(name string) (*Collection, error) {
	r, _ := regexp.Compile(validFilePattern)
	name = strings.ToLower(r.ReplaceAllString(name, ""))

	c := &Collection{
		ApprovalStatus:        "NOT_STARTED",
		CollectionOwner:       "PUBLISHING_SUPPORT",
		IsEncrypted:           false,
		PublishComplete:       false,
		Type:                  "manual",
		ID:                    fmt.Sprintf("%s-%s", name, uuid.NewV4().String()),
		Name:                  name,
		TimeSeriesImportFiles: []string{},
	}

	b, err := json.Marshal(c)
	if err != nil {
		return nil, CollectionError{message: "failed to marshall zebedee json", originalErr: err}
	}

	collectionPath := fmt.Sprintf("%s/%s", CollectionsRoot, c.Name)

	log.Info("creating root zebedee directory", log.Data{"path": collectionPath})
	if err := os.Mkdir(collectionPath, 0755); err != nil {
		return nil, CollectionError{"failed to created zebedee root dir", err}
	}

	for _, dir := range collectionDirs {
		path := fmt.Sprintf("%s/%s", collectionPath, dir)
		log.Info("creating zebedee sub directory", log.Data{"path": collectionPath, "dir": dir})

		if err := os.Mkdir(path, 0755); err != nil {
			return nil, CollectionError{"failed to created zebedee sub dir", err}
		}
	}

	log.Info("creating zebedee json file", log.Data{"path": collectionPath, "filename": collectionPath + ".json"})

	if err := writeToFile(collectionPath+".json", b); err != nil {
		return nil, CollectionError{"failed to write zebedee json file", err}
	}

	return c, nil
}

func (e CollectionError) Error() string {
	return e.message + " " + e.originalErr.Error()
}

func (c Collection) ResolveInProgress(path string) string {
	return CollectionsRoot + "/" + c.Name + "/inprogress" + path
}

func (c Collection) AddArticle(a *Article, m *mapping.MigrationDetails) error {
	if err := os.MkdirAll(c.ResolveInProgress(m.GetONSURI()), 0755); err != nil {
		return err
	}

	b, _ := json.MarshalIndent(a, "", "	")
	return writeToFile(c.ResolveInProgress(m.GetONSURI()+"/data.json"), b)
}

func writeToFile(path string, b []byte) error {
	if err := ioutil.WriteFile(path, b, 0644); err != nil {
		return CollectionError{"failed to write zebedee json to file", err}
	}
	return nil
}
