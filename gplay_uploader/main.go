package main

import (
	"context"

	"google.golang.org/api/androidpublisher/v3"

	//"google.golang.org/api/oauth2/v2"

	"google.golang.org/api/option"
)

// https://godoc.org/golang.org/x/oauth2/google

func main() {
	ctx := context.Background()

	/*jsonData := []byte{}
	tokenSrc, err := google.JWTAccessTokenSourceFromJSON(jsonData, "asd")
	if err != nil {
		return
	}*/

	/*token, err := tokenSrc.Token()

	config := &oauth2.Config{}
	config.Scopes = []string{"https://www.googleapis.com/auth/androidpublisher"}
	config.TokenSource(ctx, token)
	token, err := config.Exchange(ctx)*/

	// option.WithTokenSource(config.TokenSource(ctx, token))
	jsonData := []byte{}
	androidpublisherService, err := androidpublisher.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return
	}
	//revs := androidpublisherService.Reviews.List("")
	//apks := androidpublisherService.Edits.Apks.List("", "edit key")

	get := androidpublisherService.Edits.Get("asdsad", "asdasd")
	upload := androidpublisherService.Edits.Apks.Upload("asd", "asda")
	CallOptions
	apk, err := upload.Do()

	list := androidpublisherService.Edits.Bundles.List("asd", "asd")
	res, err := list.Do()
}
