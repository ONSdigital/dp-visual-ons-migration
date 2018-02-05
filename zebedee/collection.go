package zebedee

import (
	"encoding/json"
	"io/ioutil"
	"github.com/satori/go.uuid"
	"fmt"
	"os"
	"github.com/ONSdigital/go-ns/log"
	"regexp"
	"strings"
	"errors"
	"github.com/ONSdigital/dp-visual-ons-migration/migration"
)

const (
	inProgress       = "inprogress"
	complete         = "complete"
	reviewed         = "reviewed"
	dataJSON         = "data.json"
	approvalStatus   = "NOT_STARTED"
	collectionOwner  = "PUBLISHING_SUPPORT"
	collectionType   = "manual"
	articleURIFormat = "%s%s/articles/%s"
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
		return nil, migration.Error{Message: "failed to marshall zebedee json", OriginalErr: err, Params: nil}
	}

	for _, path := range []string{c.Metadata.Root, c.Metadata.InProgress, c.Metadata.Complete, c.Metadata.Reviewed} {
		log.Info("creating collection directory", log.Data{"path": path})

		if err := os.Mkdir(path, 0755); err != nil {
			return nil, migration.Error{
				Message:     "failed to created collection dir",
				OriginalErr: err,
				Params:      log.Data{"path": path},
			}
		}
	}

	log.Info("creating collection json file", log.Data{"path": c.Metadata.CollectionJSON})
	if err := writeToFile(c.Metadata.CollectionJSON, b); err != nil {
		return nil, migration.Error{
			Message:     "failed to write collection json file",
			OriginalErr: err,
			Params:      log.Data{"path": c.Metadata.CollectionJSON},
		}
	}

	log.Debug("collection created successfully", log.Data{"path": c.Metadata.Root})
	return c, nil
}

func (c Collection) ResolveInProgress(path string) string {
	return c.Metadata.InProgress + path
}

func (c Collection) AddArticle(zebedeeArticle *Article, visualArticle *migration.Article) error {
	dir := sanitisedFilename(visualArticle.PostTitle)
	path := fmt.Sprintf(articleURIFormat, c.Metadata.InProgress, visualArticle.TaxonomyURI, dir)

	log.Info("making article directories", log.Data{"collection": c.Name, "path": path})
	if err := os.MkdirAll(path, 0755); err != nil {
		return migration.Error{
			Message:     "error making article directories",
			OriginalErr: err,
			Params:      log.Data{"collection": c.Name, "path": path},
		}
	}

	b, err := json.MarshalIndent(zebedeeArticle, "", "	")
	if err != nil {
		return err
	}
	path = path + "/" + dataJSON
	log.Info("writing zebedeeArticle json to collection", log.Data{"collection": c.Name, "path": path})
	if err := writeToFile(path, b); err != nil {
		return migration.Error{
			Message:     "failed to write article json",
			OriginalErr: err,
			Params:      log.Data{"collection": c.Name, "path": path},
		}
	}
	log.Debug("zebedeeArticle created successfully", log.Data{"collection": c.Name, "path": path})
	return nil
}

func newCollectionID(collectionName string) string {
	return fmt.Sprintf("%s-%s", collectionName, uuid.NewV4().String())
}

func writeToFile(path string, b []byte) error {
	if err := ioutil.WriteFile(path, b, 0644); err != nil {
		return migration.Error{
			Message:     "failed to write json file",
			OriginalErr: err,
			Params:      log.Data{"path": path},
		}
	}
	return nil
}

func sanitisedFilename(name string) string {
	return strings.ToLower(validFileNameRegex.ReplaceAllString(name, ""))
}
