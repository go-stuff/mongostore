package mongostore_test

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	ms "github.com/go-stuff/mongostore"
	"github.com/gorilla/securecookie"
)

var err error
var mongoclient *mongo.Client
var mongostore *ms.MongoStore

func TestMain(m *testing.M) {
	// connect to database
	testsSetup()

	// run all tests
	ret := m.Run()

	// disconnect from database
	testsTeardown()

	// call flag.Parse() here if TestMain uses flags
	os.Exit(ret)
}

func testsSetup() {

	// set needed environment variables if none are set
	// _, ok := os.LookupEnv("GORILLA_SESSION_AUTH_KEY")
	// if !ok {
	os.Setenv("GORILLA_SESSION_AUTH_KEY", string(securecookie.GenerateRandomKey(32)))
	// }
	// _, ok = os.LookupEnv("GORILLA_SESSION_ENC_KEY")
	// if !ok {
	os.Setenv("GORILLA_SESSION_ENC_KEY", string(securecookie.GenerateRandomKey(16)))
	// }

	// A Context carries a deadline, cancelation signal, and request-scoped values
	// across API boundaries. Its methods are safe for simultaneous use by multiple
	// goroutines.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect does not do server discovery, use Ping method.
	mongoclient, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	// Ping for server discovery.
	err = mongoclient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
}

func testsTeardown() {
	err = mongoclient.Disconnect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func TestNewMongoStore(t *testing.T) {
	// without environment variables
	os.Clearenv()

	// get a new mongostore
	mongostore = ms.NewMongoStore(
		mongoclient.Database("test").Collection("sessions_test"),
		240,
		[]byte(os.Getenv("GORILLA_SESSION_AUTH_KEY")),
		[]byte(os.Getenv("GORILLA_SESSION_ENC_KEY")),
	)

	if mongostore == nil {
		t.Fatal("expected to fail with no environment variables")
	}

	// without TTL index
	_, err = mongoclient.Database("test").Collection("sessions_test").Indexes().DropAll(context.TODO())
	if err != nil {
		t.Fatalf("failed to drop mongo indexes: %v\n", err)
	}

	// with environment variables
	os.Setenv("GORILLA_SESSION_AUTH_KEY", string(securecookie.GenerateRandomKey(32)))
	os.Setenv("GORILLA_SESSION_ENC_KEY", string(securecookie.GenerateRandomKey(16)))

	// get a new mongostore
	mongostore = ms.NewMongoStore(
		mongoclient.Database("test").Collection("sessions_test"),
		240,
		[]byte(os.Getenv("GORILLA_SESSION_AUTH_KEY")),
		[]byte(os.Getenv("GORILLA_SESSION_ENC_KEY")),
	)

	// if mongostore is nil, throw an error
	if mongostore == nil {
		t.Fatal("failed to create new mongostore")
	}
}

func TestGet(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/", nil)

	_, err := mongostore.Get(req, "test-session")
	if err != nil {
		t.Fatalf("failed to get session: %v\n", err)
	}
}

func TestNew(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/", nil)

	// new session
	_, err = mongostore.New(req, "test-session")
	if err != nil {
		t.Fatalf("failed to create new session: %v\n", err)
	}

	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)

	// existing session
	_, err = mongostore.New(req, "test-session")
	if err != nil {
		t.Fatalf("failed to create new session: %v\n", err)
	}
}

func TestSave(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/", nil)
	res := httptest.NewRecorder()

	// insert mongo
	session, err := mongostore.Get(req, "test-session")
	if err != nil {
		t.Fatalf("failed to get session: %v\n", err)
	}
	session.Values["test"] = "testdata"
	err = mongostore.Save(req, res, session)
	if err != nil {
		t.Fatalf("failed to insert session: %v\n", err)
	}

	// insert cookie
	hdr := res.Header()
	cookies, ok := hdr["Set-Cookie"]
	if !ok || len(cookies) != 1 {
		t.Fatal("no cookies. header:", hdr)
	}

	// update mongo
	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	req.Header.Add("Cookie", cookies[0])
	res = httptest.NewRecorder()

	session, err = mongostore.Get(req, "test-session")
	if err != nil {
		t.Fatalf("failed to get session: %v\n", err)
	}
	session.Options.MaxAge = 7357
	err = mongostore.Save(req, res, session)
	if err != nil {
		t.Fatal("failed to update session", err)
	}

	// insert cookie
	hdr = res.Header()
	cookies, ok = hdr["Set-Cookie"]
	if !ok || len(cookies) != 1 {
		t.Fatal("no cookies. header:", hdr)
	}

	// expire mongo
	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	req.Header.Add("Cookie", cookies[0])
	res = httptest.NewRecorder()

	session, err = mongostore.Get(req, "test-session")
	if err != nil {
		t.Fatalf("failed to get session: %v\n", err)
	}
	session.Options.MaxAge = -1
	err = mongostore.Save(req, res, session)
	if err != nil {
		t.Fatal("failed to expire session", err)
	}

}

func TestMaxAge(t *testing.T) {
	mongostore.MaxAge(7357)
	if mongostore.Options.MaxAge != 7357 {
		t.Fatalf("failed to set MaxAge: %v\n", err)
	}
}
