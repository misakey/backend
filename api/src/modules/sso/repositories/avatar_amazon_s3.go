package repositories

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type AvatarAmazonS3 struct {
	session  *s3.S3
	uploader *s3manager.Uploader

	bucket       string
	avatarDomain string
	folder       string
}

// NewAvatarAmazonS3 init an S3 session
func NewAvatarAmazonS3(region, bucket string, avatarDomain string) (*AvatarAmazonS3, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		return nil, fmt.Errorf("could not create aws session (%v)", err)
	}

	// create all required actors to interact with s3
	s := &AvatarAmazonS3{
		session:  s3.New(sess),
		uploader: s3manager.NewUploader(sess),

		bucket:       bucket,
		avatarDomain: avatarDomain,
		folder:       "profile/picture/",
	}
	return s, nil
}

// getKey by concatenating some info
func (s *AvatarAmazonS3) getKey(avatar *domain.AvatarFile) string {
	return filepath.Join(s.folder, avatar.Filename)
}

func (s *AvatarAmazonS3) Upload(ctx context.Context, avatar *domain.AvatarFile) (string, error) {
	uo, err := s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.getKey(avatar)),
		Body:   avatar.Data,
	})
	if err != nil {
		return "", merror.Internal().Describef("unable to upload %q to %q, %v", s.getKey(avatar), s.bucket, err)
	}

	avatarURI, _ := url.Parse(uo.Location)
	avatarURI.Host = s.avatarDomain
	avatarURI.Path = strings.TrimPrefix(avatarURI.Path, "/"+s.bucket)

	return avatarURI.String(), nil
}

func (s *AvatarAmazonS3) Delete(ctx context.Context, avatar *domain.AvatarFile) error {
	delObj := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.getKey(avatar)),
	}
	if _, err := s.session.DeleteObjectWithContext(ctx, delObj); err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NoSuchKey" {
			return merror.NotFound().Describe(err.Error())
		}
		return merror.Transform(err).Describef("unable to delete object %q from %q, %v", s.getKey(avatar), s.bucket, err)
	}
	return nil
}
