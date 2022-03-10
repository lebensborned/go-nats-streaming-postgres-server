package config

import "os"

// SetupEnv is just mock config, should be in .env file
func SetupEnv() {
	os.Setenv("DB_USERNAME", "anton")
	os.Setenv("DB_PASSWORD", "root")
	os.Setenv("DB_NAME", "orders_service")

	os.Setenv("SERVER_PORT", ":8080")

	os.Setenv("NATS_URL", "nats://127.0.0.1:4222")
	os.Setenv("NATS_CLUSTER_ID", "test-cluster")
	os.Setenv("NATS_CLIENT_ID", "wb-test")
	os.Setenv("NATS_SUBJECT", "test")
	os.Setenv("NATS_DURABLE_NAME", "test")
}
