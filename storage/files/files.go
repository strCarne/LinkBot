package files

import (
	"LinkBot/lib/e"
	"LinkBot/storage"
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type Storage struct {
	basePath string
}

const (
	defaultPerm = 0774
)

func New(basePath string) Storage {
	return Storage{
		basePath: basePath,
	}
}

func (s Storage) Save(p *storage.Page) (err error) {
	defer func() { err = e.WrapIfErr("can't save the page", err) }()

	fPath := filepath.Join(s.basePath, p.UserName)

	if os.MkdirAll(fPath, defaultPerm); err != nil {
		return err
	}

	fName, err := fileName(p)

	if err != nil {
		return err
	}

	fPath = filepath.Join(fPath, fName)

	file, err := os.Create(fPath)

	defer func() { _ = file.Close() }()

	err = gob.NewEncoder(file).Encode(p)
	if err != nil {
		return err
	}

	return nil
}

func fileName(p *storage.Page) (string, error) {
	return p.Hash()
}

func (s Storage) PickRandom(userName string) (p *storage.Page, err error) {
	defer func() { err = e.WrapIfErr("can't pick random page", err) }()

	path := filepath.Join(s.basePath, userName)
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	rand.Seed(time.Now().UnixNano())

	n := rand.Intn(len(files))

	file := files[n]

	return s.decodePage(filepath.Join(path, file.Name()))
}

func (s Storage) List(username string) (p []*storage.Page, err error) {
	defer func() { err = e.WrapIfErr("can't make page list", err) }()

	path := filepath.Join(s.basePath, username)
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	p = make([]*storage.Page, 0, len(files))

	for _, f := range files {
		page, err := s.decodePage(filepath.Join(path, f.Name()))
		if err != nil {
			return nil, err
		}
		p = append(p, page)
	}

	return p, nil
}

func (s Storage) rmPage(username string, indexes []int) error {
	path := filepath.Join(s.basePath, username)
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for i := len(indexes) - 1; i >= 0; i-- {
		fileToRM := filepath.Join(path, files[i].Name())
		if err = os.Remove(fileToRM); err != nil {
			return e.Wrap(fmt.Sprintf("can't remove file [%v]", fileToRM), err)
		}
	}
	return nil
}

func (s Storage) decodePage(filePath string) (*storage.Page, error) {
	f, err := os.Open(filePath)

	if err != nil {
		return nil, e.Wrap("can't decode the page", err)
	}

	defer func() { _ = f.Close() }()

	var p storage.Page

	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		return nil, e.Wrap("can't decode the page", err)
	}

	return &p, nil
}

func (s Storage) Remove(p *storage.Page) (err error) {
	fileName, err := fileName(p)
	if err != nil {
		return e.Wrap("can't remove the page", err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	if err = os.Remove(path); err != nil {
		msg := fmt.Sprintf("can't remove the page %s", path)
		return e.Wrap(msg, err)
	}

	return nil
}

func (s Storage) IsExists(p *storage.Page) (bool, error) {
	fileName, err := fileName(p)
	if err != nil {
		return false, e.Wrap("can't check if file exists", err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	switch _, err = os.Stat(path); {
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	case err != nil:
		msg := fmt.Sprintf("can't check if file %s exists", path)
		return false, e.Wrap(msg, err)
	}

	return true, nil
}
