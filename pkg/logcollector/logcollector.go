package logcollector

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/harvester/harvester-installer/pkg/config"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/rand"
	"os"
	"os/exec"
	"strings"
)

type LogCollector struct {
	ctx          context.Context
	bundleSuffix string
	*config.UploadConfig
}

type SupportBundleDetails struct {
	Output bytes.Buffer
	Name   string
}

var (
	DefaultArgs         = []string{"-R", DefaultOutputPath, "-B"}
	DefaultLogCollector = "/sbin/supportconfig"
	DefaultOutputPath   = "/tmp"
	DefaultNFSVersion   = "v4"
	DefaultMountPath    = "/var/log/supportconfigupload"
	MinorVersions       = []string{"4.2", "4.1", "4.0"}
)

func NewLogCollector(ctx context.Context, c *config.UploadConfig) *LogCollector {
	return &LogCollector{
		ctx:          ctx,
		bundleSuffix: fmt.Sprintf("supportbundle_%s", rand.String(5)),
		UploadConfig: c,
	}
}

func (l *LogCollector) GenerateSupportBundle() (SupportBundleDetails, error) {
	args := append(DefaultArgs, l.bundleSuffix)
	lCmd := exec.Command(DefaultLogCollector, args...)
	var out bytes.Buffer
	lCmd.Stdout = &out

	err := lCmd.Run()

	s := SupportBundleDetails{
		Output: out,
		Name:   fmt.Sprintf("%s/scc_%s.txz", DefaultOutputPath, l.bundleSuffix),
	}

	return s, err
}

func (l *LogCollector) UploadSupportBundle() error {
	if l.UploadConfig == nil {
		return nil // no actual action to be performed
	}

	var err, childError error
	if l.UploadConfig.NFSConfig != nil {
		childError = l.nfsUpload()
	}

	if childError != nil {
		err = fmt.Errorf("error performing NFS upload: %v \n", childError)
	}

	// check and try uploaded to Object store as well
	if l.UploadConfig.ObjectStoreConfig != nil {
		childError = l.objectUpload()
	}

	if childError != nil {
		err = fmt.Errorf("error performing Object upload: %v", childError)
	}

	return err
}

func (l *LogCollector) nfsUpload() error {
	_, err := os.Stat(DefaultMountPath)
	if errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(DefaultMountPath, 0755)
		if err != nil {
			return err
		}
	}

	var errs []error
	if !isMounted(DefaultMountPath) {
		for _, version := range MinorVersions {
			_, err = exec.Command("mount", "-t", "nfs4", "-o", fmt.Sprintf("nfsvers=%v", version), "-o", "actimeo=1", l.NFSConfig.Endpoint, DefaultMountPath).Output()
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) != 0 {
		err = fmt.Errorf("error during nfs mount: ")
		for _, v := range errs {
			err = fmt.Errorf("%v %v", err, v)
		}
		return err
	}

	// copy file
	input := fmt.Sprintf("%s/scc_%s.txz", DefaultOutputPath, l.bundleSuffix)
	output := fmt.Sprintf("%s/scc_%s.txz", DefaultMountPath, l.bundleSuffix)

	inputBytes, err := ioutil.ReadFile(input)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(output, inputBytes, 0700)
	return err
}

func (l *LogCollector) objectUpload() error {
	client, err := minioClient(l.ObjectStoreConfig)

	if err != nil {
		return err
	}
	// upload the tar ball
	objectName := fmt.Sprintf("scc_%s.txz", l.bundleSuffix)
	fileName := fmt.Sprintf("%s/scc_%s.txz", DefaultOutputPath, l.bundleSuffix)

	_, err = client.FPutObject(l.ctx, l.ObjectStoreConfig.BucketName, objectName, fileName, minio.PutObjectOptions{ContentType: "application/zip"})

	return err
}

func minioClient(objConf *config.ObjectStore) (*minio.Client, error) {
	client, err := minio.New(objConf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(objConf.AccessKeyID, objConf.SecretAccessKey, objConf.SessionToken),
		Secure: objConf.InsecureTLS,
	})

	return client, err
}

func isMounted(path string) bool {
	out, err := exec.Command("mount").Output()
	if err != nil {
		return false
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, " "+path+" ") {
			return true
		}
	}

	return false
}
