package redis

import (
	"encoding/json"
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
// nodes                          set<nodeName>
// nodes:<nodename>:events        hash<checkId, eventId>
// nodes:<nodename>:events:latest eventId
// events:<eventId>               event
//
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

// store event
//
// keys: nodeName, checkId, eventId
// vals: event, ttl
//
const scriptStoreEvent = `
	if redis.call('exists', 'events:' .. KEYS[3]) > 0 then return 0 end
	redis.call('set', 'events:' .. KEYS[3], ARGV[1], 'ex', ARGV[2])
	redis.call('hset', 'nodes:' .. KEYS[1] .. ':events', KEYS[2], KEYS[3])
	redis.call('set', 'nodes:' .. KEYS[1] .. ':events:latest', KEYS[3])
	redis.call('sadd', 'nodes', KEYS[1])
	return 1
`

func (db Database) InsertEvent(e result.Event) (bool, error) {
	val, err := json.Marshal(e)

	if err != nil {
		return false, err
	}
	ttl := int64(time.Until(e.ExpiresAt) / time.Second)
	keys := []string{e.NodeName, e.CheckId, e.Id}
	vals := []interface{}{val, ttl}
	res, err := db.rdb.Eval(db.rdb.Context(), scriptStoreEvent, keys, vals...).Int()
	return res == 1, err
}

// retreive events
//
// keys: [nodeName]
//
const scriptGetEvents = `
	local function getNodes()
		return ipairs(redis.call('smembers', 'nodes'))
	end

	local function getEventIds(node)
		return ipairs(redis.call('hvals', 'nodes:' .. node .. ':events'))
	end

	local function getEvent(eventId)
		return redis.call('get', 'events:' .. eventId)
	end

	local function getEventsByNode(events, node)
		for _, eventId in getEventIds(node) do
			local event = getEvent(eventId)
			if event then events[#events+1] = event end
		end
		return events
	end

	local function getEvents(events)
		for _, node in getNodes() do
			events = getEventsByNode(events, node)
		end
		return events
	end

	if #KEYS == 0 then
		return getEvents({})
	else
		return getEventsByNode({}, KEYS[1])
	end
`

func (db Database) getEvents(path ...string) ([]result.Event, error) {
	vals, err := db.rdb.Eval(db.rdb.Context(), scriptGetEvents, path, nil).StringSlice()

	if err != nil {
		return nil, err
	}
	events := make([]result.Event, len(vals))

	for i, val := range vals {
		event := result.Event{}

		if err := json.Unmarshal([]byte(val), &event); err != nil {
			return nil, err
		}
		events[i] = event
	}
	return events, nil
}

func (db Database) GetEvents() ([]result.Event, error) {
	return db.getEvents()
}

func (db Database) GetEventsByNode(name string) ([]result.Event, error) {
	return db.getEvents(name)
}

// retreive event
//
// keys: [<key>...]
//
const scriptGetEvent = `
	local eventId = redis.call('get', table.concat(KEYS, ':'))
	if not eventId then return nil end
	return redis.call('get', 'events:' .. eventId)
`

func (db Database) getEvent(path ...string) (event result.Event, ok bool, err error) {
	val, err := db.rdb.Eval(db.rdb.Context(), scriptGetEvent, path, nil).Text()

	if err == redis.Nil {
		return event, false, nil
	}
	if err != nil {
		return event, false, err
	}
	if err := json.Unmarshal([]byte(val), &event); err != nil {
		return event, false, err
	}
	return event, true, nil
}

func (db Database) GetLatestEventByNode(name string) (event result.Event, ok bool, err error) {
	return db.getEvent("nodes", name, "events", "latest")
}
