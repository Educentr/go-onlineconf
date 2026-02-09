package onlineconf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Educentr/go-onlineconf/pkg/onlineconfInterface"
	"github.com/colinmarc/cdb"
	"golang.org/x/exp/mmap"
)

func (m *Module) Clone(name string) onlineconfInterface.Module { //nolint:ireturn
	return &Module{
		ro:          true,
		name:        name,
		filename:    m.filename,
		cache:       make(map[string][]interface{}, startCacheSize),
		cdb:         m.cdb,
		mmappedFile: m.mmappedFile,
	}
}

func (m *Module) GetMmappedFile() *mmap.ReaderAt {
	return m.mmappedFile
}

func (m *Module) Reopen(mmappedFile *mmap.ReaderAt) (*mmap.ReaderAt, error) {
	if m.ro {
		return nil, fmt.Errorf("unable to use Reopen in readonly instance")
	}

	m.Lock()

	cdb, err := cdb.New(mmappedFile, nil)
	if err != nil {
		return nil, err
	}

	callbacksToCall := []func() error{}
	for _, subscription := range m.changeSubscription {
		if subscription.GetPaths() == nil {
			callbacksToCall = append(callbacksToCall, subscription.InvokeCallback)
			continue
		}

		for _, path := range subscription.GetPaths() {
			if path == "" {
				callbacksToCall = append(callbacksToCall, subscription.InvokeCallback)
				break
			}

			newValue, newErr := cdb.Get([]byte(path))
			oldValue, oldErr := m.cdb.Get([]byte(path))

			// If one read errored and the other didn't, the state changed
			if (newErr != nil) != (oldErr != nil) {
				callbacksToCall = append(callbacksToCall, subscription.InvokeCallback)
				break
			}

			// If both errored, skip comparison for this path
			if newErr != nil {
				continue
			}

			// Both reads succeeded — compare actual byte content
			if !bytes.Equal(newValue, oldValue) {
				callbacksToCall = append(callbacksToCall, subscription.InvokeCallback)
				break
			}
		}
	}

	oldMmappedFile := m.mmappedFile
	m.mmappedFile = mmappedFile
	m.cdb = cdb

	m.cacheMutex.Lock()
	m.cache = map[string][]interface{}{}
	m.cacheMutex.Unlock()

	m.Unlock()

	for _, callback := range callbacksToCall {
		err := callback()
		if err != nil {
			// ToDo use app logger instance
			log.Printf("error in callback: %s", err)
		}
	}

	return oldMmappedFile, nil
}

func (m *Module) RegisterSubscription(subscription onlineconfInterface.SubscriptionCallback) {
	m.Lock()
	defer m.Unlock()

	m.changeSubscription = append(m.changeSubscription, subscription)
}

func (m *Module) get(path string) (byte, []byte, error) {
	m.RLock()
	defer m.RUnlock()

	data, err := m.cdb.Get([]byte(path))
	if err != nil || len(data) == 0 {
		if err != nil {
			return 0, data, fmt.Errorf("get %v:%v error: %w", m.filename, path, err)
		}

		return 0, data, nil
	}

	return data[0], data[1:], nil
}

// GetStringIfExists reads a string value of a named parameter from the module.
// It returns the boolean true if the parameter exists and is a string.
// In the other case it returns the boolean false and an empty string.
func (m *Module) GetStringIfExists(path string) (string, bool, error) {
	format, data, err := m.get(path)
	if err != nil {
		return "", false, err
	}

	switch format {
	case 0:
		return "", false, nil
	case 's':
		return string(data), true, nil
	default:
		return "", false, fmt.Errorf("%s:%s: format is not string", m.name, path)
	}
}

// GetIntIfExists reads an integer value of a named parameter from the module.
// It returns this value and the boolean true if the parameter exists and is an integer.
// In the other case it returns the boolean false and 0.
func (m *Module) GetIntIfExists(path string) (int64, bool, error) {
	str, ok, err := m.GetStringIfExists(path)
	if err != nil {
		return 0, false, err
	}

	if !ok {
		return 0, false, nil
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, false, fmt.Errorf("%s:%s: value is not an integer: %s", m.name, path, str)
	}

	return i, true, nil
}

// GetDuration reads an string value of a named parameter from the module and parse it to time.Duration.
// It returns this value if the parameter exists and is an time.Duration.
// In the other case it return error unless default value is provided in
// the second argument.
func (m *Module) GetDurationIfExists(path string) (time.Duration, bool, error) {
	str, ok, err := m.GetStringIfExists(path)
	if err != nil {
		return 0, false, err
	}

	if !ok {
		return 0, false, err
	}

	dur, err := time.ParseDuration(str)
	if err != nil {
		return 0, false, fmt.Errorf("%s:%s: value is not a duration: %s", m.name, path, str)
	}

	return dur, true, nil
}

// GetBoolIfExists reads an integer value of a named parameter from the module.
// It returns this value and the boolean true if the parameter exists and is an bool.
// In the other case it returns the boolean false and 0.
func (m *Module) GetBoolIfExists(path string) (bool, bool, error) {
	str, ok, err := m.GetStringIfExists(path)
	if err != nil {
		return false, false, err
	}

	if !ok {
		return false, false, nil
	}

	if len(str) == 0 || str == "0" {
		return false, true, nil
	}

	return true, true, nil
}

// GetString reads a string value of a named parameter from the module.
// It returns this value if the parameter exists and is a string.
// In the other case it return error unless default value is provided in
// the second argument.
func (m *Module) GetString(path string, d ...string) (string, error) {
	val, ok, err := m.GetStringIfExists(path)
	if err != nil {
		return d[0], err
	}

	if ok {
		return val, nil
	} else if len(d) > 0 {
		return d[0], nil
	} else {
		return "", fmt.Errorf("%s:%s key not exists and default not found", m.name, path)
	}
}

// GetInt reads an integer value of a named parameter from the module.
// It returns this value if the parameter exists and is an integer.
// In the other case it return error unless default value is provided in
// the second argument.
func (m *Module) GetInt(path string, d ...int64) (int64, error) {
	val, ok, err := m.GetIntIfExists(path)
	if err != nil {
		return d[0], err
	}

	if ok {
		return val, nil
	} else if len(d) > 0 {
		return d[0], nil
	} else {
		return 0, fmt.Errorf("%s:%s key not exists and default not found", m.name, path)
	}
}

// GetDuration reads an string value of a named parameter from the module and parse it to time.Duration.
// It returns this value if the parameter exists and is an time.Duration.
// In the other case it return error unless default value is provided in
// the second argument.
func (m *Module) GetDuration(path string, d ...time.Duration) (time.Duration, error) {
	val, ok, err := m.GetDurationIfExists(path)
	if err != nil {
		return d[0], err
	}

	if ok {
		return val, nil
	} else if len(d) > 0 {
		return d[0], nil
	} else {
		return 0, fmt.Errorf("%s:%s key not exists and default not found", m.name, path)
	}
}

// GetBool reads an bool value of a named parameter from the module.
// It returns this value if the parameter exists and is a bool.
// In the other case it return error unless default value is provided in
// the second argument.
func (m *Module) GetBool(path string, d ...bool) (bool, error) {
	val, ok, err := m.GetBoolIfExists(path)
	if err != nil {
		return d[0], err
	}

	if ok {
		return val, nil
	} else if len(d) > 0 {
		return d[0], nil
	} else {
		return false, fmt.Errorf("%s:%s key not exists and default not found", m.name, path)
	}
}

// GetStrings reads a []string value of a named parameter from the module.
// It returns this value if the parameter exists and is a comma-separated
// string or JSON array.
// In the other case it returns a default value provided in the second
// argument.
func (m *Module) GetStrings(path string, defaultValue []string) ([]string, error) {
	var value []string

	rv := reflect.ValueOf(&value).Elem()
	if m.getCache(path, rv) {
		return value, nil
	}

	format, data, err := m.get(path)
	if err != nil {
		return defaultValue, err
	}

	switch format {
	case 0:
		return defaultValue, nil
	case 's':
		untrimmed := strings.Split(string(data), ",")
		value = make([]string, 0, len(untrimmed))
		for _, item := range untrimmed {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				value = append(value, trimmed)
			}
		}

		m.setCache(path, rv)

		return value, nil
	case 'j':
		err := json.Unmarshal(data, &value)
		if err != nil {
			return nil, fmt.Errorf("%s:%s: failed to unmarshal JSON: %w", m.name, path, err)
		}

		m.setCache(path, rv)

		return value, nil
	default:
		return nil, fmt.Errorf("%s:%s: unexpected format", m.name, path)
	}
}

// GetStruct reads a structured value of a named parameter from the module.
// It stores this value in the value pointed by the value argument
// and returns true if the parameter exists and was unmarshaled successfully.
// In the case of error or if the parameter is not exists, the method doesn't
// touch the value argument, so you can safely pass a default value as the value
// argument and completely ignore return values of this method.
// A value is unmarshaled from JSON using json.Unmarshal and is cached internally
// until the configuration is updated, so be careful to not modify values returned by
// a reference.
// Experimental: this method can be modified or removed without any notice.
func (m *Module) GetStruct(path string, value interface{}) (bool, error) {
	var errMsg string

	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr {
		if rv.IsValid() {
			errMsg = fmt.Sprintf("%s: GetStruct(%q, non-pointer %s): invalid argument", m.name, path, rv.Type())
		} else {
			errMsg = fmt.Sprintf("%s: GetStruct(%q, nil): invalid argument", m.name, path)
		}
	} else if rv.IsNil() {
		errMsg = fmt.Sprintf("%s: GetStruct(%q, nil %s): invalid argument", m.name, path, rv.Type())
	}

	if errMsg != "" {
		return false, fmt.Errorf("%s: %w", errMsg, &json.InvalidUnmarshalError{Type: reflect.TypeOf(value)})
	}

	rv = rv.Elem()

	if m.getCache(path, rv) {
		return true, nil
	}

	format, data, err := m.get(path)
	if err != nil {
		return false, err
	}

	switch format {
	case 0:
		return false, nil
	case 'j':
		val := reflect.New(rv.Type())

		err := json.Unmarshal(data, val.Interface())
		if err != nil {
			return false, fmt.Errorf("%s:%s: failed to unmarshal JSON: %w", m.name, path, err)
		}

		rv.Set(val.Elem())
		m.setCache(path, rv)

		return true, nil
	default:
		return false, fmt.Errorf("%s:%s: %w", m.name, path, ErrFormatIsNotJSON)
	}
}

func (m *Module) getCache(path string, rv reflect.Value) bool {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	for _, cv := range m.cache[path] {
		rcv := reflect.ValueOf(cv)
		if rcv.Type() == rv.Type() {
			rv.Set(rcv)
			return true
		}
	}

	return false
}

func (m *Module) setCache(path string, rv reflect.Value) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	values := m.cache[path]
	for i := range values {
		if reflect.TypeOf(values[i]) == rv.Type() {
			values[i] = rv.Interface()
			return
		}
	}

	m.cache[path] = append(values, rv.Interface())
}
