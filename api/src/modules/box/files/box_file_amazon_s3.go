package files

import (
	"context"
	"io"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rs/zerolog/log"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type BoxFileAmazonS3 struct {
	bucket string

	session    *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

// NewBoxFileAmazonS3 init an S3 session
func NewBoxFileAmazonS3(region, bucket string) *BoxFileAmazonS3 {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		log.Fatal().Msg("could not initiate AWS S3 avatar bucket connection")
	}

	// create all required actors to interact with s3
	s3Cli := s3.New(sess)
	s := &BoxFileAmazonS3{
		session:    s3Cli,
		uploader:   s3manager.NewUploaderWithClient(s3Cli),
		downloader: s3manager.NewDownloaderWithClient(s3Cli),
		bucket:     bucket,
	}
	return s
}

// getKey by concatenating some info
func (s *BoxFileAmazonS3) getKey(boxID, fileID string) string {
	return filepath.Join(boxID, fileID)
}

// Upload data to amazon s3 at {bucket}/{boxID}/{fileID}
func (s *BoxFileAmazonS3) Upload(ctx context.Context, boxID, fileID string, data io.Reader) error {
	key := s.getKey(boxID, fileID)
	_, err := s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   data,
	})
	if err != nil {
		return merror.Internal().
			Describef("unable to upload %q to %q, %v", key, s.bucket, err)
	}
	return nil
}

// Download data from amazon S3 at {bucket}/{boxID}/{fileID}
func (s *BoxFileAmazonS3) Download(ctx context.Context, boxID, fileID string) ([]byte, error) {
	key := s.getKey(boxID, fileID)
	data := aws.NewWriteAtBuffer([]byte{})
	getObj := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	if _, err := s.downloader.DownloadWithContext(ctx, data, getObj); err != nil {
		return nil, merror.Internal().
			Describef("unable to download object %s from bucket %q, %v", key, s.bucket, err)
	}
	return data.Bytes(), nil
}

// DeleteAll data from s3 at {bucket}/{boxID}
func (s *BoxFileAmazonS3) DeleteAll(ctx context.Context, boxID string) error {
	if boxID == "" {
		return merror.Internal().Describe("box id cannot be empty to remove s3 files")
	}

	// setup BatchDeleteIterator to iterate through a list of objects.
	iter := s3manager.NewDeleteListIterator(s.session, &s3.ListObjectsInput{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(boxID + "/"),
	})

	// traverse iterator deleting each object
	if err := s3manager.NewBatchDeleteWithClient(s.session).Delete(aws.BackgroundContext(), iter); err != nil {
		return merror.Transform(err).Describef("unable to delete objects %q from %q", boxID, s.bucket)
	}
	return nil
}

// Delete data from s3 at {bucket}/{boxID}/{fileID}
func (s *BoxFileAmazonS3) Delete(ctx context.Context, boxID, fileID string) error {
	key := s.getKey(boxID, fileID)
	delObj := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	if _, err := s.session.DeleteObjectWithContext(ctx, delObj); err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NoSuchKey" {
			return merror.NotFound().Describe(err.Error())
		}
		return merror.Transform(err).
			Describef("unable to delete object %q from %q, %v", key, s.bucket, err)
	}
	return nil
}
