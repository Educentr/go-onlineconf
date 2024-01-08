package onlineconf

import (
	"context"
	"fmt"
)

// GetStringIfExists reads a string value of a named parameter from the module "TREE".
// It returns the boolean true if the parameter exists and is a string.
// In the other case it returns the boolean false and an empty string.
func GetStringIfExists(ctx context.Context, path string) (string, bool, error) {
	m, err := FromContext(ctx).GetOrAdd(DefaultModule)

	if err != nil {
		return "", false, fmt.Errorf("can't get TREE module: %w", err)
	}

	return m.GetStringIfExists(path)
}

// GetIntIfExists reads an integer value of a named parameter from the module "TREE".
// It returns this value and the boolean true if the parameter exists and is an integer.
// In the other case it returns the boolean false and 0.
func GetIntIfExists(ctx context.Context, path string) (int, bool, error) {
	m, err := FromContext(ctx).GetOrAdd(DefaultModule)
	if err != nil {
		return 0, false, fmt.Errorf("can't get TREE module: %w", err)
	}

	return m.GetIntIfExists(path)
}

// GetBoolIfExists reads an bool value of a named parameter from the module "TREE".
// It returns this value and the boolean true if the parameter exists and is a bool.
// In the other case it returns the boolean false and 0.
func GetBoolIfExists(ctx context.Context, path string) (bool, bool, error) {
	m, err := FromContext(ctx).GetOrAdd(DefaultModule)
	if err != nil {
		return false, false, fmt.Errorf("can't get TREE module: %w", err)
	}

	return m.GetBoolIfExists(path)
}

// GetString reads a string value of a named parameter from the module "TREE".
// It returns this value if the parameter exists and is a string.
// In the other case it panics unless default value is provided in
// the second argument.
func GetString(ctx context.Context, path string, d ...string) (string, error) {
	m, err := FromContext(ctx).GetOrAdd(DefaultModule)
	if err != nil {
		return "", fmt.Errorf("can't get TREE module: %w", err)
	}

	return m.GetString(path, d...)
}

// GetInt reads an integer value of a named parameter from the module "TREE".
// It returns this value if the parameter exists and is an integer.
// In the other case it panics unless default value is provided in
// the second argument.
func GetInt(ctx context.Context, path string, d ...int) (int, error) {
	m, err := FromContext(ctx).GetOrAdd(DefaultModule)
	if err != nil {
		return 0, fmt.Errorf("can't get TREE module: %w", err)
	}

	return m.GetInt(path, d...)
}

// GetBool reads an bool value of a named parameter from the module "TREE".
// It returns this value if the parameter exists and is a bool.
// In the other case it panics unless default value is provided in
// the second argument.
func GetBool(ctx context.Context, path string, d ...bool) (bool, error) {
	m, err := FromContext(ctx).GetOrAdd(DefaultModule)
	if err != nil {
		return false, fmt.Errorf("can't get TREE module: %w", err)
	}

	return m.GetBool(path, d...)
}

// GetStrings reads a []string value of a named parameter from the module "TREE".
// It returns this value if the parameter exists and is a comma-separated string
// or JSON array.
// In the other case it returns a default value provided in the second argument.
func GetStrings(ctx context.Context, path string, defaultValue []string) ([]string, error) {
	m, err := FromContext(ctx).GetOrAdd(DefaultModule)
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
func GetStruct(ctx context.Context, path string, value interface{}) (bool, error) {
	m, err := FromContext(ctx).GetOrAdd(DefaultModule)
	if err != nil {
		return false, fmt.Errorf("can't get TREE module: %w", err)
	}

	return m.GetStruct(path, value)
}
