package messaging

func (router *Router) askForRating(usernameToRate string, userId string) error {
	buttons := []ElementButton{
		{
			Type:    "postback",
			Title:   "Positive",
			Payload: "RATING:POSITIVE:" + usernameToRate,
		},
		{
			Type:    "postback",
			Title:   "Negative",
			Payload: "RATING:NEGATIVE:" + usernameToRate,
		},
		{
			Type:    "postback",
			Title:   "Cancel",
			Payload: "CANCEL",
		},
	}

	return router.sendButtonMessage(buttons, "How was your interaction with @"+usernameToRate+"?", userId)
}

func (router *Router) askForUsernameToSearch(userId string) error {
	return router.sendTextMessage("Please enter the username (with '@' symbol) of the page you want to check. (e.g. @purplecheck_org)", userId)
}

func (router *Router) invalidResponseMessage(userId string) error {
	return router.sendTextMessage("Invalid response. Please select one of the options provided. Or click cancel.", userId)
}

func (router *Router) askForRole(userId string) error {
	buttons := []ElementButton{
		{
			Type:    "postback",
			Title:   "Buyer",
			Payload: "ROLE:BUYER",
		},
		{
			Type:    "postback",
			Title:   "Seller",
			Payload: "ROLE:SELLER",
		},
		{
			Type:    "postback",
			Title:   "Cancel",
			Payload: "CANCEL",
		},
	}
	return router.sendButtonMessage(buttons, "What was your role in this interaction?", userId)
}

func (router *Router) askForDealStage(userId string) error {
	buttons := []ElementButton{
		{
			Type:    "postback",
			Title:   "Completed Deal",
			Payload: "DEAL_STAGE:COMPLETE",
		},
		{
			Type:    "postback",
			Title:   "Incomplete Deal",
			Payload: "DEAL_STAGE:INCOMPLETE",
		},
		{
			Type:    "postback",
			Title:   "Cancel",
			Payload: "CANCEL",
		},
	}
	return router.sendButtonMessage(buttons, "What was the stage of the deal?", userId)
}
