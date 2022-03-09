package fixtures

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/feeds/content"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/network/rpc"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"go.etcd.io/bbolt"
)

func SomeLogger() logging.Logger {
	return logging.NewDevNullLogger()
}

func TestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return ctx
}

func SomeProcedureName() rpc.ProcedureName {
	return rpc.MustNewProcedureName([]string{randomBase64(10)})
}

func SomeProcedureType() rpc.ProcedureType {
	return rpc.ProcedureTypeAsync
}

func SomeBool() bool {
	return true
}

func SomeRefMessage() refs.Message {
	// todo improve this by using some kind of a better constructor
	return refs.MustNewMessage(fmt.Sprintf("%%%s.sha256", randomBase64(32)))
}

func SomeRefAuthor() refs.Identity {
	// todo improve this by using some kind of a better constructor
	return refs.MustNewIdentity(fmt.Sprintf("@%s.ed25519", randomBase64(32)))
}

func SomeRefFeed() refs.Feed {
	// todo improve this by using some kind of a better constructor
	return refs.MustNewFeed(fmt.Sprintf("@%s.ed25519", randomBase64(32)))
}

func SomeTime() time.Time {
	// todo improve this by using some kind of a better constructor
	return time.Unix(rand.Int63(), 0)
}

func SomePublicIdentity() identity.Public {
	v, err := identity.NewPrivate()
	if err != nil {
		panic(err)
	}
	return v.Public()
}

func SomeContent() message.MessageContent {
	return content.MustNewUnknown(SomeBytes())
}

func SomeBytes() []byte {
	r := make([]byte, 10+rand.Intn(100))
	_, err := rand.Read(r)
	if err != nil {
		panic(err)
	}
	return r
}

func SomeJSON() []byte {
	return []byte(`{"key":"value"}`)
}

func SomeRawMessage() message.RawMessage {
	return message.NewRawMessage(SomeBytes())
}

func randomBase64(bytes int) string {
	r := make([]byte, bytes)
	_, err := rand.Read(r)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(r)
}

func File(t *testing.T) string {
	file, err := ioutil.TempFile("", "eggplant_test")
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		err := os.Remove(file.Name())
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(cleanup)

	return file.Name()
}

func Bolt(t *testing.T) *bbolt.DB {
	file := File(t)

	db, err := bbolt.Open(file, 0600, &bbolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(cleanup)

	return db
}
