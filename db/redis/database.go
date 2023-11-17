package redis

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/ushis/gesundheit/db"
	"github.com/ushis/gesundheit/result"
)

type Database struct {
	rdb *redis.Client
}

type Config struct {
	Address  string
	DB       int
	Username string
	Password string
}

func init() {
	db.Register("redis", New)
}

// Database Layout
//
// nodes                            				set<nodeName>
// nodes:<nodename>:checks                  set<checkId>
// nodes:<nodename>:checks:<checkId>:events	list<eventId>
// events:<eventId>          								event
func New(configure func(interface{}) error) (db.Database, error) {
	conf := Config{}

	if err := configure(&conf); err != nil {
		return nil, err
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Address,
		DB:       conf.DB,
		Username: conf.Username,
		Password: conf.Password,
	})
	return Database{rdb}, nil
}

func (db Database) Close() error {
	return db.rdb.Close()
}

func (db Database) InsertEvent(e result.Event) (bool, error) {
	val, err := json.Marshal(e)

	if err != nil {
		return false, err
	}
	key := mkkey("events", e.Id)
	ttl := time.Until(e.ExpiresAt)

	if ok, err := db.rdb.SetNX(db.rdb.Context(), key, val, ttl).Result(); !ok || err != nil {
		return ok, err
	}
	key = mkkey("nodes", e.NodeName, "checks", e.CheckId, "events")

	if err := db.rdb.RPush(db.rdb.Context(), key, e.Id).Err(); err != nil {
		return true, err
	}
	if err := db.rdb.LTrim(db.rdb.Context(), key, -6, -1).Err(); err != nil {
		return true, err
	}
	key = mkkey("nodes", e.NodeName, "checks")

	if err := db.rdb.SAdd(db.rdb.Context(), key, e.CheckId).Err(); err != nil {
		return true, err
	}
	key = mkkey("nodes")

	return true, db.rdb.SAdd(db.rdb.Context(), key, e.NodeName).Err()
}

func (db Database) GetEventsByCheck(nodeName, checkId string) ([]result.Event, error) {
	key := mkkey("nodes", nodeName, "checks", checkId, "events")
	ids, err := db.rdb.LRange(db.rdb.Context(), key, 0, -1).Result()

	if err != nil {
		return nil, err
	}
	keys := make([]string, len(ids))

	for i, id := range ids {
		keys[i] = mkkey("events", id)
	}
	vals, err := db.rdb.MGet(db.rdb.Context(), keys...).Result()

	if err != nil {
		return nil, err
	}
	i := 0
	events := make([]result.Event, len(vals))

	for _, val := range vals {
		str, ok := val.(string)

		if !ok {
			continue
		}
		event := result.Event{}

		if err := json.Unmarshal([]byte(str), &event); err != nil {
			return nil, err
		}
		events[i] = event
		i += 1
	}
	return events[:i], nil
}

func (db Database) GetEventsByNode(nodeName string) ([]result.Event, error) {
	key := mkkey("nodes", nodeName, "checks")
	checkIds, err := db.rdb.SMembers(db.rdb.Context(), key).Result()

	if err != nil {
		return nil, err
	}
	events := []result.Event{}

	for _, checkId := range checkIds {
		checkEvents, err := db.GetEventsByCheck(nodeName, checkId)

		if err != nil {
			return nil, err
		}
		events = append(events, checkEvents...)
	}
	return events, nil
}

func (db Database) GetEvents() ([]result.Event, error) {
	key := mkkey("nodes")
	nodeNames, err := db.rdb.SMembers(db.rdb.Context(), key).Result()

	if err != nil {
		return nil, err
	}
	events := []result.Event{}

	for _, nodeName := range nodeNames {
		nodeEvents, err := db.GetEventsByNode(nodeName)

		if err != nil {
			return nil, err
		}
		events = append(events, nodeEvents...)
	}
	return events, nil
}

const pathSep = ":"

func mkkey(path ...string) string {
	return strings.Join(path, pathSep)
}
