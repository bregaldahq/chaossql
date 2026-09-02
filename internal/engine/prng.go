package engine

import (
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// PRNG provides deterministic pseudo-random generation.
type PRNG struct {
	masterSeed uint64
}

// NewPRNG initializes a new PRNG with a master seed.
func NewPRNG(seed uint64) *PRNG {
	if seed == 0 {
		seed = uint64(time.Now().UnixNano())
	}
	return &PRNG{masterSeed: seed}
}

// MasterSeed returns the seed used.
func (p *PRNG) MasterSeed() uint64 {
	return p.masterSeed
}

// WorkerSeed derives an isolated sub-seed for a specific worker using high-performance SplitMix64.
func (p *PRNG) WorkerSeed(workerID int) uint64 {
	x := p.masterSeed ^ (uint64(workerID) * 0x9e3779b97f4a7c15)
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	return x ^ (x >> 31)
}

// EvaluateParam parses declarative parameter generators or dynamic expressions.
func (p *PRNG) EvaluateParam(expr string, rng *rand.Rand) string {
	val, err := EvaluateGenerator(expr, rng)
	if err == nil {
		return val
	}
	return expr
}

// Jitter returns a uniform delay between jitterRange[0] and jitterRange[1] milliseconds.
func (p *PRNG) Jitter(jitterRange [2]int, rng *rand.Rand) time.Duration {
	minMs := jitterRange[0]
	maxMs := jitterRange[1]
	if maxMs <= 0 || maxMs < minMs {
		return 0
	}
	delayMs := minMs + rng.IntN(maxMs-minMs+1)
	return time.Duration(delayMs) * time.Millisecond
}

type counterState struct {
	val atomic.Int64
}

var monotonicCounters sync.Map

// ResetMonotonicCounters clears all monotonic counter states (useful for test resets).
func ResetMonotonicCounters() {
	monotonicCounters.Range(func(key, value any) bool {
		monotonicCounters.Delete(key)
		return true
	})
}

var emailPrefixes = []string{
	"user", "trader", "admin", "dev", "alex", "sam", "jordan", "taylor",
	"casey", "morgan", "riley", "quinn", "avery", "skyler", "dakota",
	"charlie", "finley", "rowan", "river", "sage",
}

var emailDomains = []string{
	"example.com", "test.org", "defi.org", "chaos.io", "database.net",
	"acme.corp", "mail.dev", "domain.com", "cloud.io", "fintech.ai",
}

var firstNames = []string{
	"Alice", "Bob", "Charlie", "David", "Emma", "Frank", "Grace", "Hannah",
	"Isaac", "Julia", "Kevin", "Laura", "Michael", "Nina", "Oliver", "Paula",
	"Quinn", "Rachel", "Samuel", "Tina", "Umar", "Victor", "Wendy", "Xavier",
	"Yasmine", "Zachary", "Alexander", "Sophia", "Lucas", "Mia",
}

var lastNames = []string{
	"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller",
	"Davis", "Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez",
	"Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin",
	"Lee", "Perez", "Thompson", "White", "Harris", "Sanchez", "Clark", "Ramirez",
}

// EvaluateGenerator evaluates dynamic random generators deterministically using r.
func EvaluateGenerator(expr string, r *rand.Rand) (string, error) {
	if r == nil {
		return "", fmt.Errorf("nil rand.Rand provided to EvaluateGenerator")
	}
	trimmed := strings.TrimSpace(expr)

	if strings.HasPrefix(trimmed, "$random_int(") && strings.HasSuffix(trimmed, ")") {
		args := trimmed[len("$random_int(") : len(trimmed)-1]
		return evalRandomInt(args, r)
	}

	if strings.HasPrefix(trimmed, "$random_choice(") && strings.HasSuffix(trimmed, ")") {
		args := trimmed[len("$random_choice(") : len(trimmed)-1]
		return evalRandomChoice(args, r)
	}

	if strings.HasPrefix(trimmed, "$random_string(") && strings.HasSuffix(trimmed, ")") {
		args := trimmed[len("$random_string(") : len(trimmed)-1]
		return evalRandomString(args, r)
	}

	if strings.HasPrefix(trimmed, "$uuid(") && strings.HasSuffix(trimmed, ")") {
		args := trimmed[len("$uuid(") : len(trimmed)-1]
		return evalUUID(args, r)
	}

	if strings.HasPrefix(trimmed, "$faker_email(") && strings.HasSuffix(trimmed, ")") {
		args := trimmed[len("$faker_email(") : len(trimmed)-1]
		return evalFakerEmail(args, r)
	}

	if strings.HasPrefix(trimmed, "$faker_name(") && strings.HasSuffix(trimmed, ")") {
		args := trimmed[len("$faker_name(") : len(trimmed)-1]
		return evalFakerName(args, r)
	}

	if strings.HasPrefix(trimmed, "$faker_phone(") && strings.HasSuffix(trimmed, ")") {
		args := trimmed[len("$faker_phone(") : len(trimmed)-1]
		return evalFakerPhone(args, r)
	}

	if strings.HasPrefix(trimmed, "$monotonic_counter(") && strings.HasSuffix(trimmed, ")") {
		args := trimmed[len("$monotonic_counter(") : len(trimmed)-1]
		return evalMonotonicCounter(args)
	}

	// Legacy format: int(min, max)
	if strings.HasPrefix(trimmed, "int(") && strings.HasSuffix(trimmed, ")") {
		args := trimmed[len("int(") : len(trimmed)-1]
		return evalRandomInt(args, r)
	}

	if strings.HasPrefix(trimmed, "$") {
		return "", fmt.Errorf("unsupported generator expression: %s", expr)
	}

	return expr, nil
}

func evalRandomInt(args string, r *rand.Rand) (string, error) {
	parts := strings.Split(args, ",")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid arguments for $random_int, expected (min, max): %s", args)
	}
	minVal, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	maxVal, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return "", fmt.Errorf("invalid integer arguments for $random_int: %s", args)
	}
	if minVal > maxVal {
		return "", fmt.Errorf("invalid range for $random_int: min (%d) > max (%d)", minVal, maxVal)
	}
	if minVal == maxVal {
		return strconv.Itoa(minVal), nil
	}
	val := minVal + int(r.Int64N(int64(maxVal-minVal+1)))
	return strconv.Itoa(val), nil
}

func evalRandomChoice(args string, r *rand.Rand) (string, error) {
	choices, err := parseChoiceArgs(args)
	if err != nil {
		return "", err
	}
	if len(choices) == 0 {
		return "", fmt.Errorf("$random_choice requires at least one choice")
	}
	idx := r.IntN(len(choices))
	return choices[idx], nil
}

func parseChoiceArgs(s string) ([]string, error) {
	var choices []string
	var cur strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if escaped {
			cur.WriteByte(ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			continue
		}
		if ch == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			continue
		}
		if ch == ',' && !inSingleQuote && !inDoubleQuote {
			choice := strings.TrimSpace(cur.String())
			if choice != "" {
				choices = append(choices, choice)
			}
			cur.Reset()
			continue
		}
		cur.WriteByte(ch)
	}

	if inSingleQuote || inDoubleQuote {
		return nil, fmt.Errorf("unclosed quote in arguments: %s", s)
	}

	choice := strings.TrimSpace(cur.String())
	if choice != "" {
		choices = append(choices, choice)
	}

	if len(choices) == 0 {
		return nil, fmt.Errorf("no choices provided in $random_choice")
	}

	return choices, nil
}

func evalRandomString(args string, r *rand.Rand) (string, error) {
	n, err := strconv.Atoi(strings.TrimSpace(args))
	if err != nil {
		return "", fmt.Errorf("invalid length for $random_string: %s", args)
	}
	if n < 0 {
		return "", fmt.Errorf("length cannot be negative for $random_string: %d", n)
	}
	if n == 0 {
		return "", nil
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		buf[i] = charset[r.IntN(len(charset))]
	}
	return string(buf), nil
}

func evalUUID(args string, r *rand.Rand) (string, error) {
	if strings.TrimSpace(args) != "" {
		return "", fmt.Errorf("$uuid() takes no arguments, got: %s", args)
	}
	var b [16]byte
	for i := 0; i < 16; i++ {
		b[i] = byte(r.IntN(256))
	}
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant RFC 4122

	var buf [36]byte
	hex.Encode(buf[0:8], b[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], b[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], b[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], b[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], b[10:16])
	return string(buf[:]), nil
}

func evalFakerEmail(args string, r *rand.Rand) (string, error) {
	if strings.TrimSpace(args) != "" {
		return "", fmt.Errorf("$faker_email() takes no arguments, got: %s", args)
	}
	prefix := emailPrefixes[r.IntN(len(emailPrefixes))]
	num := r.IntN(1000)
	domain := emailDomains[r.IntN(len(emailDomains))]
	return fmt.Sprintf("%s_%d@%s", prefix, num, domain), nil
}

func evalFakerName(args string, r *rand.Rand) (string, error) {
	if strings.TrimSpace(args) != "" {
		return "", fmt.Errorf("$faker_name() takes no arguments, got: %s", args)
	}
	first := firstNames[r.IntN(len(firstNames))]
	last := lastNames[r.IntN(len(lastNames))]
	return first + " " + last, nil
}

func evalFakerPhone(args string, r *rand.Rand) (string, error) {
	if strings.TrimSpace(args) != "" {
		return "", fmt.Errorf("$faker_phone() takes no arguments, got: %s", args)
	}
	num := r.IntN(10000)
	return fmt.Sprintf("+1-555-%04d", num), nil
}

func evalMonotonicCounter(args string) (string, error) {
	trimmed := strings.TrimSpace(args)
	if trimmed == "" {
		return "", fmt.Errorf("$monotonic_counter requires at least a start argument, got empty args")
	}

	parts := strings.Split(trimmed, ",")
	var start, step int64
	var err error

	if len(parts) == 1 {
		start, err = strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid start argument for $monotonic_counter: %s", args)
		}
		step = 1
	} else if len(parts) == 2 {
		start, err = strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid start argument for $monotonic_counter: %s", args)
		}
		step, err = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid step argument for $monotonic_counter: %s", args)
		}
	} else {
		return "", fmt.Errorf("invalid arguments for $monotonic_counter, expected (start, step): %s", args)
	}

	key := fmt.Sprintf("%d:%d", start, step)
	c := &counterState{}
	c.val.Store(start)

	actual, _ := monotonicCounters.LoadOrStore(key, c)
	state := actual.(*counterState)
	currentVal := state.val.Add(step) - step

	return strconv.FormatInt(currentVal, 10), nil
}
