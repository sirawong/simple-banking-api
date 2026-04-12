package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"

	adapterdb "github.com/sirawong/simple-banking-api/internal/adapter/postgres"
	"github.com/sirawong/simple-banking-api/internal/config"
	pkgjwt "github.com/sirawong/simple-banking-api/pkg/jwt"
	"github.com/sirawong/simple-banking-api/pkg/logger"
	testdi "github.com/sirawong/simple-banking-api/test/integration/di"
	"github.com/sirawong/simple-banking-api/test/integration/testutil"
)

var (
	sharedDB    *testutil.TestDB
	sharedRedis *testutil.TestRedis
)

func TestMain(m *testing.M) {
	err := os.Setenv("ENV_FILE", "../../.env.test")
	if err != nil {
		panic("failed to get env file: " + err.Error())
	}

	ctx := context.Background()

	sharedDB, err = testutil.StartTestDB(ctx)
	if err != nil {
		panic("failed to connect to test DB: " + err.Error())
	}

	sharedRedis, err = testutil.StartTestRedis(ctx)
	if err != nil {
		panic("failed to connect to test Redis: " + err.Error())
	}

	code := m.Run()

	_ = sharedDB.Close()
	_ = sharedRedis.Close()

	os.Exit(code)
}

type BaseSuite struct {
	suite.Suite
	ctx        context.Context
	db         *adapterdb.DB
	router     *gin.Engine
	jwtManager pkgjwt.Manager
}

func (s *BaseSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.ctx = context.Background()
	s.db = sharedDB.DB

	cfg, err := config.ProvideConfig()
	s.Require().NoError(err)
	s.jwtManager = pkgjwt.ProvideManager(cfg)

	log := logger.ProvideGlobalLogger()
	s.router, err = testdi.InitializeRouter(s.db, sharedRedis.Client, log)
	s.Require().NoError(err)
}

func (s *BaseSuite) SetupTest() {
	s.Require().NoError(sharedDB.TruncateAll(s.ctx))
	s.Require().NoError(sharedRedis.FlushAll(s.ctx))
}

func (s *BaseSuite) TearDownTest() {
	s.Require().NoError(sharedDB.TruncateAll(s.ctx))
	s.Require().NoError(sharedRedis.FlushAll(s.ctx))
}
