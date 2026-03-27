// Package storage provides persistent key-value storage.
// On iOS it uses NSUserDefaults, on Android SharedPreferences.
package storage

import "encoding/json"

// Set stores a value under the given key.
// The value is JSON-marshaled before storage.
func Set(key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return platformSet(key, string(data))
}

// Get retrieves a value by key and unmarshals it into dest.
// Returns an error if the key doesn't exist or unmarshaling fails.
func Get(key string, dest any) error {
	data, err := platformGet(key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// Delete removes a value by key.
func Delete(key string) error {
	return platformDelete(key)
}
