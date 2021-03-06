package main

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DB has methods to operate JSON files.
type DB struct {
	path DBPath
}

func newDB() (DB, error) {
	dbPath, err := newDBPath()
	if err != nil {
		return DB{}, err
	}

	return DB{
		path: dbPath,
	}, nil
}

// All returns slice of path of all task's JSON files.
func (db DB) All() TaskList {
	return db.collect(db.path.Root())
}

// Finished returns slice of path of finished task's JSON files.
func (db DB) Finished() TaskList {
	return db.collect(db.path.Finished())
}

// Ongoing returns slice of path of ongoing task's JSON files.
func (db DB) Ongoing() TaskList {
	return db.collect(db.path.Ongoing())
}

func (db DB) readFrom(path string) (Task, error) {
	f, err := os.Open(path)
	if err != nil {
		return Task{}, err
	}
	defer f.Close()

	var task Task
	if err := json.NewDecoder(f).Decode(&task); err != nil {
		return Task{}, err
	}
	return task, nil
}

// collect lists all file paths under the rootpath.
func (db DB) collect(rootpath string) TaskList {
	var result TaskList
	err := filepath.Walk(rootpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		result = append(result, path)
		return nil
	})

	if err != nil {
		fmt.Println("Walk", err)
	}
	return result
}

func (db DB) Store(task Task) error {
	return db.createOf(task, ongoing)
}

func (db DB) Finish(title string) error {
	task, err := db.readFrom(db.calcFilePath(title, ongoing))
	if err != nil {
		return err
	}

	err = db.deleteOf(title, ongoing)
	if err != nil {
		return err
	}

	task.FinishedAt = time.Now()

	return db.createOf(task, finished)
}

func (db DB) deleteOf(title string, state TaskState) error {
	path := db.calcFilePath(title, state)
	return os.Remove(path)
}

func (db DB) calcFilePath(title string, state TaskState) string {
	filename := db.calcFileName(title)
	return filepath.Join(db.stateDirPath(state), filename)
}

func (db DB) calcFileName(title string) string {
	id := fmt.Sprintf("%x", sha512.Sum512([]byte(title)))[:10]
	return fmt.Sprintf("%s.json", id)
}

func (db DB) createOf(task Task, state TaskState) error {
	fout, err := os.Create(db.calcFilePath(task.Title, ongoing))
	if err != nil {
		return err
	}
	defer fout.Close()

	return json.NewEncoder(fout).Encode(&task)
	// TaskPrint(filepath.Join(path, data.Title+".json"))
}

func (db DB) stateDirPath(state TaskState) string {
	if state == ongoing {
		return db.path.Ongoing()
	} else if state == finished {
		return db.path.Finished()
	}
	return ""
}
