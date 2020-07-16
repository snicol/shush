package cache

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/keybase/go-keychain"
)

type Keychain struct {
	SecClass    keychain.SecClass
	Service     string
	AccessGroup string
}

const labelFormat = "1:%s:%s:%d"

func NewKeychain(secClass keychain.SecClass, service, accessGroup string) *Keychain {
	return &Keychain{
		SecClass:    secClass,
		Service:     service,
		AccessGroup: accessGroup,
	}
}

func (k *Keychain) Get(key string) (string, int, error) {
	query := keychain.NewItem()
	query.SetSecClass(k.SecClass)
	query.SetService(k.Service)
	query.SetAccount(key)
	query.SetAccessGroup(k.AccessGroup)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnAttributes(true)
	query.SetReturnData(true)

	results, err := keychain.QueryItem(query)
	if err != nil {
		return "", 0, err
	}

	if len(results) == 0 {
		return "", 0, ErrNotFound
	}

	if len(results) > 1 {
		return "", 0, errors.New("expected one result from MatchLimitOne query")
	}

	parts := strings.Split(results[0].Label, ":")
	if len(parts) != 4 {
		return "", 0, errors.New("invalid number of parts in keychain label")
	}

	if parts[0] != "1" {
		return "", 0, errors.New("unknown label version")
	}

	v, err := strconv.Atoi(parts[3])
	if err != nil {
		return "", 0, err
	}

	return string(results[0].Data), v, nil
}

func (k *Keychain) Set(version int, key, v string) error {
	item := keychain.NewItem()
	item.SetSecClass(k.SecClass)
	item.SetService(k.Service)
	item.SetAccount(key)
	item.SetLabel(fmt.Sprintf(labelFormat, k.Service, key, version))
	item.SetAccessGroup(k.AccessGroup)
	item.SetData([]byte(v))
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleWhenUnlocked)

	err := keychain.AddItem(item)
	if err != keychain.ErrorDuplicateItem {
		return err
	}

	item = keychain.NewItem()
	item.SetSecClass(k.SecClass)
	item.SetService(k.Service)
	item.SetAccount(key)
	if err := keychain.DeleteItem(item); err != nil {
		return err
	}

	return k.Set(version, key, v)
}
