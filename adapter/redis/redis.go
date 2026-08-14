// Package redis is furumai's KVS (Redis) Stimulus/Observation adapter.
//
// Unlike the MySQL and Kafka adapters, this one has no external
// dependency: RESP (Redis's wire protocol) is simple enough that a
// minimal client over net/bufio is a reasonable "standard library can do
// this" call under the project's dependency-minimization policy, rather
// than an unavoidable exception.
package redis

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

// Stimulus runs commands against a Redis server. It is used from both
// given and when steps, and also serves as the Observation side via Get.
type Stimulus struct {
	conn net.Conn
	r    *bufio.Reader
}

// NewStimulus dials addr (host:port).
func NewStimulus(addr string) (*Stimulus, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Stimulus{conn: conn, r: bufio.NewReader(conn)}, nil
}

// Close closes the underlying connection.
func (s *Stimulus) Close() error { return s.conn.Close() }

func (s *Stimulus) do(args ...string) (any, error) {
	if _, err := s.conn.Write(encodeCommand(args)); err != nil {
		return nil, err
	}
	return readReply(s.r)
}

// encodeCommand renders args as a RESP array of bulk strings.
func encodeCommand(args []string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(args))
	for _, a := range args {
		fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(a), a)
	}
	return []byte(b.String())
}

// readReply parses a single RESP reply: a simple/bulk string decodes to
// string, an integer to int64, an array to []any, and a nil bulk
// string/array to nil. An error reply ("-...") is returned as a Go error.
func readReply(r *bufio.Reader) (any, error) {
	line, err := readLine(r)
	if err != nil {
		return nil, err
	}
	if line == "" {
		return nil, fmt.Errorf("redis: empty reply line")
	}

	switch line[0] {
	case '+':
		return line[1:], nil
	case '-':
		return nil, fmt.Errorf("redis: %s", line[1:])
	case ':':
		return strconv.ParseInt(line[1:], 10, 64)
	case '$':
		n, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, err
		}
		if n == -1 {
			return nil, nil
		}
		buf := make([]byte, n+2) // +2 for the trailing \r\n
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		return string(buf[:n]), nil
	case '*':
		n, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, err
		}
		if n == -1 {
			return nil, nil
		}
		arr := make([]any, n)
		for i := 0; i < n; i++ {
			v, err := readReply(r)
			if err != nil {
				return nil, err
			}
			arr[i] = v
		}
		return arr, nil
	default:
		return nil, fmt.Errorf("redis: unknown reply type %q", line[0])
	}
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

type setOptions struct {
	ttl time.Duration
}

// SetOption customizes Set.
type SetOption func(*setOptions)

// WithTTL expires the key after d.
func WithTTL(d time.Duration) SetOption {
	return func(o *setOptions) { o.ttl = d }
}

// Set writes key/value, optionally with a TTL.
func (s *Stimulus) Set(key, value string, opts ...SetOption) error {
	var o setOptions
	for _, opt := range opts {
		opt(&o)
	}

	args := []string{"SET", key, value}
	if o.ttl > 0 {
		args = append(args, "PX", strconv.FormatInt(o.ttl.Milliseconds(), 10))
	}
	_, err := s.do(args...)
	return err
}

// Del deletes keys.
func (s *Stimulus) Del(keys ...string) error {
	_, err := s.do(append([]string{"DEL"}, keys...)...)
	return err
}

// Expire sets a TTL on an existing key.
func (s *Stimulus) Expire(key string, ttl time.Duration) error {
	_, err := s.do("PEXPIRE", key, strconv.FormatInt(ttl.Milliseconds(), 10))
	return err
}

// HSet sets hash fields on key.
func (s *Stimulus) HSet(key string, fields map[string]string) error {
	args := []string{"HSET", key}
	for field, value := range fields {
		args = append(args, field, value)
	}
	_, err := s.do(args...)
	return err
}

// LPush prepends values to the list at key.
func (s *Stimulus) LPush(key string, values ...string) error {
	_, err := s.do(append([]string{"LPUSH", key}, values...)...)
	return err
}

// SAdd adds members to the set at key.
func (s *Stimulus) SAdd(key string, members ...string) error {
	_, err := s.do(append([]string{"SADD", key}, members...)...)
	return err
}

// ZAdd adds member/score pairs to the sorted set at key.
func (s *Stimulus) ZAdd(key string, members map[string]float64) error {
	args := []string{"ZADD", key}
	for member, score := range members {
		args = append(args, strconv.FormatFloat(score, 'f', -1, 64), member)
	}
	_, err := s.do(args...)
	return err
}

// FlushDB deletes every key in the current database. This is destructive
// and applies to the whole database rather than a namespace, so it
// requires an explicit confirm=true to reduce the chance of an accidental
// call.
func (s *Stimulus) FlushDB(confirm bool) error {
	if !confirm {
		return fmt.Errorf("redis: FlushDB requires confirm=true")
	}
	_, err := s.do("FLUSHDB")
	return err
}

// Value is the full-state Observation for a single key. Its dynamic type
// depends on the key's Redis data type: string for string keys,
// map[string]string for hashes, []string for lists and sets, and
// map[string]float64 (member -> score) for sorted sets. It's typed any so
// a furumai.Matcher (Any, Regex, ...) can be substituted for it.
type Value any

// Get returns every key matching pattern (Redis KEYS syntax, e.g. "user:*")
// as the full-state Observation, keyed by key name.
func (s *Stimulus) Get(pattern string) (map[string]Value, error) {
	keysReply, err := s.do("KEYS", pattern)
	if err != nil {
		return nil, err
	}
	keys := toStringSlice(keysReply)

	result := make(map[string]Value, len(keys))
	for _, key := range keys {
		typReply, err := s.do("TYPE", key)
		if err != nil {
			return nil, err
		}
		typ, _ := typReply.(string)

		switch typ {
		case "string":
			v, err := s.do("GET", key)
			if err != nil {
				return nil, err
			}
			result[key] = v
		case "hash":
			v, err := s.do("HGETALL", key)
			if err != nil {
				return nil, err
			}
			result[key] = toStringMap(v)
		case "list":
			v, err := s.do("LRANGE", key, "0", "-1")
			if err != nil {
				return nil, err
			}
			result[key] = toStringSlice(v)
		case "set":
			v, err := s.do("SMEMBERS", key)
			if err != nil {
				return nil, err
			}
			result[key] = toStringSlice(v)
		case "zset":
			v, err := s.do("ZRANGE", key, "0", "-1", "WITHSCORES")
			if err != nil {
				return nil, err
			}
			result[key] = toScoreMap(v)
		default:
			result[key] = nil
		}
	}
	return result, nil
}

func toStringSlice(v any) []string {
	arr, _ := v.([]any)
	out := make([]string, 0, len(arr))
	for _, e := range arr {
		if s, ok := e.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// toStringMap converts a flat [field, value, field, value, ...] RESP array
// (as returned by HGETALL) into a map.
func toStringMap(v any) map[string]string {
	arr, _ := v.([]any)
	m := make(map[string]string, len(arr)/2)
	for i := 0; i+1 < len(arr); i += 2 {
		field, _ := arr[i].(string)
		value, _ := arr[i+1].(string)
		m[field] = value
	}
	return m
}

// toScoreMap converts a flat [member, score, member, score, ...] RESP
// array (as returned by ZRANGE ... WITHSCORES) into a member -> score map.
func toScoreMap(v any) map[string]float64 {
	arr, _ := v.([]any)
	m := make(map[string]float64, len(arr)/2)
	for i := 0; i+1 < len(arr); i += 2 {
		member, _ := arr[i].(string)
		scoreStr, _ := arr[i+1].(string)
		score, _ := strconv.ParseFloat(scoreStr, 64)
		m[member] = score
	}
	return m
}
