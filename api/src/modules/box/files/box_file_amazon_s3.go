package files

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rs/zerolog/log"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type FileAmazonS3 struct {
	bucket string

	session    *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

// NewFileAmazonS3 init an S3 session
func NewFileAmazonS3(region, bucket string) *FileAmazonS3 {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		log.Fatal().Msg("could not initiate AWS S3 avatar bucket connection")
	}

	// create all required actors to interact with s3
	s3Cli := s3.New(sess)
	s := &FileAmazonS3{
		session:    s3Cli,
		uploader:   s3manager.NewUploaderWithClient(s3Cli),
		downloader: s3manager.NewDownloaderWithClient(s3Cli),
		bucket:     bucket,
	}
	return s
}

// Upload data to amazon s3 at {bucket}/{fileID}
func (s *FileAmazonS3) Upload(ctx context.Context, fileID string, data io.Reader) error {
	_, err := s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileID),
		Body:   data,
	})
	if err != nil {
		return merror.Internal().
			Describef("unable to upload %q to %q, %v", fileID, s.bucket, err)
	}
	return nil
}

// Download data from amazon S3 at {bucket}/{fileID}
func (s *FileAmazonS3) Download(ctx context.Context, fileID string) ([]byte, error) {
	data := aws.NewWriteAtBuffer([]byte{})
	getObj := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileID),
	}
	if _, err := s.downloader.DownloadWithContext(ctx, data, getObj); err != nil {
		return nil, merror.Internal().
			Describef("unable to download object %s from bucket %q, %v", fileID, s.bucket, err)
	}
	return data.Bytes(), nil
}

// Delete data from s3 at {bucket}/{fileID}
func (s *FileAmazonS3) Delete(ctx context.Context, fileID string) error {
	delObj := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileID),
	}
	if _, err := s.session.DeleteObjectWithContext(ctx, delObj); err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NoSuchKey" {
			return merror.NotFound().Describe(err.Error())
		}
		return merror.Transform(err).
			Describef("unable to delete object %q from %q, %v", fileID, s.bucket, err)
	}
	return nil
}
