package noop

import "context"

//encore:api private method=GET path=/ping
func Ping(ctx context.Context) error {
	return nil
}
