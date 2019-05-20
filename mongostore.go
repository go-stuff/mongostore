package mongostore

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// MongoStore stores sessions in MongoDB.
type MongoStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options
	ctx     context.Context
	col     *mongo.Collection
}

// NewMongoStore returns a new MongoStore.
//
// Keys are defined in pairs to allow key rotation, but the common case is
// to set a single authentication key and optionally an encryption key.
//
// The first key in a pair is used for authentication and the second for
// encryption. The encryption key can be set to nil or omitted in the last
// pair, but the authentication key is required in all pairs.
//
// It is recommended to use an authentication key with 32 or 64 bytes.
// The encryption key, if set, must be either 16, 24, or 32 bytes to select
// AES-128, AES-192, or AES-256 modes.
func NewMongoStore(mc *mongo.Collection, maxAge int, keyPairs ...[]byte) *MongoStore {
	// set authentication key environment variables if it is not set
	_, ok := os.LookupEnv("GORILLA_SESSION_AUTH_KEY")
	if !ok {
		os.Setenv("GORILLA_SESSION_AUTH_KEY", string(securecookie.GenerateRandomKey(32)))
	}

	// set encryption key environment variable if it is not set
	_, ok = os.LookupEnv("GORILLA_SESSION_ENC_KEY")
	if !ok {
		os.Setenv("GORILLA_SESSION_ENC_KEY", string(securecookie.GenerateRandomKey(16)))
	}

	ms := &MongoStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: maxAge, // 86400 * 30,
		},
	}
	ms.MaxAge(ms.Options.MaxAge)
	ms.ctx = context.Background()
	ms.col = mc

	// add TTL index if it does not exist
	err := ms.insertTTLIndexInMongo()
	if err != nil {
		log.Fatal(err)
	}

	return ms
}

// Get returns a session for the given name after adding it to the registry.
//
// It returns a new session if the sessions doesn't exist. Access IsNew on
// the session to check if it is an existing session or a new one.
//
// It returns a new session and an error if the session exists but could
// not be decoded.
func (ms *MongoStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(ms, name)
}

// New returns a session for the given name without adding it to the registry.
//
// The difference between New() and Get() is that calling New() twice will
// decode the session data twice, while Get() registers and reuses the same
// decoded session after the first call.
func (ms *MongoStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(ms, name)
	opts := *ms.Options
	session.Options = &opts
	session.IsNew = true
	var err error
	c, errCookie := r.Cookie(name)

	// if the session cookie already exits
	if errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, ms.Codecs...)

		// using the session.ID from the cookie decode the session.Values from mongo
		if err == nil {
			err = ms.findInMongo(session)
			// found existing session in mongo, set IsNew to false
			if err == nil {
				session.IsNew = false
			}
		}
	}

	return session, err
}

// Save adds a single session to the response.
func (ms *MongoStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	_, errCookie := r.Cookie(session.Name())
	if errCookie != nil {
		// insert into mongo
		err := ms.insertInMongo(session)
		if err != nil {
			return err
		}
	} else {
		if session.Options.MaxAge == -1 {
			// if session is expired delete from mongo
			err := ms.deleteFromMongo(session)
			if err != nil {
				return err
			}
		} else {
			// else update mongo
			err := ms.updateInMongo(session)
			if err != nil {
				return err
			}
		}
	}
	// update cookie
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, ms.Codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))

	return nil
}

// MaxAge sets the maximum age for the store and the underlying cookie
// implementation. Individual sessions can be deleted by setting Options.MaxAge
// = -1 for that session.
func (ms *MongoStore) MaxAge(age int) {
	ms.Options.MaxAge = age

	// Set the maxAge for each securecookie instance.
	for _, codec := range ms.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

func (ms *MongoStore) insertTTLIndexInMongo() error {
	// search for an index-ttl index in this collection
	cursor, err := ms.col.Indexes().List(ms.ctx)
	if err != nil {
		log.Fatal(err)
	}
	var foundTTLIndex bool
	for cursor.Next(ms.ctx) {
		var result bson.D
		err := cursor.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		if result.Map()["name"] == "modifiedAt_1" {
			foundTTLIndex = true
		}
	}

	//https://docs.mongodb.com/manual/core/index-ttl/
	//
	// TTL indexes are special single-field indexes that MongoDB can use to automatically
	// remove documents from a collection after a certain amount of time or at a specific
	// clock time. Data expiration is useful for certain types of information like machine
	// generated event data, logs, and session information that only need to persist in a
	// database for a finite amount of time.
	//
	// To create a TTL index, use the db.collection.createIndex() method with the
	// expireAfterSeconds option on a field whose value is either a date or an array that
	// contains date values.
	//
	// TTL indexes expire documents after the specified number of seconds has passed since
	// the indexed field value; i.e. the expiration threshold is the indexed field value
	// plus the specified number of seconds.
	//
	// The _id field does not support TTL indexes.
	if !foundTTLIndex {
		_, err = ms.col.Indexes().CreateOne(
			ms.ctx,
			mongo.IndexModel{
				Keys: bsonx.Doc{
					{Key: "modifiedAt", Value: bsonx.Int32(1)},
				},
				Options: options.Index().
					SetBackground(true).
					SetSparse(true).
					SetExpireAfterSeconds(int32(ms.Options.MaxAge)),
			},
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func (ms *MongoStore) findInMongo(session *sessions.Session) error {
	// get the session.ID as a mongo ObjectID type
	oid, err := primitive.ObjectIDFromHex(session.ID)
	if err != nil {
		return err
	}

	// find the session in mongo using the ObjectID and put the result in singleResult
	var singleResult interface{}
	err = ms.col.FindOne(ms.ctx, bson.M{"_id": oid}).Decode(&singleResult)
	if err != nil {
		return err
	}

	// use session.Values values from mongo
	for k, v := range singleResult.(primitive.D).Map() {
		session.Values[k] = v
	}

	return nil
}

func (ms *MongoStore) insertInMongo(session *sessions.Session) error {
	// load session.Values into a bson.D object
	var insert bson.D
	for k, v := range session.Values {
		insert = append(insert, bson.E{Key: k.(string), Value: v})
	}
	insert = append(insert, bson.E{Key: "createdAt", Value: primitive.DateTime(time.Now().Truncate(time.Millisecond).UnixNano() / int64(time.Millisecond))})
	insert = append(insert, bson.E{Key: "modifiedAt", Value: primitive.DateTime(time.Now().Truncate(time.Millisecond).UnixNano() / int64(time.Millisecond))})
	insert = append(insert, bson.E{Key: "expiresAt", Value: primitive.DateTime(time.Now().Add(time.Duration(ms.Options.MaxAge)*time.Second).Truncate(time.Millisecond).UnixNano() / int64(time.Millisecond))})

	// insert session.Values into mongo and get the returned ObjectID
	insertResult, err := ms.col.InsertOne(ms.ctx, insert)
	if err != nil {
		return err
	}

	// use the mongo ObjectID as the session.ID
	session.ID = insertResult.InsertedID.(primitive.ObjectID).Hex()

	return nil
}

func (ms *MongoStore) updateInMongo(session *sessions.Session) error {
	// get the session.ID as a mongo ObjectID type
	oid, err := primitive.ObjectIDFromHex(session.ID)
	if err != nil {
		return err
	}

	// load session.Values into a bson.D object
	var update bson.D
	for k, v := range session.Values {
		switch k.(string) {
		case "modifiedAt":
			update = append(update, bson.E{Key: k.(string), Value: primitive.DateTime(time.Now().Truncate(time.Millisecond).UnixNano() / int64(time.Millisecond))})
		case "expiresAt":
			update = append(update, bson.E{Key: k.(string), Value: primitive.DateTime(time.Now().Add(time.Duration(ms.Options.MaxAge)*time.Second).Truncate(time.Millisecond).UnixNano() / int64(time.Millisecond))})
		default:
			update = append(update, bson.E{Key: k.(string), Value: v})
		}
	}

	// update session.Values in mongo
	_, err = ms.col.UpdateOne(ms.ctx, bson.M{"_id": oid}, bson.D{{Key: "$set", Value: update}})
	if err != nil {
		return err
	}

	return nil
}

func (ms *MongoStore) deleteFromMongo(session *sessions.Session) error {
	// get the session.ID as a mongo ObjectID type
	oid, err := primitive.ObjectIDFromHex(session.ID)
	if err != nil {
		return err
	}

	// delete the document with ObjectID from mongo
	_, err = ms.col.DeleteOne(ms.ctx, bson.D{
		{Key: "_id", Value: oid},
	})
	if err != nil {
		return err
	}

	return nil
}
