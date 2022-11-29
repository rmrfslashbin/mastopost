package config

import (
	"encoding/gob"
	"os"
	"time"
)

type FilenameRequired struct {
	Err error
}

func (e *FilenameRequired) Error() string {
	if e.Err == nil {
		return "filename required"
	}
	return e.Err.Error()
}

type FileNotExist struct {
	Err      error
	Filename string
	Msg      string
}

func (e *FileNotExist) Error() string {
	if e.Msg == "" {
		e.Msg = "file does not exist"
	}
	if e.Filename != "" {
		e.Msg += ": " + e.Filename
	}
	return e.Msg
}

// FeedConfig is the configuration for a single RSS feed
type FeedConfig struct {
	FeedURL       string     `json:"feedurl"`
	LastUpdated   *time.Time `json:"lastupdated"`
	LastPublished *time.Time `json:"lastpublished"`
}
type LastUpdates struct {
	filename string
	Feeds    map[string]FeedConfig `json:"feeds"`
}

func NewLastUpdates(filename string) (*LastUpdates, error) {
	if filename == "" {
		return nil, &FilenameRequired{}
	}

	// Load the file
	l := &LastUpdates{filename: filename}
	if err := l.Load(); err != nil {
		return nil, err
	}
	return l, nil
}

// loadGOB loads the last update config data
func (l *LastUpdates) Load() error {
	file, err := os.Open(l.filename)
	if err != nil {
		// if the file doesn't exist, return an empty config
		if os.IsNotExist(err) {
			return &FileNotExist{Err: err, Filename: l.filename}
		}
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&l)
	return err
}

// saveGOB saves the last update config data
func (l *LastUpdates) Save() error {
	file, err := os.Create(l.filename)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := gob.NewEncoder(file)
	err = enc.Encode(l)
	if err != nil {
		return err
	}
	return nil
}
