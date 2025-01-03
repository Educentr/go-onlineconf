package onlineconf

import (
	"context"
	"fmt"
	"time"
)

// GetStringIfExists reads a string value of a named parameter from the module "TREE".
// It returns the boolean true if the parameter exists and is a string.
// In the other case it returns the boolean false and an empty string.
func GetStringIfExists(ctx context.Context, path string) (string, bool, error) {
	return FromContext(ctx).GetStringIfExists(path)
}

// GetIntIfExists reads an integer value of a named parameter from the module "TREE".
// It returns this value and the boolean true if the parameter exists and is an integer.
// In the other case it returns the boolean false and 0.
func GetIntIfExists(ctx context.Context, path string) (int64, bool, error) {
	return FromContext(ctx).GetIntIfExists(path)
}

// GetDurationIfExists reads an string value of a named parameter from the module "TREE" and parse it to time.Duration.
// It returns this value and the boolean true if the parameter exists and is an integer.
// In the other case it returns the boolean false and 0.
func GetDurationIfExists(ctx context.Context, path string) (time.Duration, bool, error) {
	return FromContext(ctx).GetDurationIfExists(path)
}

// GetBoolIfExists reads an bool value of a named parameter from the module "TREE".
// It returns this value and the boolean true if the parameter exists and is a bool.
// In the other case it returns the boolean false and 0.
func GetBoolIfExists(ctx context.Context, path string) (bool, bool, error) {
	return FromContext(ctx).GetBoolIfExists(path)
}

// GetString reads a string value of a named parameter from the module "TREE".
// It returns this value if the parameter exists and is a string.
// In the other case it panics unless default value is provided in
// the second argument.
func GetString(ctx context.Context, path string, d ...string) (string, error) {
	return FromContext(ctx).GetString(path, d...)
}

// GetInt reads an integer value of a named parameter from the module "TREE".
// It returns this value if the parameter exists and is an integer.
// In the other case it panics unless default value is provided in
// the second argument.
func GetInt(ctx context.Context, path string, d ...int64) (int64, error) {
	return FromContext(ctx).GetInt(path, d...)
}

// GetDuration reads an string value and parse to time.Duration of a named parameter from the module "TREE".
// It returns this value if the parameter exists and is an integer.
// In the other case it panics unless default value is provided in
// the second argument.
func GetDuration(ctx context.Context, path string, d ...time.Duration) (time.Duration, error) {
	return FromContext(ctx).GetDuration(path, d...)
}

// GetBool reads an bool value of a named parameter from the module "TREE".
// It returns this value if the parameter exists and is a bool.
// In the other case it panics unless default value is provided in
// the second argument.
func GetBool(ctx context.Context, path string, d ...bool) (bool, error) {
	return FromContext(ctx).GetBool(path, d...)
}

// GetStrings reads a []string value of a named parameter from the module "TREE".
// It returns this value if the parameter exists and is a comma-separated string
// or JSON array.
// In the other case it returns a default value provided in the second argument.
func GetStrings(ctx context.Context, path string, defaultValue []string) ([]string, error) {
	return FromContext(ctx).GetStrings(path, defaultValue)
}

// GetStruct reads a structured value of a named parameter from the module "TREE".
// It stores this value in the value pointed by the value argument
// and returns true if the parameter exists and was unmarshaled successfully.
// In the case of error or if the parameter is not exists, the function doesn't
// touch the value argument, so you can safely pass a default value as the value
// argument and completely ignore return values of this function.
// A value is unmarshaled from JSON using json.Unmarshal and is cached internally
// until the configuration is updated, so be careful to not modify values returned by
// a reference.
// Experimental: this function can be modified or removed without any notice.
func GetStruct(ctx context.Context, path string, value interface{}) (bool, error) {
	return FromContext(ctx).GetStruct(path, value)
}

// GetStringIfExists reads a string value of a named parameter from the module "TREE".
// It returns the boolean true if the parameter exists and is a string.
// In the other case it returns the boolean false and an empty string.
func (oi *OnlineconfInstance) GetStringIfExists(path string) (string, bool, error) {
	m, err := oi.GetOrAddModule(DefaultModule)

	if err != nil {
		return "", false, fmt.Errorf("can't get TREE module: %w", err)
	}

	return m.GetStringIfExists(path)
}

// GetIntIfExists reads an integer value of a named parameter from the module "TREE".
// It returns this value and the boolean true if the parameter exists and is an integer.
// In the other case it returns the boolean false and 0.
func (oi *OnlineconfInstance) GetIntIfExists(path string) (int64, bool, error) {
	m, err := oi.GetOrAddModule(DefaultModule)
	if err != nil {
		return 0, false, fmt.Errorf("can't get TREE module: %w", err)
	}

	return m.GetIntIfExists(path)
}

// GetIntIfExists reads an integer value of a named parameter from the module "TREE".
// It returns this value and the boolean true if the parameter exists and is an integer.
// In the other case it returns the boolean false and 0.
func (oi *OnlineconfInstance) GetDurationIfExists(path string) (time.Duration, bool, error) {
	m, err := oi.GetOrAddModule(DefaultModule)
	if err != nil {
		return 0, false, fmt.Errorf("can't get TREE module: %w", err)
	}

	return m.GetDurationIfExists(path)
}

// GetBoolIfExists reads an bool value of a named parameter from the module "TREE".
// It returns this value and the boolean true if the parameter exists and is a bool.
// In the other case it returns the boolean false and 0.
func (oi *OnlineconfInstance) GetBoolIfExists(path string) (bool, bool, error) {
	m, err := oi.GetOrAddModule(DefaultModule)
	if err != nil {
		return false, false, fmt.Errorf("can't get TREE module: %w", err)
	}

	return m.GetBoolIfExists(path)
}

// GetString reads a string value of a named parameter from the module "TREE".
// It returns this value if the parameter exists and is a string.
// In the other case it panics unless default value is provided in
// the second argument.
func (oi *OnlineconfInstance) GetString(path string, d ...string) (string, error) {
	m, err := oi.GetOrAddModule(DefaultModule)
	if err != nil {
		return "", fmt.Errorf("can't get TREE module: %w", err)
	}

	ret, err := m.GetString(path, d...)

	return ret, err
}

// GetInt reads an integer value of a named parameter from the module "TREE".
// It returns this value if the parameter exists and is an integer.
// In the other case it panics unless default value is provided in
// the second argument.
func (oi *OnlineconfInstance) GetInt(path string, d ...int64) (int64, error) {
	m, err := oi.GetOrAddModule(DefaultModule)
	if err != nil {
		return 0, fmt.Errorf("can't get TREE module: %w", err)
	}

	return m.GetInt(path, d...)
}

// GetDuration reads an string value of a named parameter from the module "TREE" and parse it to time.Duration.
// It returns this value if the parameter exists and is an integer.
// In the other case it panics unless default value is provided in
// the second argument.
func (oi *OnlineconfInstance) GetDuration(path string, d ...time.Duration) (time.Duration, error) {
	m, err := oi.GetOrAddModule(DefaultModule)
	if err != nil {
		return 0, fmt.Errorf("can't get TREE module: %w", err)
	}

	return m.GetDuration(path, d...)
}

// GetBool reads an bool value of a named parameter from the module "TREE".
// It returns this value if the parameter exists and is a bool.
// In the other case it panics unless default value is provided in
// the second argument.
func (oi *OnlineconfInstance) GetBool(path string, d ...bool) (bool, error) {
	m, err := oi.GetOrAddModule(DefaultModule)
	if err != nil {
		return false, fmt.Errorf("can't get TREE module: %w", err)
	}

	return m.GetBool(path, d...)
}

// GetStrings reads a []string value of a named parameter from the module "TREE".
// It returns this value if the parameter exists and is a comma-separated string
// or JSON array.
// In the other case it returns a default value provided in the second argument.
func (oi *OnlineconfInstance) GetStrings(path string, defaultValue []string) ([]string, error) {
	m, err := oi.GetOrAddModule(DefaultModule)
	if err != nil {
		return nil, fmt.Errorf("can't get TREE module: %w", err)
	}

	return m.GetStrings(path, defaultValue)
}

// GetStruct reads a structured value of a named parameter from the module "TREE".
// It stores this value in the value pointed by the value argument
// and returns true if the parameter exists and was unmarshaled successfully.
// In the case of error or if the parameter is not exists, the function doesn't
// touch the value argument, so you can safely pass a default value as the value
// argument and completely ignore return values of this function.
// A value is unmarshaled from JSON using json.Unmarshal and is cached internally
// until the configuration is updated, so be careful to not modify values returned by
// a reference.
// Experimental: this function can be modified or removed without any notice.
func (oi *OnlineconfInstance) GetStruct(path string, value interface{}) (bool, error) {
	m, err := oi.GetOrAddModule(DefaultModule)
	if err != nil {
		return false, fmt.Errorf("can't get TREE module: %w", err)
	}

	return m.GetStruct(path, value)
}
