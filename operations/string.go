package operations

import (
	"fmt"
	"my-redis-go/datastore"
	"my-redis-go/resp"
)

const (
	getCommand = "GET"
	setCommand = "SET"
	command    = "COMMAND"
)

var kvStore = datastore.NewKeyValueStore()

func executeGetCommand(array *resp.Array) (resp.IDataType, resp.RedisError) {
	numberOfItems := array.GetNumberOfItems()

	if numberOfItems == 1 {
		return nil, resp.NewDefaultRedisError("wrong number of arguments for (get) command")
	} else if numberOfItems > 2 {
		fmt.Printf("WARN: GET command acccepts only one argument. But received %d. Other arguments will be ignored\n", numberOfItems-1)
	}

	key := array.GetItemAtIndex(1)

	value, ok := kvStore.Load(key.ToString())
	if !ok {
		return resp.EmptyBulkString, resp.EmptyRedisError
	}
	bs, e := resp.NewBulkString(value)
	if e != nil {
		return nil, resp.NewDefaultRedisError(e.Error())
	}
	return bs, resp.EmptyRedisError
}

func executeSetCommand(array *resp.Array) (resp.IDataType, resp.RedisError) {
	numberOfItems := array.GetNumberOfItems()

	if numberOfItems <= 2 {
		return nil, resp.NewDefaultRedisError("wrong number of arguments for (get) command")
	} else if numberOfItems > 3 {
		fmt.Printf("WARN: SET command acccepts only two argument. But received %d. Other arguments will be ignored\n", numberOfItems-1)
	}

	key := array.GetItemAtIndex(1)
	value := array.GetItemAtIndex(2)
	kvStore.Store(key.ToString(), value.ToString())

	bs, e := resp.NewBulkString("OK")
	if e != nil {
		return nil, resp.NewDefaultRedisError(e.Error())
	}
	return bs, resp.EmptyRedisError
}

func ExecuteCommand(array resp.Array) (resp.IDataType, resp.RedisError) {
	if array.GetNumberOfItems() == 0 {
		return nil, resp.NewDefaultRedisError("No command found")
	}
	first := array.GetItemAtIndex(0)
	switch first.ToString() {
	case getCommand:
		return executeGetCommand(&array)
	case setCommand:
		return executeSetCommand(&array)
	case command:
		fmt.Println("COM")
		bs, _ := resp.NewBulkString("OK")
		return bs, resp.EmptyRedisError
	default:
		break
	}
	return nil, resp.NewDefaultRedisError(fmt.Sprintf("Unknown or disabled command '%s'", first.ToString()))
}
