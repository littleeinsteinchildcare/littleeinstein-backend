package firebase

import (
	"context"
	"fmt"
	"os"
	"sync"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var (
	app  *firebase.App
	once sync.Once
	err  error
)

func Init() *firebase.App {
	once.Do(func() {
		serviceAccount := os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON")
		if serviceAccount == "" {
			panic("FIREBASE_SERVICE_ACCOUNT_JSON is not set")
		}

		opt := option.WithCredentialsJSON([]byte(serviceAccount))
		app, err = firebase.NewApp(context.Background(), nil, opt)
		if err != nil {
			panic(fmt.Sprintf("error initializing Firebase: %v", err))
		}
	})
	return app
}
