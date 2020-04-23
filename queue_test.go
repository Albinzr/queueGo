package queue

import "testing"

var queueConfig = &Config{
	StoragePath: "/location",
	FileSize:    1,
	NoOfRetries: 3,
}

func Test(t *testing.T) {
	queueConfig.Init()
}
