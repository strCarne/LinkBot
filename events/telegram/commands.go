package telegram

import (
	"LinkBot/lib/e"
	"LinkBot/storage"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	RndCmd    = "/rnd"
	HelpCmd   = "/help"
	StartCmd  = "/start"
	ListCmd   = "/list"
	RemoveCmd = "/rm"
)

func (p *Processor) doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	log.Printf("got new command '%s' from '%s'", text, username)

	if isAddCmd(text) {
		return p.savePage(text, chatID, username)
	}

	canRemove, err := checkRemoveCondition(text)
	if err != nil {
		return err
	}

	if canRemove {
		indexes := parseRemovalList(text)
		return p.removePages(chatID, username, indexes)
	}

	switch text {
	case RndCmd:
		return p.sendRandom(chatID, username) // HERE
	case HelpCmd:
		return p.sendHelp(chatID)
	case StartCmd:
		return p.sendHello(chatID)
	case ListCmd:
		return p.sendList(chatID, username)
	case RemoveCmd:
		return nil
	default:
		return p.tg.SendMessages(chatID, msgUnknown)
	}
}

func parseRemovalList(text string) []int {
	strSlice := strings.Split(text[len(RemoveCmd)+1:], " ")
	res := make([]int, 0, len(strSlice))
	for _, el := range strSlice {
		tmp, _ := strconv.Atoi(el)
		tmp--
		res = append(res, tmp)
	}
	return res
}

func (p *Processor) removePages(chatID int, username string, indexes []int) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: can't remove pages", err) }()

	pages, err := p.storage.List(username)
	if err != nil {
		return err
	}

	sort.Ints(indexes)

	for i := len(indexes) - 1; i >= 0; i-- {
		if err = p.storage.Remove(pages[indexes[i]]); err != nil {
			log.Println(err)
		}
	}

	p.tg.SendMessages(chatID, msgRM)

	return nil
}

func (p *Processor) sendList(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: can't send list", err) }()

	pages, err := p.storage.List(username)
	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}

	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessages(chatID, msgNoSavedPages)
	}

	list := "1. " + pages[0].URL
	for i, p := range pages {
		if i == 0 {
			continue
		}
		list += fmt.Sprintf("\n%d. ", i+1)
		list += p.URL
	}

	if err := p.tg.SendMessages(chatID, list); err != nil {
		return err
	}

	return nil
}

func checkRemoveCondition(text string) (ok bool, err error) {
	defer func() { err = e.WrapIfErr("can't parse expression", err) }()
	exp, err := regexp.Compile("\\" + RemoveCmd + "( \\d)+")
	if err != nil {
		return false, err
	}
	return exp.MatchString(text), nil
}

func (p *Processor) savePage(pageURL string, chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: save page", err) }()

	page := &storage.Page{
		URL:      pageURL,
		UserName: username,
	}

	isExists, err := p.storage.IsExists(page)

	if err != nil {
		return err
	}

	if isExists {
		return p.tg.SendMessages(chatID, msgAlreadyExists)
	}

	if err = p.storage.Save(page); err != nil {
		return err
	}

	if err = p.tg.SendMessages(chatID, msgSaved); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendRandom(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: can't send random", err) }()

	page, err := p.storage.PickRandom(username)
	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}

	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessages(chatID, msgNoSavedPages)
	}

	if err := p.tg.SendMessages(chatID, page.URL); err != nil {
		return err
	}

	return p.storage.Remove(page)
}

func (p *Processor) sendHelp(chatID int) error {
	return p.tg.SendMessages(chatID, msgHelp)
}

func (p *Processor) sendHello(chatID int) error {
	return p.tg.SendMessages(chatID, msgHello)
}

func isAddCmd(text string) bool {
	return isURL(text)
}

func isURL(text string) bool {
	u, err := url.Parse(text)
	return err == nil && u.Host != ""
}
