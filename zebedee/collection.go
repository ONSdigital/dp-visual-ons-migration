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
	"errors"
)

const (
	inProgress      = "inprogress"
	complete        = "complete"
	reviewed        = "reviewed"
	dataJSON        = "data.json"
	approvalStatus  = "NOT_STARTED"
	collectionOwner = "PUBLISHING_SUPPORT"
	collectionType  = "manual"
)

var (
	collectionDirs     = []string{inProgress, complete, reviewed}
	CollectionsRoot    = ""
	validFilePattern   = "[^a-zA-Z0-9]+"
	validFileNameRegex *regexp.Regexp
)

func init() {
	var err error
	validFileNameRegex, err = regexp.Compile(validFilePattern)
	if err != nil {
		panic(errors.New("failed to compile valid filename regex"))
		os.Exit(1)
	}
}

type CollectionMetadata struct {
	Root           string
	CollectionJSON string
	InProgress     string
	Complete       string
	Reviewed       string
	DataJSON       string
}

type Collection struct {
	Metadata              *CollectionMetadata `json:"-"`
	ApprovalStatus        string              `json:"approvalStatus"`
	PublishComplete       bool                `json:"publishComplete"`
	IsEncrypted           bool                `json:"isEncrypted"`
	CollectionOwner       string              `json:"collectionOwner"`
	TimeSeriesImportFiles []string            `json:"timeseriesImportFiles"`
	ID                    string              `json:"id"`
	Name                  string              `json:"name"`
	Type                  string              `json:"type"`
}

type CollectionError struct {
	message     string
	originalErr error
}

func CreateCollection(name string) (*Collection, error) {
	name = strings.ToLower(validFileNameRegex.ReplaceAllString(name, ""))
	name = "visual_" + name

	collectionRootDir := fmt.Sprintf("%s/%s", CollectionsRoot, name)

	metadata := &CollectionMetadata{
		Root:           collectionRootDir,
		CollectionJSON: collectionRootDir + ".json",
		InProgress:     collectionRootDir + "/" + inProgress,
		Complete:       collectionRootDir + "/" + complete,
		Reviewed:       collectionRootDir + "/" + reviewed,
		DataJSON:       collectionRootDir + "/" + dataJSON,
	}

	c := &Collection{
		Metadata:              metadata,
		ApprovalStatus:        approvalStatus,
		CollectionOwner:       collectionOwner,
		IsEncrypted:           false,
		PublishComplete:       false,
		Type:                  collectionType,
		ID:                    newCollectionID(name),
		Name:                  name,
		TimeSeriesImportFiles: []string{},
	}

	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return nil, CollectionError{message: "failed to marshall zebedee json", originalErr: err}
	}

	log.Info("creating Root zebedee directory", log.Data{"path": c.Metadata.Root})
	if err := os.Mkdir(c.Metadata.Root, 0755); err != nil {
		return nil, CollectionError{"failed to created collection Root dir", err}
	}

	if err := os.Mkdir(c.Metadata.InProgress, 0755); err != nil {
		return nil, CollectionError{"failed to created collection InProgress dir", err}
	}

	if err := os.Mkdir(c.Metadata.Complete, 0755); err != nil {
		return nil, CollectionError{"failed to created collection Complete dir", err}
	}

	if err := os.Mkdir(c.Metadata.Reviewed, 0755); err != nil {
		return nil, CollectionError{"failed to created collection Reviewed dir", err}
	}

	if err := writeToFile(c.Metadata.CollectionJSON, b); err != nil {
		return nil, CollectionError{"failed to write collection json file", err}
	}

	return c, nil
}

func (e CollectionError) Error() string {
	return e.message + " " + e.originalErr.Error()
}

func (c Collection) ResolveInProgress(path string) string {
	return c.Metadata.InProgress + path
}

func (c Collection) AddArticle(article *Article, migrationDetails *mapping.MigrationDetails) error {
	path := c.Metadata.InProgress + migrationDetails.GetTaxonomyURI()
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	b, _ := json.MarshalIndent(article, "", "	")
	return writeToFile(path+"/"+dataJSON, b)
}

func newCollectionID(collectionName string) string {
	return fmt.Sprintf("%s-%s", collectionName, uuid.NewV4().String())
}

func writeToFile(path string, b []byte) error {
	if err := ioutil.WriteFile(path, b, 0644); err != nil {
		return CollectionError{"failed to write json file", err}
	}
	return nil
}
