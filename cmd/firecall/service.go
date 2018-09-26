package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/seibert-media/events/pkg/api"
	"bitbucket.org/seibert-media/events/pkg/service"

	"firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/dgraph-io/badger"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/seibert-media/golibs/log"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

const (
	svcName = "FireCall"
	svcKey  = "firecall"
)

// Spec for the service
type Spec struct {
	service.BaseSpec

	ServiceAccount string `envconfig:"google_application_credentials" required:"true" help:"path to google service account file"`
}

// DB is the local file based database
var DB *badger.DB

func main() {
	var svc Spec
	ctx := service.Init(svcKey, svcName, &svc)
	defer service.Defer(ctx)

	err := PrepareDB(ctx, svc)
	if err != nil {
		log.From(ctx).Fatal("initializing database", zap.Error(err))
	}
	defer DB.Close()

	srv := api.New(":8080", svc.Debug)
	Routes(ctx, svc, srv)
	go srv.GracefulHandler(ctx)

	err = srv.Start(ctx)
	if err != nil {
		log.From(ctx).Fatal("running server", zap.Error(err))
	}

	log.From(ctx).Info("finished")
}

// Routes for this service
func Routes(ctx context.Context, svc Spec, srv *api.Server) {
	srv.Router.Route("/dial", func(r chi.Router) {
		r.Post("/register", api.NewHandler(ctx, RegisterHandler(ctx, svc)))
		r.Get("/call/{number}", api.NewHandler(ctx, DialHandler(ctx, svc)))
	})
}

// RegisterHandler storing new devices and returning their API Key
func RegisterHandler(ctx context.Context, svc Spec) api.HandlerFunc {
	type Request struct {
		Device string `json:"device"`
	}

	type Response struct {
		*api.Error
		APIKey string `json:"apiKey"`
	}

	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) api.Response {
		var (
			resp = &Response{Error: &api.Error{}}
			data *Request
		)

		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			resp.Fail(errors.Wrap(err, "decoding request"))
			return resp
		}

		u, err := uuid.NewRandom()
		if err != nil {
			resp.Fail(errors.Wrap(err, "generating uuid"))
			return resp
		}
		h := sha256.New()
		uuid, err := u.MarshalText()
		if err != nil {
			resp.Fail(errors.Wrap(err, "encoding uuid"))
			return resp
		}
		_, err = h.Write(append(uuid, []byte(data.Device)...))
		if err != nil {
			resp.Fail(errors.Wrap(err, "generating api key"))
			return resp
		}

		apiKey := []byte(hex.EncodeToString(h.Sum(nil)))

		log.From(ctx).Info("storing api key", zap.String("apiKey", string(apiKey)))
		err = DB.Update(func(txn *badger.Txn) error {
			return txn.Set(apiKey, []byte(data.Device))
		})

		if err != nil {
			resp.Fail(errors.Wrap(err, "storing api key"))
			return resp
		}

		resp.APIKey = string(apiKey)

		return resp
	}
}

// DialHandler sends a dial message for the number based on the api key stored in the auth cookie
func DialHandler(ctx context.Context, svc Spec) api.HandlerFunc {
	type Response struct {
		*api.Error
	}

	opt := option.WithCredentialsFile(svc.ServiceAccount)
	fb, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.From(ctx).Fatal("creating firebase client", zap.Error(err))
	}

	msg, err := fb.Messaging(ctx)
	if err != nil {
		log.From(ctx).Fatal("preparing firebase messaging", zap.Error(err))
	}

	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) api.Response {
		var (
			resp = &Response{Error: &api.Error{}}
		)

		number := chi.URLParam(r, "number")
		apiKey := r.Header.Get("Authorization")

		var token []byte
		err := DB.View(func(txn *badger.Txn) error {

			log.From(ctx).Info("retrieving device token", zap.String("apiKey", apiKey))
			i, err := txn.Get([]byte(apiKey))
			if err != nil {
				return errors.Wrap(err, "retrieving device token")
			}

			token, err = i.Value()
			if err != nil {
				return errors.Wrap(err, "decoding device token")
			}

			return nil
		})

		if err != nil {
			resp.Fail(err)
			return resp
		}

		message := &messaging.Message{
			Data: map[string]string{
				"num": number,
			},
			Notification: &messaging.Notification{
				Title: "//S/M FireCall",
				Body:  fmt.Sprintf("Click to dial: %s", number),
			},
			Token: string(token),
		}

		_, err = msg.Send(ctx, message)
		if err != nil {
			resp.Fail(errors.Wrap(err, "sending message"))
			return resp
		}

		return resp
	}
}

// PrepareDB and return error
func PrepareDB(ctx context.Context, svc Spec) (err error) {
	opts := badger.DefaultOptions
	opts.Dir = "data/"
	opts.ValueDir = "data/"
	DB, err = badger.Open(opts)
	return err
}
