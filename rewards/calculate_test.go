package rewards

import (
	"context"
	"fmt"
	"fsp-rewards-calculator/common/params"
	ty2 "fsp-rewards-calculator/common/ty"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	network               = "coston"
	testEpochId           = 4253
	testEpochExpectedRoot = "0x5c3b8dced8aca4a35d0a5608b45f30334b0b4af9a51b8ce01a9cf6ff446a0bf5"

	dbPort = "3306/tcp"
	dbUser = "root"
	dbPass = "password"
)

// TestCalculateResults tests calculating rewards for a specific reward epoch against a Coston c-chain-indexer database snapshot.
func TestCalculateResults(t *testing.T) {
	ctx := context.Background()
	gormDB, mysqlC := setUpMySqlDb(t, ctx)
	defer mysqlC.Terminate(ctx) //nolint:errcheck

	params.InitNetwork(network)
	testEpoch := ty2.RewardEpochId(testEpochId)
	result := CalculateResults(gormDB, testEpoch)

	if result.MerkleRoot != testEpochExpectedRoot {
		t.Errorf("Wrong merkle root, expected %s, got %s", testEpochExpectedRoot, result.MerkleRoot)
	}
}

// setUpMySqlDb sets up a MySQL database Docker container.
func setUpMySqlDb(t *testing.T, ctx context.Context) (*gorm.DB, testcontainers.Container) {
	req := testcontainers.ContainerRequest{
		Image:        "mysql",
		ExposedPorts: []string{dbPort},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": dbPass,
			"MYSQL_DATABASE":      "testdb",
		},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      fmt.Sprintf(`../testdata/%s-%d.sql`, network, testEpochId),
				ContainerFilePath: "/docker-entrypoint-initdb.d/init.sql",
				FileMode:          0644,
			},
		},
		WaitingFor: wait.ForLog("MySQL Community Server - GPL"),
	}
	mysqlC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Fatalf("Failed to start MySQL container: %v", err)
	}

	host, err := mysqlC.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}
	port, err := mysqlC.MappedPort(ctx, nat.Port(dbPort))
	if err != nil {
		t.Fatalf("Failed to get mapped port: %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/testdb?parseTime=true", dbUser, dbPass, host, port.Port())

	// Wait for MySQL to be ready
	var gormDB *gorm.DB
	for i := 0; i < 10; i++ {
		gormDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		t.Fatalf("Failed to connect to MySQL: %v", err)
	}
	return gormDB, mysqlC
}
