package telegram

const msgHelp = `I can save and keep your pages. Also I can offer you read options based on my page storage.

In order to save the page, just send me the link.

In order to get a random page from your list, send me commmand /rnd
Caution! After that page will be removed from the list.`

const msgHello = "Hi! I'm LinkBot_V1 \n\n" + msgHelp

const (
	msgUnknown       = "Unknown command 🤔"
	msgNoSavedPages  = "No saved pages 🥸"
	msgSaved         = "Saved! 🥳"
	msgAlreadyExists = "You already have this page in your list 🤓"
	msgRM            = "Pages removed 🫡"
)
