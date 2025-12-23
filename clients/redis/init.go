package redis

import (
	goredis "github.com/redis/go-redis/v9"

	"github.com/codoworks/codo-framework/core/errors"
)

func init() {
	// Get the global error mapper
	mapper := errors.GetMapper()

	// Register Redis-specific error types for automatic mapping

	// redis.Nil - Key does not exist (404)
	// This is the go-redis sentinel for when a key lookup returns nothing
	mapper.RegisterSentinel(goredis.Nil, errors.MappingSpec{
		Code:       errors.CodeNotFound,
		HTTPStatus: 404,
		LogLevel:   errors.LogLevelWarn,
		Message:    "Key not found",
	})
}
