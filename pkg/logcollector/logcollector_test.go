package logcollector

import (
	"context"
	"fmt"
	"github.com/harvester/harvester-installer/pkg/config"
	"github.com/harvester/harvester-installer/pkg/util"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestNewLogCollector(t *testing.T) {
	hc, err := config.LoadHarvesterConfig(util.LoadFixture(t, "harvester-config.yaml"))
	assert.NoError(t, err)

	l := NewLogCollector(context.TODO(), &hc.LogCollector.UploadConfig)
	assert.NotNil(t, l)
}

func TestGenerateSupportBundleWithFailure(t *testing.T) {
	hc, err := config.LoadHarvesterConfig(util.LoadFixture(t, "harvester-config.yaml"))
	assert.NoError(t, err)

	l := NewLogCollector(context.TODO(), &hc.LogCollector.UploadConfig)
	assert.NotNil(t, l)

	out, err := l.GenerateSupportBundle()
	assert.Error(t, err, "expected error")
	assert.NotNil(t, out)
}

func TestObjectUpload(t *testing.T) {
	hc, err := config.LoadHarvesterConfig(util.LoadFixture(t, "harvester-config.yaml"))
	assert.NoError(t, err)

	l := NewLogCollector(context.TODO(), &hc.LogCollector.UploadConfig)
	assert.NotNil(t, l)

	// Create a temp file for upload
	content := util.LoadFixture(t, "kernel.log")
	fileName := fmt.Sprintf("%s/scc_%s.txz", DefaultOutputPath, l.bundleSuffix)
	err = ioutil.WriteFile(fileName, content, 0755)
	assert.NoError(t, err)

	// setup a temp bucket
	tmpClient, err := minioClient(l.ObjectStoreConfig)
	assert.NoError(t, err)

	err = tmpClient.MakeBucket(context.TODO(), l.ObjectStoreConfig.BucketName, minio.MakeBucketOptions{
		Region: "us-east-1",
	})

	assert.NoError(t, err)

	// test upload
	err = l.objectUpload()
	assert.NoError(t, err)

	// cleanup bucket
	err = tmpClient.RemoveBucketWithOptions(context.TODO(), l.ObjectStoreConfig.BucketName, minio.RemoveBucketOptions{
		ForceDelete: true,
	})
	assert.NoError(t, err)
}
