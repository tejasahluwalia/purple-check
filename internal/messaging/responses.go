package messaging

import "context"

func (router *Router) askForRating(ctx context.Context, usernameToRate string, userId string) error {
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

	return router.sendButtonMessage(ctx, buttons, "How was your interaction with @"+usernameToRate+"?", userId)
}

func (router *Router) askForUsernameToSearch(ctx context.Context, userId string) error {
	return router.sendTextMessage(ctx, "Please enter the username of the page to search \n(e.g. @purplecheck_org)", userId)
}

func (router *Router) invalidResponseMessage(ctx context.Context, userId string) error {
	return router.sendTextMessage(ctx, "Invalid response. Please select one of the options provided. Or click cancel.", userId)
}

func (router *Router) askForRole(ctx context.Context, userId string) error {
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
	return router.sendButtonMessage(ctx, buttons, "What was your role in this interaction?", userId)
}

func (router *Router) askForDealStage(ctx context.Context, userId string) error {
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
	return router.sendButtonMessage(ctx, buttons, "What was the stage of the deal?", userId)
}
